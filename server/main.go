package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caphefalumi/smart-home/config"
	"github.com/caphefalumi/smart-home/database"
	"github.com/caphefalumi/smart-home/handlers"
	"github.com/caphefalumi/smart-home/serial"
	"github.com/caphefalumi/smart-home/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Connect(cfg.MongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer db.Close(context.Background())

	// Initialize default rules
	if err := services.InitializeDefaultRules(db); err != nil {
		log.Printf("Failed to initialize default rules: %v", err)
	}

	// Initialize services
	sensorService := services.NewSensorService(db)
	ruleService := services.NewRuleService(db)
	serialService := serial.NewArduinoSerial(sensorService, ruleService)

	// Initialize handlers
	h := handlers.NewHandlers(serialService, sensorService, ruleService)

	// Setup Gin router
	r := setupRouter(h)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		fmt.Printf("🚀 Edge server running on http://localhost:%s\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("\n🛑 Shutting down gracefully...")

	// Shutdown server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Disconnect serial
	if err := serialService.Disconnect(); err != nil {
		log.Printf("Error disconnecting serial: %v", err)
	}

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exited")
}

func setupRouter(h *handlers.Handlers) *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API routes
	api := r.Group("/api")
	{
		// Health check
		api.GET("/health", h.HealthCheck)

		// Serial port endpoints
		serial := api.Group("/serial")
		{
			serial.GET("/ports", h.ListSerialPorts)
			serial.POST("/connect", h.ConnectSerial)
			serial.POST("/disconnect", h.DisconnectSerial)
			serial.POST("/command", h.SendSerialCommand)
		}

		// Sensor endpoints
		sensors := api.Group("/sensors")
		{
			sensors.GET("/current", h.GetCurrentSensorData)
			sensors.GET("/history", h.GetSensorHistory)
		}

		// Actuator endpoints
		actuators := api.Group("/actuators")
		{
			actuators.GET("/states", h.GetActuatorStates)
			actuators.POST("/sync", h.SyncActuatorState)
		}

		// Analytics endpoints
		analytics := api.Group("/analytics")
		{
			analytics.GET("/statistics", h.GetStatistics)
			analytics.GET("/trends", h.GetTrends)
		}

		// Rules endpoints
		rules := api.Group("/rules")
		{
			rules.GET("", h.GetRules)
			rules.POST("", h.CreateRule)
			rules.PUT("/:id", h.UpdateRule)
			rules.DELETE("/:id", h.DeleteRule)
		}

		// Alerts endpoint
		api.GET("/alerts", h.GetAlerts)
	}

	return r
}
