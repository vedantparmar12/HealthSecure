package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"healthsecure/configs"
	"healthsecure/internal/auth"
	"healthsecure/internal/database"
	"healthsecure/internal/handlers"
	"healthsecure/internal/middleware"
	"healthsecure/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	if err := database.Initialize(config); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Start database cleanup scheduler
	database.StartCleanupScheduler()

	// Initialize services
	jwtService := auth.NewJWTService(config)
	oauthService := auth.NewOAuthService(config)
	auditService := services.NewAuditService(database.GetDB())
	userService := services.NewUserService(database.GetDB(), jwtService, auditService)
	patientService := services.NewPatientService(database.GetDB(), auditService)
	medicalRecordService := services.NewMedicalRecordService(database.GetDB(), auditService)
	emergencyService := services.NewEmergencyService(database.GetDB(), auditService, config)
	chatThreadService := services.NewChatThreadService(database.GetDB(), auditService)

	// Set Gin mode based on environment
	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	router := gin.New()

	// Apply global middleware
	router.Use(auth.LoggingMiddleware())
	router.Use(gin.Recovery())
	router.Use(auth.SecurityHeadersMiddleware())
	router.Use(auth.RequestIDMiddleware())
	router.Use(auth.CORSMiddleware(config))

	if !config.IsProduction() {
		router.Use(gin.Logger())
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		if err := database.Health(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "unhealthy",
				"database": "disconnected",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":      "healthy",
			"database":    "connected",
			"version":     "1.0.0",
			"environment": config.App.Environment,
		})
	})

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService, oauthService, jwtService)
	patientHandler := handlers.NewPatientHandler(patientService, jwtService)
	medicalRecordHandler := handlers.NewMedicalRecordHandler(medicalRecordService, jwtService)
	emergencyHandler := handlers.NewEmergencyHandler(emergencyService, jwtService)
	auditHandler := handlers.NewAuditHandler(auditService, jwtService)
	adminHandler := handlers.NewAdminHandler(userService, auditService, jwtService)
	chatHandler := handlers.NewChatHandler(userService, patientService, auditService, chatThreadService, jwtService)

	// API routes
	api := router.Group("/api")
	{
		// Authentication routes
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/signup", authHandler.Signup)
			authRoutes.POST("/refresh", authHandler.RefreshToken)
			authRoutes.POST("/logout", auth.AuthMiddleware(jwtService), authHandler.Logout)
			authRoutes.GET("/me", auth.AuthMiddleware(jwtService), authHandler.GetCurrentUser)
			
			// OAuth routes (if configured)
			if oauthService.IsConfigured() {
				authRoutes.GET("/oauth/:provider", authHandler.OAuthLogin)
				authRoutes.GET("/oauth/callback", authHandler.OAuthCallback)
			}
		}

		// Patient routes
		patients := api.Group("/patients")
		patients.Use(auth.AuthMiddleware(jwtService))
		patients.Use(auth.MedicalStaffOnly())
		{
			patients.GET("", patientHandler.GetPatients)
			patients.POST("", patientHandler.CreatePatient)
			patients.GET("/:id", patientHandler.GetPatient)
			patients.PUT("/:id", patientHandler.UpdatePatient)
			patients.DELETE("/:id", auth.AdminOnly(), patientHandler.DeletePatient)
			patients.GET("/:id/records", medicalRecordHandler.GetPatientMedicalRecords)
			patients.POST("/:id/records", auth.DoctorOnly(), medicalRecordHandler.CreateMedicalRecord)
			patients.GET("/search", patientHandler.SearchPatients)
		}

		// Medical records routes
		records := api.Group("/records")
		records.Use(auth.AuthMiddleware(jwtService))
		records.Use(auth.MedicalStaffOnly())
		{
			records.GET("/:id", medicalRecordHandler.GetMedicalRecord)
			records.PUT("/:id", auth.DoctorOnly(), medicalRecordHandler.UpdateMedicalRecord)
		}

		// Emergency access routes
		emergency := api.Group("/emergency")
		emergency.Use(auth.AuthMiddleware(jwtService))
		emergency.Use(auth.MedicalStaffOnly())
		{
			emergency.POST("/request", emergencyHandler.RequestEmergencyAccess)
			emergency.POST("/activate/:id", emergencyHandler.ActivateEmergencyAccess)
			emergency.POST("/revoke/:id", emergencyHandler.RevokeEmergencyAccess)
			emergency.GET("/active", auth.AdminOnly(), emergencyHandler.GetActiveEmergencyAccess)
			emergency.GET("/user/:id", emergencyHandler.GetUserEmergencyAccess)
			emergency.GET("/patient/:id", auth.AdminOnly(), emergencyHandler.GetPatientEmergencyAccess)
		}

		// Audit routes
		audit := api.Group("/audit")
		audit.Use(auth.AuthMiddleware(jwtService))
		{
			audit.GET("/logs", auditHandler.GetAuditLogs)
			audit.GET("/users/:id", auditHandler.GetUserAuditHistory)
			audit.GET("/patients/:id", auth.MedicalStaffOnly(), auditHandler.GetPatientAuditHistory)
			audit.GET("//security-events", auth.AdminOnly(), auditHandler.GetSecurityEvents)
			audit.POST("/security-events/:id/resolve", auth.AdminOnly(), auditHandler.ResolveSecurityEvent)
			audit.GET("/statistics", auth.AdminOnly(), auditHandler.GetAuditStatistics)
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(auth.AuthMiddleware(jwtService))
		admin.Use(auth.AdminOnly())
		{
			admin.GET("/users", adminHandler.GetAllUsers)
			admin.POST("/users", adminHandler.CreateUser)
			admin.GET("/users/:id", adminHandler.GetUser)
			admin.PUT("/users/:id", adminHandler.UpdateUser)
			admin.POST("/users/:id/deactivate", adminHandler.DeactivateUser)
			admin.GET("/users/:id/sessions", adminHandler.GetUserSessions)
			admin.GET("/dashboard/stats", adminHandler.GetDashboardStats)
		}

		// User profile routes
		profile := api.Group("/profile")
		profile.Use(auth.AuthMiddleware(jwtService))
		{
			profile.GET("", authHandler.GetCurrentUser)
			profile.PUT("", authHandler.UpdateProfile)
			profile.POST("/change-password", authHandler.ChangePassword)
			profile.GET("/sessions", authHandler.GetUserSessions)
		}

		// Chat routes
		chat := api.Group("/chat")
		chat.Use(auth.AuthMiddleware(jwtService))
		chat.Use(middleware.RateLimiter(20, time.Minute))
		{
			chat.POST("", chatHandler.ProcessChatMessage)
			chat.POST("/threads", chatHandler.CreateThread)
			chat.GET("/threads", chatHandler.GetThreads)
			chat.GET("/threads/:thread_id/messages", chatHandler.GetThreadMessages)
			chat.POST("/feedback", chatHandler.SubmitFeedback)
			chat.GET("/feedback/stats", chatHandler.GetFeedbackStats)
		}

		// AI routes
	ai := api.Group("/ai")
	ai.Use(middleware.AIAuthMiddleware(config.AI.AuthToken))
	{
		aiHandler := handlers.NewAIHandler(patientService)
		ai.GET("/patients", aiHandler.GetPatients)
	}
	}

	// Metrics endpoint (if monitoring is enabled)
	if config.Monitoring.Enabled {
		router.GET("/metrics", func(c *gin.Context) {
			// This would typically serve Prometheus metrics
			c.String(http.StatusOK, "# Metrics endpoint placeholder\n")
		})
	}

	// Start server
	serverAddr := fmt.Sprintf(":%d", config.App.ServerPort)
	
	log.Printf("Starting HealthSecure server on port %d", config.App.ServerPort)
	log.Printf("Environment: %s", config.App.Environment)
	log.Printf("Database: %s@%s:%d/%s", config.Database.User, config.Database.Host, config.Database.Port, config.Database.Name)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		
		log.Println("Shutting down server...")
		database.Close()
		os.Exit(0)
	}()

	// Start server with or without TLS
	if config.RequiresSSL() {
		log.Printf("Starting HTTPS server with TLS")
		if err := router.RunTLS(serverAddr, config.SSL.CertPath, config.SSL.KeyPath); err != nil {
			log.Fatalf("Failed to start HTTPS server: %v", err)
		}
	} else {
		if config.IsProduction() {
			log.Println("WARNING: Running in production without TLS")
		}
		if err := router.Run(serverAddr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
}