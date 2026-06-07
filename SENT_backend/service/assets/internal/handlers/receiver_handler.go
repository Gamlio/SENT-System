package handlers

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	assetService "SENT_backend/service/assets/internal/service"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Payload đón dữ liệu thô
type assetPayload struct {
	LogType  string          `json:"log_type" binding:"required,max=50"`
	AssetID  string          `json:"asset_hwid" binding:"required,max=64"`
	Hostname string          `json:"hostname" binding:"required,min=1,max=255"`
	Data     json.RawMessage `json:"data" binding:"required"`
}

// PushDataHandler: Cổng tiếp nhận duy nhất
func PushDataHandler(c *gin.Context) {
	// [BẢO MẬT & HIỆU NĂNG] Middleware AssetHMACAuth đã xử lý HMAC, Replay Attack và đọc an toàn Body.
	// Lấy trực tiếp dữ liệu thô từ Context để tránh đọc lại Body và ngăn chặn rủi ro DoS (OOM).
	rawBody, exists := c.Get("raw_body")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể truy xuất dữ liệu đã xác thực"})
		return
	}

	bodyBytes := rawBody.([]byte)

	// 1. Bind dữ liệu
	var req assetPayload
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// [SỬA LỖI]: Phân loại Type dựa trên LogType để kích hoạt luồng xử lý nhanh trong Service
	// Nếu không, mọi gói tin đều bị coi là DATA và phải truy vấn DB nặng nề
	payloadType := "DATA"
	if req.LogType == "heartbeat" {
		payloadType = "HEARTBEAT"
	} else if req.LogType == "offline" {
		payloadType = "OFFLINE"
	}

	// 3. Đẩy dữ liệu cho Bộ não xử lý và bắt trọn các lỗi trả về
	err := assetService.ProcessassetData(assetService.AssetPayload{
		Type:     payloadType,
		LogType:  req.LogType,
		AssetID:  req.AssetID,
		Hostname: req.Hostname,
		Data:     req.Data,
	})

	if err != nil {
		log.Printf(" Lỗi xử lý máy %s: %v", req.AssetID, err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// 4. Phản hồi ngay lập tức cho asset
	c.JSON(http.StatusOK, gin.H{"message": "Dữ liệu đã được tiếp nhận và lưu trữ an toàn"})
}
func UpdateDepartment(c *gin.Context) {
	hwid := c.Param("hwid")
	var req struct {
		DepartmentTag string `json:"department_tag" binding:"required,min=1,max=50,-"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	// [BẢO MẬT] Fix lỗi IDOR: Ép buộc cập nhật phải thuộc về OrgID của người gọi lệnh
	orgID := c.GetUint("org_id")

	if err := database.DB.Model(&models.Asset{}).Where("asset_hwid = ? AND org_id = ?", hwid, orgID).Update("department_tag", req.DepartmentTag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi cập nhật phòng ban"})
		return
	}

	// Notify Behavior Service to trigger a centralized recalculation (Behavior will call Scoring)
	go func(id string) {
		payload := map[string]interface{}{
			"org_id":        c.GetUint("org_id"),
			"asset_hwid":    id,
			"category":      "DepartmentChange",
			"value":         "",
			"title":         "Department tag updated",
			"desc":          "Department updated, request centralized scoring",
			"base_priority": "P3",
		}
		b, _ := json.Marshal(payload)
		url := "http://behavior-service:8000/api/v1/behaviors/log"
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(b))
		if err != nil {
			log.Printf("⚠️ Lỗi gọi Behavior Service cho máy %s: %v\n", id, err)
			return
		}
		resp.Body.Close()
	}(hwid)

	c.JSON(http.StatusOK, gin.H{"message": "Đã cập nhật Nhãn bảo mật (Department)"})
}

// =========================================================================
// SERVER-SIDE PAGINATION ENDPOINTS (Cho các Tab Chi tiết Máy trạm)
// =========================================================================

func GetAssetSoftware(c *gin.Context) {
	hwid := c.Param("hwid")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	skip := int64((page - 1) * limit)
	limit64 := int64(limit)

	filter := bson.M{"asset_hwid": hwid}
	// Bắt buộc ép kiểu sang int64 để truy vấn chính xác trên MongoDB
	filter["org_id"] = int64(c.GetUint("org_id"))

	if search != "" {
		filter["$or"] = []bson.M{
			{"software_name": primitive.Regex{Pattern: search, Options: "i"}},
			{"publisher": primitive.Regex{Pattern: search, Options: "i"}},
		}
	}

	if database.SoftwareCollection == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database chưa sẵn sàng"})
		return
	}

	total, _ := database.SoftwareCollection.CountDocuments(c, filter)

	cursor, err := database.SoftwareCollection.Find(c, filter, &options.FindOptions{
		Skip:  &skip,
		Limit: &limit64,
		Sort:  bson.M{"updated_at": -1}, // Ưu tiên bản ghi mới cập nhật
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy vấn dữ liệu"})
		return
	}
	defer cursor.Close(c)

	var items []bson.M
	if err = cursor.All(c, &items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi đọc dữ liệu"})
		return
	}
	if items == nil {
		items = []bson.M{}
	}

	c.JSON(http.StatusOK, gin.H{"total": total, "page": page, "limit": limit, "items": items})
}

func GetAssetUSB(c *gin.Context) {
	hwid := c.Param("hwid")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	skip := int64((page - 1) * limit)
	limit64 := int64(limit)

	filter := bson.M{"asset_hwid": hwid}
	// Bắt buộc ép kiểu sang int64 để truy vấn chính xác trên MongoDB
	filter["org_id"] = int64(c.GetUint("org_id"))

	if search != "" {
		filter["$or"] = []bson.M{
			{"device_name": primitive.Regex{Pattern: search, Options: "i"}},
			{"vid": primitive.Regex{Pattern: search, Options: "i"}},
			{"pid": primitive.Regex{Pattern: search, Options: "i"}},
		}
	}

	total, _ := database.USBCollection.CountDocuments(c, filter)
	cursor, _ := database.USBCollection.Find(c, filter, &options.FindOptions{
		Skip:  &skip,
		Limit: &limit64,
		Sort:  bson.M{"updated_at": -1},
	})
	defer cursor.Close(c)

	var items []bson.M
	cursor.All(c, &items)
	if items == nil {
		items = []bson.M{}
	}

	c.JSON(http.StatusOK, gin.H{"total": total, "page": page, "limit": limit, "items": items})
}

func GetAssetPorts(c *gin.Context) {
	hwid := c.Param("hwid")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	skip := int64((page - 1) * limit)
	limit64 := int64(limit)

	filter := bson.M{"asset_hwid": hwid}
	// Bắt buộc ép kiểu sang int64 để truy vấn chính xác trên MongoDB
	filter["org_id"] = int64(c.GetUint("org_id"))

	if search != "" {
		orConditions := []bson.M{
			{"process_name": primitive.Regex{Pattern: search, Options: "i"}},
		}
		// Nếu search là một con số, có thể họ đang tìm số Port
		if portInt, err := strconv.Atoi(search); err == nil {
			orConditions = append(orConditions, bson.M{"port": portInt})
		}
		filter["$or"] = orConditions
	}

	total, _ := database.OpenPortCollection.CountDocuments(c, filter)
	// Gom nhóm cổng nào mới cập nhật nhất đưa lên đầu
	cursor, _ := database.OpenPortCollection.Find(c, filter, &options.FindOptions{
		Skip:  &skip,
		Limit: &limit64,
		Sort:  bson.M{"updated_at": -1},
	})
	defer cursor.Close(c)

	var items []bson.M
	cursor.All(c, &items)
	if items == nil {
		items = []bson.M{}
	}

	c.JSON(http.StatusOK, gin.H{"total": total, "page": page, "limit": limit, "items": items})
}
func InternalCleanupHandler(c *gin.Context) {
	var req struct {
		HWIDs []string `json:"hwids"`
		OrgID uint     `json:"org_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// Gọi logic dọn dẹp MongoDB đã có của bạn
	lifecycleSvc := assetService.AssetLifecycleService{}
	lifecycleSvc.CleanupassetTelemetry(req.HWIDs, req.OrgID)

	c.JSON(200, gin.H{"status": "Cleanup started"})
}
