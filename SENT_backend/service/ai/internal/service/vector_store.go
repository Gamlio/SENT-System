package service

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AutoCheckAndLoadPlaybooks() {
	if database.DocumentContentCollection == nil {
		fmt.Println("❌ [VECTOR] MongoDB connection not ready for playbook ingestion.")
		return
	}

	vectorCollection := database.DocumentContentCollection.Database().Collection("playbook_vectors")

	// 1. Kiểm tra và nạp Playbook
	pbCount, _ := vectorCollection.CountDocuments(context.TODO(), bson.M{"playbook_id": bson.M{"$ne": "ISO27001"}})
	if pbCount > 0 {
		fmt.Printf("✅ [VECTOR] Playbook vector database is operational. Found [%d] playbook vectors.\n", pbCount)
	} else {
		fmt.Println("⚠️ [VECTOR] Playbook vector storage is empty! Attempting to load playbook file...")
		targetPath := findFile([]string{"doc/playbook.md", "service/ai/doc/playbook.md", "../doc/playbook.md"})
		if targetPath != "" {
			if contentBytes, err := os.ReadFile(targetPath); err == nil {
				fmt.Printf("🚀 [VECTOR] Found playbook file at [%s]. Ingesting...\n", targetPath)
				_ = IngestPlaybookMarkdownToVectorDB(string(contentBytes))
			}
		}
	}

	// 2. Kiểm tra và nạp ISO27001
	isoCount, _ := vectorCollection.CountDocuments(context.TODO(), bson.M{"playbook_id": "ISO27001"})
	if isoCount > 0 {
		fmt.Printf("✅ [VECTOR] ISO27001 vector database is operational. Found [%d] ISO chunks.\n", isoCount)
	} else {
		fmt.Println("⚠️ [VECTOR] ISO27001 vector storage is empty! Attempting to load ISO file...")
		targetPath := findFile([]string{"doc/ISO27001_2022.md", "service/ai/doc/ISO27001_2022.md", "../doc/ISO27001_2022.md"})
		if targetPath != "" {
			if contentBytes, err := os.ReadFile(targetPath); err == nil {
				fmt.Printf("🚀 [VECTOR] Found ISO file at [%s]. Ingesting...\n", targetPath)
				_ = IngestISOMarkdownToVectorDB(string(contentBytes))
			}
		}
	}
}

func findFile(paths []string) string {
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func IngestISOMarkdownToVectorDB(mdContent string) error {
	if database.DocumentContentCollection == nil {
		return fmt.Errorf("chưa kết nối cơ sở dữ liệu MongoDB")
	}

	vectorCollection := database.DocumentContentCollection.Database().Collection("playbook_vectors")

	// Xóa dữ liệu ISO cũ nếu có
	_, _ = vectorCollection.DeleteMany(context.TODO(), bson.M{"playbook_id": "ISO27001"})

	successCount := 0
	// Thay vì cắt bằng \n\n rất dễ dính khối text khổng lồ, ta cắt nhỏ theo từng dòng \n
	lines := strings.Split(mdContent, "\n")

	var currentChunk strings.Builder
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Nếu thêm dòng này vào mà vượt quá 600 ký tự (mức an toàn cao cho Ollama), ta tiến hành sinh vector khối cũ trước
		if currentChunk.Len()+len(trimmed) > 600 && currentChunk.Len() > 0 {
			blockText := currentChunk.String()
			vector, err := GetTextEmbedding(blockText)
			if err != nil {
				fmt.Printf("❌ Lỗi sinh vector cho chunk ISO: %v\n", err)
				currentChunk.Reset() // Reset để bỏ qua block lỗi, tránh nghẽn luồng
				continue
			}

			chunkDoc := models.PlaybookVectorChunk{
				ID:         primitive.NewObjectID(),
				PlaybookID: "ISO27001",
				Title:      "Tiêu chuẩn ISO 27001",
				Content:    blockText,
				Embedding:  vector,
			}

			_, err = vectorCollection.InsertOne(context.TODO(), chunkDoc)
			if err == nil {
				successCount++
			}
			currentChunk.Reset()
		}

		currentChunk.WriteString(trimmed)
		currentChunk.WriteString("\n")
	}

	// Xử lý nốt phần văn bản còn dư lại cuối cùng
	if currentChunk.Len() > 20 {
		blockText := currentChunk.String()
		vector, err := GetTextEmbedding(blockText)
		if err == nil {
			chunkDoc := models.PlaybookVectorChunk{
				ID:         primitive.NewObjectID(),
				PlaybookID: "ISO27001",
				Title:      "Tiêu chuẩn ISO 27001",
				Content:    blockText,
				Embedding:  vector,
			}
			_, err = vectorCollection.InsertOne(context.TODO(), chunkDoc)
			if err == nil {
				successCount++
			}
		}
	}

	if successCount > 0 {
		fmt.Printf("✅ Đã nạp thành công [%d] đoạn dữ liệu ISO27001 vào Vector DB!\n", successCount)
	}
	return nil
}

