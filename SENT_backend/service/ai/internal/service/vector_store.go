package service

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
