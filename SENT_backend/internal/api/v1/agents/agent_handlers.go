package agents

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

// PushDataHandler: Tiếp nhận dữ liệu từ Agent v3.1
func PushDataHandler(c *gin.Context) {
	var req struct {
		LogType     string      `json:"log_type"`
		CompanyCode string      `json:"company_code"` // Nhận Mã Công Ty
		HWID        string      `json:"hwid"`
		Hostname    string      `json:"hostname"`
		Data        interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// 1. XÁC THỰC CÔNG TY: Dùng CompanyCode để tìm OrgID
	var org models.Organization
	if err := database.DB.Where("company_code = ?", req.CompanyCode).First(&org).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Mã công ty không tồn tại hoặc sai"})
		return
	}

	// --- FIX LỖI FK_REGIONS_AGENTS TẠI ĐÂY ---
	// 1.5. TÌM VÙNG (REGION) MẶC ĐỊNH ĐỂ GÁN CHO AGENT
	var region models.Region
	// Thử tìm xem công ty này đã có vùng nào chưa (Lấy vùng đầu tiên tìm thấy)
	if err := database.DB.Where("org_id = ?", org.ID).First(&region).Error; err != nil {
		// Nếu chưa có (Công ty mới tinh), tạo tự động vùng "Trụ sở chính"
		region = models.Region{
			OrgID:       org.ID,
			Name:        "Trụ sở chính",
			EnrollToken: "AUTO-" + req.CompanyCode, // Tạo token ngẫu nhiên để tránh lỗi Unique
		}
		if err := database.DB.Create(&region).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khởi tạo vùng mặc định"})
			return
		}
	}
	// ------------------------------------------

	// 2. TÌM HOẶC TẠO AGENT
	var agent models.Agent
	result := database.DB.Where("hw_id = ?", req.HWID).First(&agent)

	if result.Error != nil {
		// Máy mới -> Tạo mới & Gán vào OrgID + RegionID vừa tìm được
		agent = models.Agent{
			HWID:      req.HWID,
			OrgID:     org.ID,    // Gán đúng OrgID
			RegionID:  region.ID, // <--- QUAN TRỌNG: Gán ID của vùng vừa tìm/tạo được
			Hostname:  req.Hostname,
			IPAddress: c.ClientIP(),
			Status:    "online",
			LastSeen:  time.Now(),
		}

		if err := database.DB.Create(&agent).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lưu Agent vào DB: " + err.Error()})
			return
		}
	} else {
		// Máy cũ -> Cập nhật trạng thái
		agent.IPAddress = c.ClientIP()
		agent.LastSeen = time.Now()
		agent.Status = "online"

		// Logic tự sửa lỗi: Nếu máy cũ bị lỗi mất OrgID hoặc RegionID thì cập nhật lại luôn
		if agent.OrgID == 0 {
			agent.OrgID = org.ID
		}
		if agent.RegionID == 0 {
			agent.RegionID = region.ID
		}

		database.DB.Save(&agent)
	}

	// 3. XỬ LÝ DỮ LIỆU LOG (Giữ nguyên logic cũ)
	switch req.LogType {
	case "inventory":
		service.ProcessInventory(agent, req.Data)
	case "telemetry":
		service.ProcessTelemetry(agent, req.Data)
	case "software":
		service.ProcessSoftware(agent, req.Data)
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed", "type": req.LogType})
}

// GetAgents: Lấy danh sách máy trạm kèm trạng thái Realtime
func GetAgents(c *gin.Context) {
	var agents []models.Agent
	if err := database.DB.Find(&agents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy dữ liệu"})
		return
	}

	// TÍNH TOÁN TRẠNG THÁI ĐỘNG: Nếu LastSeen cũ hơn 2 phút -> Đánh dấu Offline
	threshold := time.Now().Add(-2 * time.Minute)
	for i := range agents {
		if agents[i].LastSeen.After(threshold) {
			agents[i].Status = "online"
		} else {
			agents[i].Status = "offline"
		}
	}

	c.JSON(http.StatusOK, agents)
}

