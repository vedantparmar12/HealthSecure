package database

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"healthsecure/configs"
	"healthsecure/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

var DB *gorm.DB

// Initialize establishes database connection and runs migrations
func Initialize(config *configs.Config) error {
	var err error
	
	// Configure GORM logger based on environment
	var gormConfig *gorm.Config
	if config.IsProduction() {
		gormConfig = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		}
	} else {
		gormConfig = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Error),
		}
	}

	// Use MySQL for Railway or remote hosts, SQLite for local development
	if config.Database.Host != "" && config.Database.Host != "localhost" && config.Database.Host != "127.0.0.1" {
		DB, err = gorm.Open(mysql.Open(config.GetDatabaseDSN()), gormConfig)
		if err != nil {
			log.Printf("MySQL connection failed, falling back to SQLite: %v", err)
			// Create data directory if it doesn't exist
			if err := os.MkdirAll("data", 0755); err != nil {
				return fmt.Errorf("failed to create data directory: %w", err)
			}
			DB, err = gorm.Open(sqlite.Dialector{
			DriverName: "sqlite",
			DSN:        "data/healthsecure.db",
		}, gormConfig)
			if err != nil {
				return fmt.Errorf("failed to connect to SQLite database: %w", err)
			}
			log.Println("Using SQLite database for development")
		}
	} else {
		// Default to SQLite for development
		if err := os.MkdirAll("data", 0755); err != nil {
			return fmt.Errorf("failed to create data directory: %w", err)
		}
		DB, err = gorm.Open(sqlite.Dialector{
			DriverName: "sqlite",
			DSN:        "data/healthsecure.db",
		}, gormConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to SQLite database: %w", err)
		}
		log.Println("Using SQLite database for development")
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established successfully")

	// Run auto migrations
	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Seed demo users for development
	if !config.IsProduction() {
		if err := seedDemoUsers(); err != nil {
			log.Printf("Failed to seed demo users: %v", err)
		}
	}

	return nil
}

