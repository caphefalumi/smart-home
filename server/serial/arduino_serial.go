package serial

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/caphefalumi/smart-home/server/models"
	"github.com/caphefalumi/smart-home/server/services"
	"github.com/tarm/serial"
)

// ArduinoSerial handles serial communication with Arduino
type ArduinoSerial struct {
	port           io.ReadWriteCloser
	isConnected    bool
	currentData    *models.SensorReading
	actuatorStates *models.ActuatorStates
	dataBuffer     []models.SensorData
	sensorService  *services.SensorService
	ruleService    *services.RuleService
	commandQueue   []string
	isReady        bool
	mutex          sync.RWMutex
	stopChan       chan bool
	saveInterval   time.Duration
}

// NewArduinoSerial creates a new Arduino serial handler
func NewArduinoSerial(sensorService *services.SensorService, ruleService *services.RuleService) *ArduinoSerial {
	return &ArduinoSerial{
		actuatorStates: &models.ActuatorStates{},
		dataBuffer:     make([]models.SensorData, 0),
		sensorService:  sensorService,
		ruleService:    ruleService,
		commandQueue:   make([]string, 0),
		isReady:        true,
		stopChan:       make(chan bool),
		saveInterval:   2 * time.Second,
	}
}

// ListPorts returns available serial ports
func (a *ArduinoSerial) ListPorts() ([]string, error) {
	// On Windows, common ports are COM1, COM2, etc.
	// This is a simplified implementation - in production you'd use a proper port discovery
	commonPorts := []string{"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "COM10"}

	var availablePorts []string
	for _, portName := range commonPorts {
		config := &serial.Config{
			Name:        portName,
			Baud:        9600,
			ReadTimeout: time.Millisecond * 100,
		}

		if port, err := serial.OpenPort(config); err == nil {
			port.Close()
			availablePorts = append(availablePorts, portName)
		}
	}

	return availablePorts, nil
}

// Connect establishes connection to Arduino
func (a *ArduinoSerial) Connect(portName string, baudRate int) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.isConnected {
		return fmt.Errorf("already connected to a port")
	}

	config := &serial.Config{
		Name:        portName,
		Baud:        baudRate,
		ReadTimeout: time.Second * 5,
	}

	port, err := serial.OpenPort(config)
	if err != nil {
		return fmt.Errorf("failed to open port %s: %w", portName, err)
	}

	a.port = port
	a.isConnected = true

	// Wait for Arduino to reset
	time.Sleep(2 * time.Second)

	// Start data processing goroutines
	go a.startDataListener()
	go a.startDataSaver()

	log.Printf("✓ Connected to Arduino on %s", portName)
	return nil
}

// Disconnect closes the serial connection
func (a *ArduinoSerial) Disconnect() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.isConnected {
		return nil
	}

	// Stop background processes
	close(a.stopChan)

	// Save any remaining buffered data
	a.saveBufferedData()

	if a.port != nil {
		a.port.Close()
		a.port = nil
	}

	a.isConnected = false
	log.Println("✓ Disconnected from Arduino")
	return nil
}

// IsConnected returns connection status
func (a *ArduinoSerial) IsConnected() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.isConnected
}

// SendCommand sends a command to Arduino
func (a *ArduinoSerial) SendCommand(command string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.isConnected || a.port == nil {
		return fmt.Errorf("not connected to Arduino")
	}

	if !a.isReady {
		// Queue the command
		a.commandQueue = append(a.commandQueue, command)
		return nil
	}

	// Send command immediately
	_, err := a.port.Write([]byte(command + "\n"))
	if err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	a.isReady = false
	log.Printf("→ %s", command)
	return nil
}

// GetCurrentData returns the latest sensor reading
func (a *ArduinoSerial) GetCurrentData() *models.SensorReading {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.currentData
}

// GetActuatorStates returns current actuator states
func (a *ArduinoSerial) GetActuatorStates() *models.ActuatorStates {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.actuatorStates
}

// SetActuatorState manually sets actuator state for synchronization
func (a *ArduinoSerial) SetActuatorState(actuator string, value interface{}) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	switch actuator {
	case "white_light":
		if val, ok := value.(bool); ok {
			a.actuatorStates.WhiteLight = val
		}
	case "yellow_light":
		if val, ok := value.(bool); ok {
			a.actuatorStates.YellowLight = val
		}
	case "relay":
		if val, ok := value.(bool); ok {
			a.actuatorStates.Relay = val
		}
	case "fan":
		if val, ok := value.(bool); ok {
			a.actuatorStates.Fan = val
		}
	case "buzzer":
		if val, ok := value.(bool); ok {
			a.actuatorStates.Buzzer = val
		}
	case "door_angle":
		if val, ok := value.(float64); ok {
			a.actuatorStates.DoorAngle = int(val)
		}
	case "window_angle":
		if val, ok := value.(float64); ok {
			a.actuatorStates.WindowAngle = int(val)
		}
	case "fan_speed":
		if val, ok := value.(float64); ok {
			a.actuatorStates.FanSpeed = int(val)
		}
	}
}