// GetAgentDetail: Lấy chi tiết kèm Inventory và Software
func GetAgentDetail(c *gin.Context) {
	hwid := c.Param("hwid")
	var agent models.Agent

	err := database.DB.Preload("Inventory").
		Preload("Software").
		Preload("Alerts").
		Where("hw_id = ?", hwid).
		First(&agent).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy thiết bị"})
		return
	}

	// TÍNH TOÁN TRẠNG THÁI ĐỘNG TRƯỚC KHI TRẢ VỀ
	threshold := time.Now().Add(-2 * time.Minute)
	if agent.LastSeen.After(threshold) {
		agent.Status = "online"
	} else {
		agent.Status = "offline"
	}

	c.JSON(http.StatusOK, agent)
}

// GetStats: Trả về số liệu Realtime cho Dashboard
func GetStats(c *gin.Context) {
	var total, online, alerts, regions int64

	// 1. Tổng số máy
	database.DB.Model(&models.Agent{}).Count(&total)

	// 2. TÍNH SỐ MÁY ONLINE: Chỉ đếm những máy có LastSeen trong 2 phút đổ lại đây
	threshold := time.Now().Add(-2 * time.Minute)
	database.DB.Model(&models.Agent{}).Where("last_seen >= ?", threshold).Count(&online)

	// 3. Các thông số khác
	database.DB.Model(&models.SecurityAlert{}).Where("is_resolved = ?", false).Count(&alerts)
	database.DB.Model(&models.Region{}).Count(&regions)

	c.JSON(http.StatusOK, gin.H{
		"total": total, "online": online, "alerts": alerts, "regions": regions,
	})
}
func GetAgentLogs(c *gin.Context) {
	hwid := c.Param("hwid")
	var alerts []models.SecurityAlert

	// Lấy tất cả cảnh báo của HWID này, sắp xếp mới nhất trước
	result := database.DB.Where("hw_id = ?", hwid).Order("created_at desc").Find(&alerts)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": "Lỗi truy vấn Log"})
		return
	}

	c.JSON(200, alerts)
}

// Lấy danh sách Whitelist của 1 máy
func GetAgentWhitelist(c *gin.Context) {
	var list []models.AgentWhitelist
	database.DB.Where("hwid = ?", c.Param("hwid")).Find(&list)
	c.JSON(200, list)
}

func AddAgentWhitelist(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}
	item := models.AgentWhitelist{HWID: c.Param("hwid"), SoftwareName: req.Name}
	database.DB.Create(&item)
	c.JSON(200, item)
}

func DeleteAgentWhitelist(c *gin.Context) {
	database.DB.Delete(&models.AgentWhitelist{}, c.Param("id"))
	c.JSON(200, gin.H{"status": "ok"})
}

// Hàm mới: CẤP PHÉP HÀNG LOẠT CHO NHIỀU MÁY CÙNG LÚC
func AddBulkWhitelist(c *gin.Context) {
	var req struct {
		HWIDs        []string `json:"hwids"`
		SoftwareName string   `json:"software_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// Lặp qua danh sách các máy được tích chọn và lưu vào DB
	for _, hwid := range req.HWIDs {
		item := models.AgentWhitelist{HWID: hwid, SoftwareName: req.SoftwareName}
		database.DB.Create(&item)
	}
	c.JSON(200, gin.H{"status": "success"})
}
func AssignManager(c *gin.Context) {
	hwid := c.Param("hwid")
	var req struct {
		UserID uint `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// Cập nhật trường user_id cho Agent có HWID tương ứng
	if err := database.DB.Model(&models.Agent{}).Where("hw_id = ?", hwid).Update("user_id", req.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể phân bổ quản lý"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã phân bổ người quản lý thành công"})
}