// runMigrations performs automatic schema migrations
func runMigrations() error {
	log.Println("Running database migrations...")

	// Define models to migrate in dependency order
	modelsToMigrate := []interface{}{
		&models.User{},
		&models.Patient{},
		&models.MedicalRecord{},
		&models.AuditLog{},
		&models.EmergencyAccess{},
		&models.ChatThread{},
		&models.ChatMessage{},
		&BlacklistedToken{},
		&UserSession{},
		&SystemSetting{},
		&SecurityEvent{},
	}

	// Run migrations with error handling for duplicate indexes
	for _, model := range modelsToMigrate {
		if err := DB.AutoMigrate(model); err != nil {
			isDuplicateKeyError := false
			errStr := err.Error()

			// Check for database-specific duplicate key errors
			switch DB.Dialector.Name() {
			case "mysql":
				if strings.Contains(errStr, "Error 1062:") {
					isDuplicateKeyError = true
				}
			case "sqlite":
				if strings.Contains(errStr, "UNIQUE constraint failed") {
					isDuplicateKeyError = true
				}
			default:
				// Generic check for other databases
				if strings.Contains(errStr, "Duplicate key name") || strings.Contains(errStr, "already exists") {
					isDuplicateKeyError = true
				}
			}

			if isDuplicateKeyError {
				log.Printf("Warning: Index or column already exists for %T, skipping: %v", model, err)
				continue
			}
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// Additional models for system functionality
type BlacklistedToken struct {
	ID        uint      `json:"id" gorm:"primaryKey;type:bigint unsigned;autoIncrement"`
	TokenHash string    `json:"token_hash" gorm:"unique;not null"`
	UserID    uint      `json:"user_id" gorm:"not null;index;type:bigint unsigned"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null;index"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	User models.User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (bt *BlacklistedToken) TableName() string {
	return "blacklisted_tokens"
}

type UserSession struct {
	ID           uint      `json:"id" gorm:"primaryKey;type:bigint unsigned;autoIncrement"`
	UserID       uint      `json:"user_id" gorm:"not null;index;type:bigint unsigned"`
	SessionID    string    `json:"session_id" gorm:"unique;not null"`
	IPAddress    string    `json:"ip_address" gorm:"not null"`
	UserAgent    string    `json:"user_agent" gorm:"type:text"`
	ExpiresAt    time.Time `json:"expires_at" gorm:"not null;index"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	LastActivity time.Time `json:"last_activity" gorm:"autoUpdateTime"`

	User models.User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (us *UserSession) TableName() string {
	return "user_sessions"
}

type SystemSetting struct {
	ID           uint      `json:"id" gorm:"primaryKey;type:bigint unsigned;autoIncrement"`
	SettingKey   string    `json:"setting_key" gorm:"unique;not null"`
	SettingValue string    `json:"setting_value" gorm:"type:text"`
	Description  string    `json:"description" gorm:"type:text"`
	UpdatedBy    *uint     `json:"updated_by" gorm:"type:bigint unsigned"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	UpdatedByUser *models.User `json:"updated_by_user,omitempty" gorm:"foreignKey:UpdatedBy"`
}

func (ss *SystemSetting) TableName() string {
	return "system_settings"
}

type SecurityEventType string

const (
	SecurityEventFailedLogin        SecurityEventType = "FAILED_LOGIN"
	SecurityEventSuspiciousActivity SecurityEventType = "SUSPICIOUS_ACTIVITY"
	SecurityEventUnauthorizedAccess SecurityEventType = "UNAUTHORIZED_ACCESS"
	SecurityEventEmergencyAccess    SecurityEventType = "EMERGENCY_ACCESS"
	SecurityEventDataBreach         SecurityEventType = "DATA_BREACH"
	SecurityEventSystemAlert        SecurityEventType = "SYSTEM_ALERT"
)

type SecurityEventSeverity string

const (
	SecuritySeverityLow      SecurityEventSeverity = "LOW"
	SecuritySeverityMedium   SecurityEventSeverity = "MEDIUM"
	SecuritySeverityHigh     SecurityEventSeverity = "HIGH"
	SecuritySeverityCritical SecurityEventSeverity = "CRITICAL"
)

type SecurityEvent struct {
	ID          uint                  `json:"id" gorm:"primaryKey;type:bigint unsigned;autoIncrement"`
	EventType   SecurityEventType     `json:"event_type" gorm:"not null;index"`
	Severity    SecurityEventSeverity `json:"severity" gorm:"default:'MEDIUM';index"`
	UserID      *uint                 `json:"user_id" gorm:"index;type:bigint unsigned"`
	IPAddress   string                `json:"ip_address" gorm:"size:45"`
	Description string                `json:"description" gorm:"type:text;not null"`
	Details     string                `json:"details" gorm:"type:json"`
	Resolved    bool                  `json:"resolved" gorm:"default:false;index"`
	ResolvedBy  *uint                 `json:"resolved_by" gorm:"type:bigint unsigned"`
	ResolvedAt  *time.Time            `json:"resolved_at"`
	CreatedAt   time.Time             `json:"created_at" gorm:"autoCreateTime;index"`

	User         *models.User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ResolvedByUser *models.User `json:"resolved_by_user,omitempty" gorm:"foreignKey:ResolvedBy"`
}

func (se *SecurityEvent) TableName() string {
	return "security_events"
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// Health checks database connectivity
func Health() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// CleanupExpiredTokens removes expired blacklisted tokens
func CleanupExpiredTokens() error {
	result := DB.Where("expires_at < ?", time.Now()).Delete(&BlacklistedToken{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", result.Error)
	}

	log.Printf("Cleaned up %d expired tokens", result.RowsAffected)
	return nil
}

// CleanupExpiredSessions removes expired user sessions
func CleanupExpiredSessions() error {
	result := DB.Where("expires_at < ?", time.Now()).Delete(&UserSession{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", result.Error)
	}

	log.Printf("Cleaned up %d expired sessions", result.RowsAffected)
	return nil
}

// UpdateEmergencyAccessStatus updates expired emergency access records
func UpdateEmergencyAccessStatus() error {
	result := DB.Model(&models.EmergencyAccess{}).
		Where("expires_at < ? AND status NOT IN ?", time.Now(), []string{"expired", "revoked"}).
		Update("status", "expired")

	if result.Error != nil {
		return fmt.Errorf("failed to update emergency access status: %w", result.Error)
	}

	log.Printf("Updated %d expired emergency access records", result.RowsAffected)
	return nil
}

// RunCleanupTasks performs routine database cleanup
func RunCleanupTasks() error {
	log.Println("Running database cleanup tasks...")

	tasks := []func() error{
		CleanupExpiredTokens,
		CleanupExpiredSessions,
		UpdateEmergencyAccessStatus,
	}

	for _, task := range tasks {
		if err := task(); err != nil {
			log.Printf("Cleanup task failed: %v", err)
			// Continue with other cleanup tasks
		}
	}

	log.Println("Database cleanup tasks completed")
	return nil
}

// GetSystemSetting retrieves a system setting value
func GetSystemSetting(key string) (string, error) {
	var setting SystemSetting
	if err := DB.Where("setting_key = ?", key).First(&setting).Error; err != nil {
		return "", err
	}
	return setting.SettingValue, nil
}

// SetSystemSetting updates or creates a system setting
func SetSystemSetting(key, value, description string, updatedBy uint) error {
	setting := SystemSetting{
		SettingKey:   key,
		SettingValue: value,
		Description:  description,
		UpdatedBy:    &updatedBy,
	}

	return DB.Save(&setting).Error
}

// StartCleanupScheduler runs cleanup tasks periodically
func StartCleanupScheduler() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // Run every hour
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				RunCleanupTasks()
			}
		}
	}()

	log.Println("Database cleanup scheduler started")
}

// seedDemoUsers creates demo users for development if they don't exist
func seedDemoUsers() error {
	log.Println("Seeding demo users for development...")

	demoUsers := []models.User{
		{
			Name:     "Dr. John Smith",
			Email:    "dr.smith@hospital.local",
			Password: "$2a$10$b6v3JhBelIMbx4fWAf2aYurrPP7HHNQ63ZL5VZJNWlN/tfQQzhhu.", // "Doctor123"
			Role:     models.RoleDoctor,
			Active:   true,
		},
		{
			Name:     "Nurse Jane Wilson",
			Email:    "nurse.jane@hospital.local",
			Password: "$2a$10$Y2jnrSx4oi2byIQZEUBov.w4S.EP24vTUlmjiM7n3vORfq.Cs.eF.", // "Nurse123"
			Role:     models.RoleNurse,
			Active:   true,
		},
		{
			Name:     "Admin User",
			Email:    "admin@hospital.local",
			Password: "$2a$10$8.OLHoNHEF6KV2wWWWcgzuPqC1EphG8z7J8.7ySFqHJNhpDqDkBFS", // "admin123"
			Role:     models.RoleAdmin,
			Active:   true,
		},
	}

	for _, user := range demoUsers {
		var existingUser models.User
		if err := DB.Where("email = ?", user.Email).First(&existingUser).Error; err != nil {
			// User doesn't exist, create it
			if err := DB.Create(&user).Error; err != nil {
				return fmt.Errorf("failed to create demo user %s: %w", user.Email, err)
			}
			log.Printf("Created demo user: %s (%s)", user.Name, user.Email)
		} else {
			// User exists, update password if it's different
			if existingUser.Password != user.Password {
				existingUser.Password = user.Password
				if err := DB.Save(&existingUser).Error; err != nil {
					return fmt.Errorf("failed to update demo user password %s: %w", user.Email, err)
				}
				log.Printf("Updated demo user password: %s (%s)", user.Name, user.Email)
			}
		}
	}

	log.Println("Demo users seeded successfully")
	return nil
}