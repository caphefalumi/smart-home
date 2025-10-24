package services

import (
	"context"
	"fmt"
	"time"

	"github.com/caphefalumi/smart-home/database"
	"github.com/caphefalumi/smart-home/models"
	"go.mongodb.org/mongo-driver/bson"
)

// SensorService handles sensor data operations
type SensorService struct {
	db         *database.Database
	collection string
}

// NewSensorService creates a new sensor service
func NewSensorService(db *database.Database) *SensorService {
	return &SensorService{
		db:         db,
		collection: "sensordatas",
	}
}

// SaveSensorData saves sensor data to MongoDB
func (s *SensorService) SaveSensorData(data *models.SensorData) error {
	ctx := context.Background()
	coll := s.db.GetCollection(s.collection)

	_, err := coll.InsertOne(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to save sensor data: %w", err)
	}

	return nil
}

// SaveBulkSensorData saves multiple sensor readings efficiently
func (s *SensorService) SaveBulkSensorData(data []models.SensorData) error {
	if len(data) == 0 {
		return nil
	}

	ctx := context.Background()
	coll := s.db.GetCollection(s.collection)

	// Convert to interface slice for bulk insert
	docs := make([]interface{}, len(data))
	for i, item := range data {
		docs[i] = item
	}

	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to save bulk sensor data: %w", err)
	}

	return nil
}

// GetSensorHistory retrieves paginated sensor history
func (s *SensorService) GetSensorHistory(limit, skip int, startDate, endDate *time.Time) ([]models.SensorData, int64, error) {
	ctx := context.Background()
	coll := s.db.GetCollection(s.collection)

	// Build query filter
	filter := bson.M{}
	if startDate != nil || endDate != nil {
		timeFilter := bson.M{}
		if startDate != nil {
			timeFilter["$gte"] = *startDate
		}
		if endDate != nil {
			timeFilter["$lte"] = *endDate
		}
		filter["timestamp"] = timeFilter
	}
	// Get total count
	total, err := coll.Find(ctx, filter).Count()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count sensor data: %w", err)
	}

	// Get paginated results
	var results []models.SensorData
	err = coll.Find(ctx, filter).Sort("-timestamp").Limit(int64(limit)).Skip(int64(skip)).All(&results)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get sensor history: %w", err)
	}

	return results, total, nil
}

// GetStatistics calculates sensor statistics for the specified time range
func (s *SensorService) GetStatistics(sensorType string, hours int) (*models.Statistics, error) {
	ctx := context.Background()
	coll := s.db.GetCollection(s.collection)

	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	var pipeline []bson.M

	// Match time range
	pipeline = append(pipeline, bson.M{
		"$match": bson.M{
			"timestamp": bson.M{"$gte": startTime},
		},
	})

	// Group and calculate statistics
	if sensorType != "" {
		// Statistics for specific sensor
		pipeline = append(pipeline, bson.M{
			"$group": bson.M{
				"_id":   nil,
				"mean":  bson.M{"$avg": "$" + sensorType},
				"min":   bson.M{"$min": "$" + sensorType},
				"max":   bson.M{"$max": "$" + sensorType},
				"count": bson.M{"$sum": 1},
			},
		})
	} else {
		// Statistics for all sensors
		pipeline = append(pipeline, bson.M{
			"$group": bson.M{
				"_id":        nil,
				"light_mean": bson.M{"$avg": "$light"},
				"light_min":  bson.M{"$min": "$light"},
				"light_max":  bson.M{"$max": "$light"},
				"gas_mean":   bson.M{"$avg": "$gas"},
				"gas_min":    bson.M{"$min": "$gas"},
				"gas_max":    bson.M{"$max": "$gas"},
				"soil_mean":  bson.M{"$avg": "$soil"},
				"soil_min":   bson.M{"$min": "$soil"},
				"soil_max":   bson.M{"$max": "$soil"},
				"water_mean": bson.M{"$avg": "$water"},
				"water_min":  bson.M{"$min": "$water"},
				"water_max":  bson.M{"$max": "$water"},
				"count":      bson.M{"$sum": 1},
			},
		})
	}

	var results []bson.M
	err := coll.Aggregate(ctx, pipeline).All(&results)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate statistics: %w", err)
	}

	if len(results) == 0 {
		return &models.Statistics{Count: 0}, nil
	}

	result := results[0]
	stats := &models.Statistics{}

	// Convert results to Statistics struct
	if count, ok := result["count"]; ok {
		if c, ok := count.(int32); ok {
			stats.Count = int(c)
		}
	}

	// Handle specific sensor statistics
	if sensorType != "" {
		if mean, ok := result["mean"]; ok {
			if m, ok := mean.(float64); ok {
				switch sensorType {
				case "light":
					stats.LightMean = m
				case "gas":
					stats.GasMean = m
				case "soil":
					stats.SoilMean = m
				case "water":
					stats.WaterMean = m
				}
			}
		}
		// Similar for min and max...
	} else {
		// Handle all sensor statistics
		if lightMean, ok := result["light_mean"]; ok {
			if m, ok := lightMean.(float64); ok {
				stats.LightMean = m
			}
		}
		if gasMean, ok := result["gas_mean"]; ok {
			if m, ok := gasMean.(float64); ok {
				stats.GasMean = m
			}
		}
		if soilMean, ok := result["soil_mean"]; ok {
			if m, ok := soilMean.(float64); ok {
				stats.SoilMean = m
			}
		}
		if waterMean, ok := result["water_mean"]; ok {
			if m, ok := waterMean.(float64); ok {
				stats.WaterMean = m
			}
		}
		// Add min/max handling...
	}

	return stats, nil
}

// GetTrends calculates hourly trends for the specified time range
func (s *SensorService) GetTrends(hours int) ([]models.TrendData, error) {
	ctx := context.Background()
	coll := s.db.GetCollection(s.collection)

	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"timestamp": bson.M{"$gte": startTime},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"$dateToString": bson.M{
						"format": "%Y-%m-%d %H:00",
						"date":   "$timestamp",
					},
				},
				"light": bson.M{"$avg": "$light"},
				"gas":   bson.M{"$avg": "$gas"},
				"soil":  bson.M{"$avg": "$soil"},
				"water": bson.M{"$avg": "$water"},
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	var results []models.TrendData
	err := coll.Aggregate(ctx, pipeline).All(&results)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate trends: %w", err)
	}

	return results, nil
}

// GetAlerts retrieves recent sensor data with alerts
func (s *SensorService) GetAlerts(limit int) ([]models.SensorData, error) {
	ctx := context.Background()
	coll := s.db.GetCollection(s.collection)

	filter := bson.M{
		"alerts": bson.M{
			"$exists": true,
			"$not":    bson.M{"$size": 0},
		},
	}

	var results []models.SensorData
	err := coll.Find(ctx, filter).Sort("-timestamp").Limit(int64(limit)).All(&results)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}

	return results, nil
}