func SearchRelevantContext(userQuery string) string {
	ctx := context.TODO()
	var contextBuilder strings.Builder

	var policies []models.Policy
	database.DB.Where("is_active = ? AND (title ILIKE ? OR value ILIKE ?)",
		true, "%"+userQuery+"%", "%"+userQuery+"%").
		Limit(3).
		Find(&policies)

	if len(policies) > 0 {
		contextBuilder.WriteString("\n--- CHÍNH SÁCH LIÊN QUAN ---\n")
		for _, p := range policies {
			contextBuilder.WriteString(fmt.Sprintf("- %s: %s\n", p.Title, p.Value))
		}
	}

	if database.DocumentContentCollection != nil {
		filter := bson.M{"$text": bson.M{"$search": userQuery}}
		opts := options.Find().SetLimit(2)

		cursor, err := database.DocumentContentCollection.Find(ctx, filter, opts)
		if err == nil {
			defer cursor.Close(ctx)

			var docs []struct {
				Content string `bson:"content"`
			}
			if err := cursor.All(ctx, &docs); err == nil && len(docs) > 0 {
				contextBuilder.WriteString("\n--- TRÍCH DẪN TÀI LIỆU PDF ---\n")
				for _, d := range docs {
					limit := 800
					if len(d.Content) < limit {
						limit = len(d.Content)
					}
					contextBuilder.WriteString(fmt.Sprintf("- ...%s...\n", d.Content[:limit]))
				}
			}
		}
	}

	return contextBuilder.String()
}

type OllamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type OllamaEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

func GetTextEmbedding(text string) ([]float32, error) {
	// Ép cấu hình gọi sang model nomic-embed-text của Ollama
	reqBody := OllamaEmbeddingRequest{
		Model:  "nomic-embed-text",
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("lỗi mã hóa json payload cho Ollama: %v", err)
	}

	// Chuyển sang endpoint của Ollama
	url := fmt.Sprintf("%s/api/embeddings", strings.TrimRight(OLLAMA_BASE, "/"))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("lỗi tạo request embeddings tới Ollama: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lỗi kết nối tới Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama embedding error (status %d): %s", resp.StatusCode, string(body))
	}

	var embedResp OllamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("lỗi giải mã dữ liệu Vector từ Ollama: %v", err)
	}

	if len(embedResp.Embedding) == 0 {
		return nil, fmt.Errorf("không nhận được dữ liệu vector từ Ollama")
	}

	return embedResp.Embedding, nil
}

func IngestPlaybookMarkdownToVectorDB(mdContent string) error {
	if database.DocumentContentCollection == nil {
		return fmt.Errorf("chưa kết nối cơ sở dữ liệu MongoDB")
	}

	vectorCollection := database.DocumentContentCollection.Database().Collection("playbook_vectors")

	// 1. DỌN SẠCH KHO CŨ: Chỉ xóa các bản ghi không phải là tài liệu ISO27001
	_, _ = vectorCollection.DeleteMany(context.TODO(), bson.M{"playbook_id": bson.M{"$ne": "ISO27001"}})

	successCount := 0
	// Tách kịch bản bằng thẻ tiêu đề "## "
	playbooks := strings.Split(mdContent, "## ")

	for _, pbBlock := range playbooks {
		if strings.TrimSpace(pbBlock) == "" || strings.HasPrefix(pbBlock, "# DANH SÁCH") {
			continue
		}

		fullBlockText := "## " + pbBlock
		lines := strings.Split(pbBlock, "\n")

		// 2. TỰ ĐỘNG TRÍCH XUẤT TIÊU ĐỀ LÀM ID ĐỊNH DANH
		playbookID := strings.TrimSpace(lines[0])
		if playbookID == "" {
			playbookID = "UNKNOWN_" + primitive.NewObjectID().Hex()
		}

		fmt.Printf("⏳ Đang nhúng dữ liệu toán học cho kịch bản: [%s]...\n", playbookID)

		vector, err := GetTextEmbedding(fullBlockText)
		if err != nil {
			fmt.Printf("❌ Lỗi sinh vector cho %s: %v\n", playbookID, err)
			continue
		}

		chunkDoc := models.PlaybookVectorChunk{
			ID:         primitive.NewObjectID(),
			PlaybookID: playbookID, // Lưu trực tiếp tiêu đề sạch làm ID định danh
			Title:      "Kịch bản ứng phó sự cố: " + playbookID,
			Content:    fullBlockText,
			Embedding:  vector,
		}

		_, err = vectorCollection.InsertOne(context.TODO(), chunkDoc)
		if err != nil {
			fmt.Printf("❌ Lỗi lưu trữ bản ghi Playbook vào MongoDB: %v\n", err)
			continue
		}
		successCount++
	}

	if successCount > 0 {
		fmt.Printf("✅ Toàn bộ [%d] kịch bản Playbooks đã được nạp thành công vào Vector DB!\n", successCount)
	}
	return nil
}

