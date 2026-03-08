package agents

import (
	"fmt"
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/agent_data"
	"sent_backend/internal/service/scoring" // <--- Import service tính điểm
	"sent_backend/internal/service/security"
	"time"

	"github.com/gin-gonic/gin"
)

// PushDataHandler: Tiếp nhận dữ liệu từ Agent v3.1
func PushDataHandler(c *gin.Context) {
	var req struct {
		LogType     string      `json:"log_type"`
		CompanyCode string      `json:"company_code"`
		HWID        string      `json:"hwid"`
		Hostname    string      `json:"hostname"`
		Data        interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// 1. XÁC THỰC CÔNG TY
	var org models.Organization
	if err := database.DB.Where("company_code = ?", req.CompanyCode).First(&org).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Mã công ty không tồn tại hoặc sai"})
		return
	}

	// 1.5. TÌM VÙNG (REGION) MẶC ĐỊNH
	var region models.Region
	if err := database.DB.Where("org_id = ?", org.ID).First(&region).Error; err != nil {
		region = models.Region{
			OrgID:       org.ID,
			Name:        "Trụ sở chính",
			EnrollToken: "AUTO-" + req.CompanyCode,
		}
		database.DB.Create(&region)
	}

	// 2. TÌM HOẶC TẠO AGENT
	var agent models.Agent
	result := database.DB.Where("hw_id = ?", req.HWID).First(&agent)

	if result.Error != nil {
		// Máy mới -> Tạo mới (Lần đầu thì lấy IP kết nối làm tạm)
		agent = models.Agent{
			HWID:      req.HWID,
			OrgID:     org.ID,
			RegionID:  region.ID,
			Hostname:  req.Hostname,
			IPAddress: c.ClientIP(), // Tạm thời lấy IP kết nối
			Status:    "online",
			LastSeen:  time.Now(),
		}
		if err := database.DB.Create(&agent).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lưu Agent: " + err.Error()})
			return
		}
	} else {
		// Máy cũ -> Cập nhật trạng thái
		updates := map[string]interface{}{
			// [QUAN TRỌNG] ĐÃ XÓA DÒNG "ip_address": c.ClientIP()
			// Để không ghi đè IP thật mà Telemetry đã gửi lên
			"last_seen": time.Now(),
			"status":    "online",
		}

		// Chỉ cập nhật IP từ kết nối nếu IP trong DB đang rỗng hoặc lỗi
		if agent.IPAddress == "" || agent.IPAddress == "::1" || agent.IPAddress == "127.0.0.1" {
			updates["ip_address"] = c.ClientIP()
		}

		// Tự sửa lỗi mất OrgID/RegionID
		if agent.OrgID == 0 {
			updates["org_id"] = org.ID
		}
		if agent.RegionID == 0 {
			updates["region_id"] = region.ID
		}

		database.DB.Model(&agent).Updates(updates)
	}
	// 3. XỬ LÝ DỮ LIỆU LOG & KÍCH HOẠT TÍNH ĐIỂM
	switch req.LogType {
	case "inventory":
		agent_data.ProcessInventory(agent, req.Data)
	case "telemetry":
		agent_data.ProcessTelemetry(agent, req.Data)
		go func() {
			security.AnalyzeBehaviorAI(agent.HWID, "telemetry", req.Data)
			scoring.RecalculateRiskScore(agent.HWID) // <--- Cập nhật điểm sau khi phân tích
		}()
	case "software":
		agent_data.ProcessSoftware(agent, req.Data)
		go func() {
			security.CheckSoftwareCompliance(agent, req.Data)
			scoring.RecalculateRiskScore(agent.HWID) // <--- Cập nhật điểm
		}()
	case "usb":
		agent_data.ProcessUSB(agent, req.Data)
		go func() {
			security.AnalyzeBehaviorAI(agent.HWID, "usb", req.Data)
			scoring.RecalculateRiskScore(agent.HWID) // <--- Cập nhật điểm
		}()
	case "alert":
		alertMap, ok := req.Data.(map[string]interface{})
		if ok {
			// 1. Lấy dữ liệu từ Agent gửi lên
			alertType := fmt.Sprintf("%v", alertMap["alert_type"])
			desc := fmt.Sprintf("%v", alertMap["message"])

			// 2. MAPPING THEO MỨC ĐỘ ƯU TIÊN (Khớp với 5 Use-case chuẩn SOC)
			priority := "P4"
			severity := "Low"

			switch alertType {
			case "Firewall Disabled": // [MỚI] Tắt tường lửa
				priority = "P1"
				severity = "Critical"
			case "Malware/AV Alert": // [MỚI] Mã độc / Tắt Antivirus
				priority = "P2"
				severity = "High"
			case "Unpatched OS": // [MỚI] Thiếu bản vá Windows
				priority = "P3"
				severity = "Medium"
			case "Unauthorized Port": // Mở cổng mạng nguy hiểm
				priority = "P2"
				severity = "High"
			case "Software Violation": // Phần mềm cấm
				priority = "P3"
				severity = "Medium"
			case "USB Violation": // Cắm USB lạ
				priority = "P3"
				severity = "Medium"
			}

			// 3. Tạo Alert lưu vào DB
			newAlert := models.SecurityAlert{
				OrgID:       agent.OrgID,
				HWID:        agent.HWID,
				AlertType:   alertType,
				Title:       fmt.Sprintf("[%s] %s", priority, alertType),
				Description: desc,
				Severity:    severity,
				Priority:    priority,
				IsResolved:  false,
			}
			database.DB.Create(&newAlert)

			// 4. Gọi hàm Gom nhóm Alert vào Incident (Nếu bạn đã tạo hàm này ở file security)
			go security.GroupAlertToIncident(&newAlert)

			// 5. Cập nhật điểm rủi ro
			go scoring.RecalculateRiskScore(agent.HWID)
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed", "type": req.LogType})
}
