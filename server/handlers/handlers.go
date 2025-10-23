package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/caphefalumi/smart-home/server/models"
	"github.com/caphefalumi/smart-home/server/serial"
	"github.com/caphefalumi/smart-home/server/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handlers contains all HTTP request handlers
type Handlers struct {
	serialService *serial.ArduinoSerial
	sensorService *services.SensorService
	ruleService   *services.RuleService
}

// NewHandlers creates a new handlers instance
func NewHandlers(serialService *serial.ArduinoSerial, sensorService *services.SensorService, ruleService *services.RuleService) *Handlers {
	return &Handlers{
		serialService: serialService,
		sensorService: sensorService,
		ruleService:   ruleService,
	}
}

// HealthCheck returns the health status of the server
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":           "ok",
		"arduinoConnected": h.serialService.IsConnected(),
		"mongodbConnected": true, // Simplified - in production you'd check actual DB connection
	})
}

// ListSerialPorts returns available serial ports
func (h *Handlers) ListSerialPorts(c *gin.Context) {
	ports, err := h.serialService.ListPorts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "ports": []string{}})
		return
	}
	// Log the detected ports to the console
	fmt.Println("Detected serial ports:", ports)
	// Return as [{ name: "COM3" }, ...] for frontend compatibility
	var result []map[string]string
	for _, port := range ports {
		result = append(result, map[string]string{"name": port})
	}
	c.JSON(http.StatusOK, result)
}

// ConnectSerial connects to Arduino via serial port
func (h *Handlers) ConnectSerial(c *gin.Context) {
	var req struct {
		Port     string `json:"port" binding:"required"`
		BaudRate int    `json:"baudRate"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.BaudRate == 0 {
		req.BaudRate = 9600
	}

	err := h.serialService.Connect(req.Port, req.BaudRate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Connected to Arduino",
		"port":    req.Port,
	})
}

// DisconnectSerial disconnects from Arduino
func (h *Handlers) DisconnectSerial(c *gin.Context) {
	err := h.serialService.Disconnect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Disconnected from Arduino"})
}

// SendSerialCommand sends a command to Arduino
func (h *Handlers) SendSerialCommand(c *gin.Context) {
	var req struct {
		Command string `json:"command" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !h.serialService.IsConnected() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arduino not connected"})
		return
	}

	err := h.serialService.SendCommand(req.Command)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Command sent",
		"command": req.Command,
	})
}

// GetCurrentSensorData returns current sensor readings
func (h *Handlers) GetCurrentSensorData(c *gin.Context) {
	if !h.serialService.IsConnected() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arduino not connected"})
		return
	}

	data := h.serialService.GetCurrentData()
	c.JSON(http.StatusOK, data)
}

// GetActuatorStates returns current actuator states
func (h *Handlers) GetActuatorStates(c *gin.Context) {
	if !h.serialService.IsConnected() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arduino not connected"})
		return
	}

	states := h.serialService.GetActuatorStates()
	c.JSON(http.StatusOK, states)
}

// SyncActuatorState manually sets actuator state for synchronization
func (h *Handlers) SyncActuatorState(c *gin.Context) {
	if !h.serialService.IsConnected() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arduino not connected"})
		return
	}

	var req struct {
		Actuator string      `json:"actuator" binding:"required"`
		Value    interface{} `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.serialService.SetActuatorState(req.Actuator, req.Value)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Actuator state updated",
		"actuator": req.Actuator,
		"value":    req.Value,
	})
}

// GetSensorHistory returns paginated sensor history
func (h *Handlers) GetSensorHistory(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))

	var startDate, endDate *time.Time

	if startDateStr := c.Query("startDate"); startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		}
	}

	if endDateStr := c.Query("endDate"); endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		}
	}

	data, total, err := h.sensorService.GetSensorHistory(limit, skip, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  data,
		"total": total,
		"limit": limit,
		"skip":  skip,
	})
}

// GetStatistics returns sensor statistics
func (h *Handlers) GetStatistics(c *gin.Context) {
	sensor := c.Query("sensor")
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))

	// Validate sensor type if provided
	if sensor != "" {
		validSensors := []string{"light", "gas", "soil", "water"}
		valid := false
		for _, validSensor := range validSensors {
			if sensor == validSensor {
				valid = true
				break
			}
		}
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor type"})
			return
		}
	}

	stats, err := h.sensorService.GetStatistics(sensor, hours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetTrends returns sensor trends (hourly averages)
func (h *Handlers) GetTrends(c *gin.Context) {
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))

	trends, err := h.sensorService.GetTrends(hours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trends)
}

// GetRules returns all rules
func (h *Handlers) GetRules(c *gin.Context) {
	rules, err := h.ruleService.GetAllRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rules)
}

// CreateRule creates a new rule
func (h *Handlers) CreateRule(c *gin.Context) {
	var rule struct {
		Name        string `json:"name" binding:"required"`
		Sensor      string `json:"sensor" binding:"required,oneof=gas light soil water"`
		Operator    string `json:"operator" binding:"required,oneof=> < >= <= =="`
		Threshold   int    `json:"threshold" binding:"required"`
		Action      string `json:"action" binding:"required"`
		Enabled     bool   `json:"enabled"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newRule := &models.Rule{
		Name:        rule.Name,
		Sensor:      rule.Sensor,
		Operator:    rule.Operator,
		Threshold:   rule.Threshold,
		Action:      rule.Action,
		Enabled:     rule.Enabled,
		Description: rule.Description,
	}

	err := h.ruleService.CreateRule(newRule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newRule)
}

// UpdateRule updates an existing rule
func (h *Handlers) UpdateRule(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule, err := h.ruleService.UpdateRule(id, updates)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// DeleteRule deletes a rule
func (h *Handlers) DeleteRule(c *gin.Context) {
	id := c.Param("id")

	err := h.ruleService.DeleteRule(id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rule deleted"})
}

// GetAlerts returns recent alerts
func (h *Handlers) GetAlerts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	alerts, err := h.sensorService.GetAlerts(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "alerts": []interface{}{}})
		return
	}

	// Always return an array, never null
	if alerts == nil {
		alerts = []models.SensorData{}
	}
	c.JSON(http.StatusOK, alerts)
}
