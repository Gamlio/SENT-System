package data

import (
	"context"
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/incidents"
	"sync"
)

// Cache lưu giá trị cũ để tính toán Delta (tốc độ)
var ioCache = make(map[string]models.AgentIOActivity)
var ioMutex sync.Mutex

func ProcessDataTransfer(agent models.Agent, data interface{}) {
	var current models.AgentIOActivity
	bytes, _ := json.Marshal(data)
	if err := json.Unmarshal(bytes, &current); err != nil {
		return
	}

	ioMutex.Lock()
	lastRecord, exists := ioCache[agent.HWID]
	ioCache[agent.HWID] = current // Cập nhật giá trị mới nhất vào cache
	ioMutex.Unlock()

	if !exists {
		return // Lần đầu tiên chỉ lưu cache để lấy mốc so sánh
	}

	// 1. TÍNH TOÁN DELTA (Lượng dữ liệu phát sinh trong 30 giây qua)
	diffSent := current.NetBytesSent - lastRecord.NetBytesSent
	diffWrite := current.DiskBytesWritten - lastRecord.DiskBytesWritten

	// 2. NGƯỠNG CẢNH BÁO (Ví dụ: Gửi > 500MB hoặc Ghi > 1GB trong 30 giây)
	// Đây mới là con số phản ánh hoạt động "truyền tải/lưu trữ dữ liệu lớn" thực tế
	isNetworkSpike := diffSent > 500*1024*1024
	isDiskSpike := diffWrite > 1024*1024*1024

	if isNetworkSpike || isDiskSpike {
		// 3. CHỈ LƯU VÀO MongoDB KHI CÓ BẤT THƯỜNG
		current.AgentHWID = agent.HWID
		if database.AgentIOActivityCollection != nil {
			_, _ = database.AgentIOActivityCollection.InsertOne(context.TODO(), current)
		}

		// 4. KÍCH HOẠT SỰ CỐ (Chỉ hiển thị ở phần Incidents)
		incSvc := &incidents.IncidentService{}
		reason := "Phát hiện lưu lượng mạng đột biến (Nghi vấn rò rỉ dữ liệu)"
		if isDiskSpike {
			reason = "Phát hiện ghi đĩa khối lượng lớn (Nghi vấn mã hóa Ransomware hoặc Copy dữ liệu)"
		}

		incSvc.TriggerSecurityEvent(agent,
			"Data Exfiltration/Intensive I/O",
			"[P1] Hoạt động I/O bất thường",
			fmt.Sprintf("%s. Lượng dữ liệu: %d MB/30s", reason, diffSent/1024/1024),
			"P1",
		)
	}
}
