package assets

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	assetService "sent_backend/internal/service/assets"
	"sent_backend/internal/service/scoring"
	"strconv"
	"time"

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
	// Lấy toàn bộ body gốc để xác thực chữ ký (SENT ký trên toàn bộ payload)
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Không thể đọc dữ liệu yêu cầu"})
		return
	}

	// 1. Bind dữ liệu
	var req assetPayload
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// 2. Xác thực danh tính thiết bị
	var asset models.Asset
	if err := database.DB.Where("asset_hwid = ?", req.AssetID).First(&asset).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Thiết bị không tồn tại"})
		return
	}

	// [BẢO MẬT] Xác thực HMAC & Chống Replay Attack
	signature := c.GetHeader("X-Sent-Signature")
	timestampStr := c.GetHeader("X-Sent-Timestamp")

	if signature == "" || timestampStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yêu cầu từ chối: Thiếu Headers bảo mật"})
		return
	}

	// 1. Kiểm tra Timestamp Drift (Chống Replay Attack - trễ tối đa 5 phút)
	ts, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil || time.Since(time.Unix(ts, 0)).Abs() > 5*time.Minute {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yêu cầu đã hết hạn (Replay Attack Detected)"})
		return
	}

	// 2. Xác thực HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(asset.SecretKey))
	mac.Write(bodyBytes) // Dùng TOÀN BỘ raw body thay vì chỉ req.Data
	if !hmac.Equal([]byte(signature), []byte(hex.EncodeToString(mac.Sum(nil)))) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Chữ ký số không hợp lệ"})
		return
	}

	// 3. Quăng dữ liệu vào hàng đợi bất đồng bộ (Goroutine) cho Bộ não xử lý
	go assetService.ProcessassetData(assetService.AssetPayload{
		Type:     "DATA",
		LogType:  req.LogType,
		AssetID:  req.AssetID,
		Hostname: req.Hostname,
		Data:     req.Data,
	})

	// 4. Phản hồi ngay lập tức cho asset
	c.JSON(http.StatusOK, gin.H{"message": "Dữ liệu đã được tiếp nhận", "status": asset.Status})
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

	// [TÙY CHỌN] Tính lại điểm rủi ro ngay lập tức vì đổi ngữ cảnh có thể làm thay đổi P1/P4
	scoring.RecalculateRiskScore(hwid)

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
	if orgID, exists := c.Get("org_id"); exists {
		filter["org_id"] = orgID
	}

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
	if orgID, exists := c.Get("org_id"); exists {
		filter["org_id"] = orgID
	}

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
	if orgID, exists := c.Get("org_id"); exists {
		filter["org_id"] = orgID
	}

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
