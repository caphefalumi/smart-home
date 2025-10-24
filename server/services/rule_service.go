package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/caphefalumi/smart-home/database"
	"github.com/caphefalumi/smart-home/models"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RuleService handles rule operations
type RuleService struct {
	db         *database.Database
	collection string
}

// NewRuleService creates a new rule service
func NewRuleService(db *database.Database) *RuleService {
	return &RuleService{
		db:         db,
		collection: "rules",
	}
}

// GetAllRules retrieves all rules
func (r *RuleService) GetAllRules() ([]models.Rule, error) {
	ctx := context.Background()
	coll := r.db.GetCollection(r.collection)

	var rules []models.Rule
	err := coll.Find(ctx, bson.M{}).Sort("-createdAt").All(&rules)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}

	return rules, nil
}

// CreateRule creates a new rule
func (r *RuleService) CreateRule(rule *models.Rule) error {
	ctx := context.Background()
	coll := r.db.GetCollection(r.collection)

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	_, err := coll.InsertOne(ctx, rule)
	if err != nil {
		return fmt.Errorf("failed to create rule: %w", err)
	}

	return nil
}

// UpdateRule updates an existing rule
func (r *RuleService) UpdateRule(id string, updates map[string]interface{}) (*models.Rule, error) {
	ctx := context.Background()
	coll := r.db.GetCollection(r.collection)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid rule ID: %w", err)
	}

	updates["updatedAt"] = time.Now()

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": updates}

	var rule models.Rule
	err = coll.Find(ctx, filter).Apply(qmgo.Change{
		Update:    update,
		ReturnNew: true,
	}, &rule)

	if err != nil {
		return nil, fmt.Errorf("failed to update rule: %w", err)
	}

	return &rule, nil
}

// DeleteRule deletes a rule by ID
func (r *RuleService) DeleteRule(id string) error {
	ctx := context.Background()
	coll := r.db.GetCollection(r.collection)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid rule ID: %w", err)
	}

	err = coll.RemoveId(ctx, objectID)
	if err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	return nil
}

// EvaluateRules evaluates all active rules against sensor data
func (r *RuleService) EvaluateRules(sensorData *models.SensorReading) []string {
	rules, err := r.GetAllRules()
	if err != nil {
		log.Printf("Error getting rules for evaluation: %v", err)
		return nil
	}

	var triggeredActions []string
	var alerts []string

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		var sensorValue int
		switch rule.Sensor {
		case "gas":
			sensorValue = sensorData.Gas
		case "light":
			sensorValue = sensorData.Light
		case "soil":
			sensorValue = sensorData.Soil
		case "water":
			sensorValue = sensorData.Water
		default:
			continue
		}

		triggered := false
		switch rule.Operator {
		case ">":
			triggered = sensorValue > rule.Threshold
		case "<":
			triggered = sensorValue < rule.Threshold
		case ">=":
			triggered = sensorValue >= rule.Threshold
		case "<=":
			triggered = sensorValue <= rule.Threshold
		case "==":
			triggered = sensorValue == rule.Threshold
		}

		if triggered {
			triggeredActions = append(triggeredActions, rule.Action)
			alerts = append(alerts, fmt.Sprintf("%s: %s %s %d (current: %d)",
				rule.Name, rule.Sensor, rule.Operator, rule.Threshold, sensorValue))

			log.Printf("Rule triggered: %s - %s %d %s %d",
				rule.Name, rule.Sensor, sensorValue, rule.Operator, rule.Threshold)
		}
	}

	return alerts
}

// CountRules returns the total number of rules
func (r *RuleService) CountRules() (int64, error) {
	ctx := context.Background()
	coll := r.db.GetCollection(r.collection)

	count, err := coll.Find(ctx, bson.M{}).Count()
	if err != nil {
		return 0, fmt.Errorf("failed to count rules: %w", err)
	}

	return count, nil
}

// InitializeDefaultRules creates default rules if none exist
func InitializeDefaultRules(db *database.Database) error {
	ruleService := NewRuleService(db)

	count, err := ruleService.CountRules()
	if err != nil {
		return fmt.Errorf("failed to check rule count: %w", err)
	}

	if count > 0 {
		return nil // Rules already exist
	}

	defaultRules := []models.Rule{
		{
			Name:        "Gas Danger Alert",
			Sensor:      "gas",
			Operator:    ">",
			Threshold:   700,
			Action:      "buzzer_on",
			Enabled:     true,
			Description: "Trigger buzzer when gas level exceeds danger threshold",
		},
		{
			Name:        "Rain Detection - Close Window",
			Sensor:      "water",
			Operator:    ">",
			Threshold:   800,
			Action:      "window_close",
			Enabled:     true,
			Description: "Automatically close window when rain is detected",
		},
		{
			Name:        "Low Soil Moisture Alert",
			Sensor:      "soil",
			Operator:    ">",
			Threshold:   50,
			Action:      "buzzer_on",
			Enabled:     true,
			Description: "Alert when soil moisture is too low",
		},
		{
			Name:        "Auto Light - Low Light Detection",
			Sensor:      "light",
			Operator:    "<",
			Threshold:   300,
			Action:      "white_light_on",
			Enabled:     true,
			Description: "Automatically turn on LED when light level is low",
		},
	}

	for _, rule := range defaultRules {
		if err := ruleService.CreateRule(&rule); err != nil {
			return fmt.Errorf("failed to create default rule %s: %w", rule.Name, err)
		}
	}

	log.Println("✓ Default rules initialized")
	return nil
}
