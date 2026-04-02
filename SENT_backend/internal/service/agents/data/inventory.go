package data

import (
	"context"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ProcessInventory(agent models.Agent, data interface{}) {
	invData, ok := data.(map[string]interface{})
	if !ok || database.AgentInventoryCollection == nil {
		return
	}

	agentHWID := agent.HWID
	update := bson.M{"$set": bson.M{
		"agent_hwid": agentHWID,
		"os_info":    fmt.Sprintf("%v", invData["os_info"]),
		"cpu_model":  fmt.Sprintf("%v", invData["cpu_model"]),
		"updated_at": time.Now(),
	}}

	if ramFloat, ok := invData["ram_total_gb"].(float64); ok {
		update["$set"].(bson.M)["ram_total_gb"] = int(ramFloat)
	}

	_, _ = database.AgentInventoryCollection.UpdateOne(context.TODO(), bson.M{"agent_hwid": agentHWID}, update, options.Update().SetUpsert(true))
}
