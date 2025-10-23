package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SensorData represents sensor readings from Arduino
type SensorData struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Light     int                `bson:"light" json:"light"`
	Gas       int                `bson:"gas" json:"gas"`
	Soil      int                `bson:"soil" json:"soil"`
	Water     int                `bson:"water" json:"water"`
	Infrared  int                `bson:"infrared" json:"infrared"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
	Alerts    []string           `bson:"alerts,omitempty" json:"alerts,omitempty"`
}

// SensorReading represents real-time sensor data from Arduino
type SensorReading struct {
	Gas       int       `json:"gas"`
	Light     int       `json:"light"`
	Soil      int       `json:"soil"`
	Water     int       `json:"water"`
	Infrar    int       `json:"infrar"`
	Btn1      int       `json:"btn1"`
	Btn2      int       `json:"btn2"`
	Timestamp time.Time `json:"timestamp"`
}

// ActuatorStates represents the current state of all actuators
type ActuatorStates struct {
	WhiteLight  bool `json:"white_light"`
	YellowLight bool `json:"yellow_light"`
	Relay       bool `json:"relay"`
	DoorAngle   int  `json:"door_angle"`
	WindowAngle int  `json:"window_angle"`
	Fan         bool `json:"fan"`
	FanSpeed    int  `json:"fan_speed"`
	Buzzer      bool `json:"buzzer"`
}

// Rule represents automation rules
type Rule struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	Sensor      string             `bson:"sensor" json:"sensor"`
	Operator    string             `bson:"operator" json:"operator"`
	Threshold   int                `bson:"threshold" json:"threshold"`
	Action      string             `bson:"action" json:"action"`
	Enabled     bool               `bson:"enabled" json:"enabled"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Statistics represents sensor statistics
type Statistics struct {
	LightMean float64 `json:"light_mean,omitempty"`
	LightMin  int     `json:"light_min,omitempty"`
	LightMax  int     `json:"light_max,omitempty"`
	GasMean   float64 `json:"gas_mean,omitempty"`
	GasMin    int     `json:"gas_min,omitempty"`
	GasMax    int     `json:"gas_max,omitempty"`
	SoilMean  float64 `json:"soil_mean,omitempty"`
	SoilMin   int     `json:"soil_min,omitempty"`
	SoilMax   int     `json:"soil_max,omitempty"`
	WaterMean float64 `json:"water_mean,omitempty"`
	WaterMin  int     `json:"water_min,omitempty"`
	WaterMax  int     `json:"water_max,omitempty"`
	Count     int     `json:"count"`
}

// TrendData represents hourly trend data
type TrendData struct {
	Hour  string  `bson:"_id" json:"hour"`
	Light float64 `bson:"light" json:"light"`
	Gas   float64 `bson:"gas" json:"gas"`
	Soil  float64 `bson:"soil" json:"soil"`
	Water float64 `bson:"water" json:"water"`
	Count int     `bson:"count" json:"count"`
}