func CosineSimilarity(v1, v2 []float32) float32 {
	if len(v1) != len(v2) || len(v1) == 0 {
		return 0
	}
	var dotProduct, normA, normB float32
	for i := range v1 {
		dotProduct += v1[i] * v2[i]
		normA += v1[i] * v1[i]
		normB += v2[i] * v2[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// Tìm kiếm kịch bản Playbook phù hợp nhất bằng Vector Search
func SearchPlaybookByVector(userQuery string) string {
	queryVector, err := GetTextEmbedding(userQuery)
	if err != nil {
		fmt.Printf("⚠️ Lỗi tạo embedding cho query: %v\n", err)
		return ""
	}

	if database.DocumentContentCollection == nil {
		fmt.Println("⚠️ Lỗi tìm kiếm playbook: chưa kết nối cơ sở dữ liệu MongoDB.")
		return ""
	}
	vectorCollection := database.DocumentContentCollection.Database().Collection("playbook_vectors")

	// pipeline := mongo.Pipeline{
	// 	{{"$vectorSearch", bson.M{
	// 		"index":         "vector_index",
	// 		"path":          "embedding",
	// 		"queryVector":   queryVector,
	// 		"numCandidates": 100,
	// 		"limit":         3,
	// 	}}},
	// }

	// cursor, err := vectorCollection.Aggregate(context.TODO(), pipeline)
	// if err != nil {
	// 	fmt.Printf("⚠️ Lỗi truy vấn playbook vectors bằng $vectorSearch: %v\n", err)
	// 	// Fallback to basic loop if vector search fails (e.g., index not created or local mongo without Atlas)
	// 	return searchVectorFallback(queryVector, vectorCollection)
	// }
	// defer cursor.Close(context.TODO())

	// var bestContent strings.Builder
	// for cursor.Next(context.TODO()) {
	// 	var chunk models.PlaybookVectorChunk
	// 	if err := cursor.Decode(&chunk); err == nil {
	// 		bestContent.WriteString(chunk.Content)
	// 		bestContent.WriteString("\n\n")
	// 	}
	// }

	// if bestContent.Len() > 0 {
	// 	return fmt.Sprintf("\n--- [DỮ LIỆU TỪ HỆ THỐNG PLAYBOOK & ISO] ---\n%s\n", bestContent.String())
	// }

	// Fallback in case of empty aggregate response
	return searchVectorFallback(queryVector, vectorCollection)
}

func searchVectorFallback(queryVector []float32, vectorCollection *mongo.Collection) string {
	cursor, err := vectorCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		return ""
	}
	defer cursor.Close(context.TODO())

	var bestContent string
	var maxSimilarity float32 = -1.0

	for cursor.Next(context.TODO()) {
		var chunk models.PlaybookVectorChunk
		if err := cursor.Decode(&chunk); err == nil {
			similarity := CosineSimilarity(queryVector, chunk.Embedding)
			if similarity > maxSimilarity && similarity > 0.15 {
				maxSimilarity = similarity
				bestContent = chunk.Content
			}
		}
	}

	if bestContent != "" {
		return fmt.Sprintf("\n--- [DỮ LIỆU TỪ HỆ THỐNG PLAYBOOK & ISO] ---\n%s\n", bestContent)
	}
	return ""
}