// startDataListener listens for incoming data from Arduino
func (a *ArduinoSerial) startDataListener() {
	scanner := bufio.NewScanner(a.port)

	for scanner.Scan() {
		select {
		case <-a.stopChan:
			return
		default:
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				log.Printf("← %s", line)
				a.handleIncomingData(line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Serial scanner error: %v", err)
	}
}

// startDataSaver periodically saves buffered data
func (a *ArduinoSerial) startDataSaver() {
	ticker := time.NewTicker(a.saveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopChan:
			return
		case <-ticker.C:
			a.saveBufferedData()
		}
	}
}

// handleIncomingData processes incoming serial data
func (a *ArduinoSerial) handleIncomingData(line string) {
	// Handle ACK
	if line == "ACK" {
		a.mutex.Lock()
		a.isReady = true

		// Send next queued command
		if len(a.commandQueue) > 0 {
			nextCommand := a.commandQueue[0]
			a.commandQueue = a.commandQueue[1:]
			a.mutex.Unlock()

			// Send with slight delay
			time.AfterFunc(100*time.Millisecond, func() {
				a.SendCommand(nextCommand)
			})
		} else {
			a.mutex.Unlock()
		}
		return
	}

	// Parse sensor data
	if strings.Contains(line, "GAS:") {
		a.parseSensorData(line)
	} else {
		// Parse actuator responses
		a.parseActuatorResponse(line)
	}
}

// parseSensorData parses sensor data from Arduino
func (a *ArduinoSerial) parseSensorData(line string) {
	// Format: "GAS:123,LIGHT:456,SOIL:789,WATER:101,INFRAR:1,BTN1:0,BTN2:1"
	data := &models.SensorReading{
		Timestamp: time.Now(),
	}

	parts := strings.Split(line, ",")
	for _, part := range parts {
		keyValue := strings.Split(part, ":")
		if len(keyValue) != 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(keyValue[0]))
		valueStr := strings.TrimSpace(keyValue[1])
		value, err := strconv.Atoi(valueStr)
		if err != nil {
			continue
		}

		switch key {
		case "gas":
			data.Gas = value
		case "light":
			data.Light = value
		case "soil":
			data.Soil = value
		case "water":
			data.Water = value
		case "infrar":
			data.Infrar = value
		case "btn1":
			data.Btn1 = value
		case "btn2":
			data.Btn2 = value
		}
	}

	// Validate that we have the main sensor readings
	if data.Gas == 0 && data.Light == 0 && data.Soil == 0 && data.Water == 0 {
		return
	}

	a.mutex.Lock()
	a.currentData = data

	// Add to buffer for saving
	sensorData := models.SensorData{
		Light:     data.Light,
		Gas:       data.Gas,
		Soil:      data.Soil,
		Water:     data.Water,
		Infrared:  data.Infrar,
		Timestamp: data.Timestamp,
	}

	// Evaluate rules and get alerts
	alerts := a.ruleService.EvaluateRules(data)
	if len(alerts) > 0 {
		sensorData.Alerts = alerts

		// Execute triggered actions
		a.executeTriggeredActions(alerts)
	}

	a.dataBuffer = append(a.dataBuffer, sensorData)

	// Keep buffer size manageable
	if len(a.dataBuffer) > 100 {
		a.dataBuffer = a.dataBuffer[len(a.dataBuffer)-50:]
	}

	a.mutex.Unlock()
}

// parseActuatorResponse parses actuator state changes from Arduino
func (a *ArduinoSerial) parseActuatorResponse(line string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	line = strings.ToLower(line)

	// Parse various actuator responses
	if strings.Contains(line, "light on") {
		if strings.Contains(line, "white") {
			a.actuatorStates.WhiteLight = true
		} else if strings.Contains(line, "yellow") {
			a.actuatorStates.YellowLight = true
		}
	} else if strings.Contains(line, "light off") {
		if strings.Contains(line, "white") {
			a.actuatorStates.WhiteLight = false
		} else if strings.Contains(line, "yellow") {
			a.actuatorStates.YellowLight = false
		}
	} else if strings.Contains(line, "fan on") {
		a.actuatorStates.Fan = true
	} else if strings.Contains(line, "fan off") {
		a.actuatorStates.Fan = false
	} else if strings.Contains(line, "relay on") {
		a.actuatorStates.Relay = true
	} else if strings.Contains(line, "relay off") {
		a.actuatorStates.Relay = false
	} else if strings.Contains(line, "buzzer on") {
		a.actuatorStates.Buzzer = true
	} else if strings.Contains(line, "buzzer off") {
		a.actuatorStates.Buzzer = false
	}

	// Parse servo positions
	if matches := regexp.MustCompile(`door.*?(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
		if angle, err := strconv.Atoi(matches[1]); err == nil {
			a.actuatorStates.DoorAngle = angle
		}
	}
	if matches := regexp.MustCompile(`window.*?(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
		if angle, err := strconv.Atoi(matches[1]); err == nil {
			a.actuatorStates.WindowAngle = angle
		}
	}
}

// executeTriggeredActions executes actions based on triggered rules
func (a *ArduinoSerial) executeTriggeredActions(alerts []string) {
	// Extract actions from alerts and execute them
	for _, alert := range alerts {
		// Simple action extraction - in production, you'd parse this more robustly
		if strings.Contains(alert, "buzzer_on") {
			a.SendCommand("BUZZER_ON")
		} else if strings.Contains(alert, "white_light_on") {
			a.SendCommand("WHITE_LIGHT_ON")
		} else if strings.Contains(alert, "window_close") {
			a.SendCommand("WINDOW_CLOSE")
		}
	}
}

// saveBufferedData saves buffered sensor data to database
func (a *ArduinoSerial) saveBufferedData() {
	a.mutex.Lock()
	if len(a.dataBuffer) == 0 {
		a.mutex.Unlock()
		return
	}

	// Copy buffer and clear it
	dataToSave := make([]models.SensorData, len(a.dataBuffer))
	copy(dataToSave, a.dataBuffer)
	a.dataBuffer = a.dataBuffer[:0]
	a.mutex.Unlock()

	// Save to database
	if err := a.sensorService.SaveBulkSensorData(dataToSave); err != nil {
		log.Printf("Error saving sensor data: %v", err)

		// Put data back in buffer if save failed
		a.mutex.Lock()
		a.dataBuffer = append(dataToSave, a.dataBuffer...)
		a.mutex.Unlock()
	}
}
