package service

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AutoCheckAndLoadPlaybooks() {
	if database.DocumentContentCollection == nil {
		fmt.Println("❌ [VECTOR] MongoDB connection not ready for playbook ingestion.")
		return
	}

	vectorCollection := database.DocumentContentCollection.Database().Collection("playbook_vectors")

	count, err := vectorCollection.CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		fmt.Printf("❌ [VECTOR] Error querying playbook vector database: %v\n", err)
		return
	}

	if count > 0 {
		fmt.Printf("✅ [VECTOR] Playbook vector database is operational. Found [%d] playbook vectors.\n", count)
		return
	}

	fmt.Println("⚠️ [VECTOR] Playbook vector storage is empty (0 records)! Attempting to find and load playbook file...")

	possiblePaths := []string{
		"doc/playbook.md",
		"service/ai/doc/playbook.md",
		"../doc/playbook.md",
	}

	var targetPath string
	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err == nil {
			targetPath = p
			break
		}
	}

	if targetPath == "" {
		fmt.Println("❌ [VECTOR] CRITICAL: Could not find 'playbook.md' in any of the expected locations. Vector search will be unavailable.")
		return
	}

	contentBytes, err := os.ReadFile(targetPath)
	if err != nil {
		fmt.Printf("❌ [VECTOR] Failed to read playbook file content at %s: %v\n", targetPath, err)
		return
	}

	fmt.Printf("🚀 [VECTOR] Found valid playbook file at [%s]. Starting ingestion process into Vector DB...\n", targetPath)

	err = IngestPlaybookMarkdownToVectorDB(string(contentBytes))
	if err != nil {
		fmt.Printf("❌ [VECTOR] Playbook vector synchronization process failed: %v\n", err)
	} else {
		fmt.Println("✅ [VECTOR] Semantic structure conversion and ingestion completed successfully!")
	}
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
	Model string `json:"model"`
	Input string `json:"input"`
}

type OllamaEmbeddingResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func GetTextEmbedding(text string) ([]float32, error) {
	if AIProvider == "gemini" {
		return GetGeminiEmbedding(text)
	}
	return getOllamaEmbedding(text)
}

func getOllamaEmbedding(text string) ([]float32, error) {
	reqBody := OllamaEmbeddingRequest{
		Model: MODEL_EMBED,
		Input: text,
	}

	jsonData, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/api/embed", strings.TrimRight(OLLAMA_BASE, "/"))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("lỗi kết nối Ollama Embeddings: %v", err)
	}
	defer resp.Body.Close()

	var embedResp OllamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("lỗi giải mã dữ liệu Vector từ AI: %v", err)
	}

	if len(embedResp.Embeddings) == 0 {
		return nil, fmt.Errorf("không nhận được dữ liệu vector từ Ollama")
	}

	return embedResp.Embeddings[0], nil
}

func GetGeminiEmbedding(text string) ([]float32, error) {
	if GEMINI_API_KEY == "" {
		return nil, fmt.Errorf("missing Gemini API key")
	}

	reqBody := map[string]interface{}{
		"model": MODEL_EMBED,
		"input": text,
	}

	jsonData, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/models/%s:embedText", strings.TrimRight(GEMINI_BASE, "/"), MODEL_EMBED)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("lỗi tạo request Gemini embedding: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+GEMINI_API_KEY)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lỗi kết nối Gemini Embeddings: %v", err)
	}
	defer resp.Body.Close()

	var embedResp GeminiEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("lỗi giải mã embedding Gemini: %v", err)
	}

	if len(embedResp.Embeddings) > 0 {
		return embedResp.Embeddings[0], nil
	}
	if len(embedResp.Embedding) > 0 {
		return embedResp.Embedding, nil
	}

	return nil, fmt.Errorf("không nhận được dữ liệu vector từ Gemini")
}

func IngestPlaybookMarkdownToVectorDB(mdContent string) error {
	if database.DocumentContentCollection == nil {
		return fmt.Errorf("chưa kết nối cơ sở dữ liệu MongoDB")
	}

	vectorCollection := database.DocumentContentCollection.Database().Collection("playbook_vectors")

	// 1. DỌN SẠCH KHO CŨ: Xóa bỏ triệt để dữ liệu rác trước đó
	_, _ = vectorCollection.DeleteMany(context.TODO(), bson.M{})

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
			return fmt.Errorf("lỗi lưu trữ bản ghi vào Vector DB: %v", err)
		}
	}

	fmt.Println("✅ Toàn bộ hệ thống Playbooks tối giản đã được nạp thành công!")
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
	cursor, err := vectorCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Printf("⚠️ Lỗi truy vấn playbook vectors: %v\n", err)
		return ""
	}
	defer cursor.Close(context.TODO())

	var bestContent string
	var maxSimilarity float32 = -1.0

	for cursor.Next(context.TODO()) {
		var chunk models.PlaybookVectorChunk
		if err := cursor.Decode(&chunk); err == nil {
			similarity := CosineSimilarity(queryVector, chunk.Embedding)
			if similarity > maxSimilarity && similarity > 0.2 {
				maxSimilarity = similarity
				bestContent = chunk.Content
			}
		}
	}

	if bestContent != "" {
		return fmt.Sprintf("\n--- [QUY TRÌNH PHẢN ỨNG CHUẨN (PLAYBOOK SOP)] ---\n%s\n", bestContent)
	}
	return ""
}
