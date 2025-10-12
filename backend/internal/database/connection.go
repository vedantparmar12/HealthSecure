package database

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"healthsecure/configs"
	"healthsecure/internal/models"

	"github.com/go-sql-driver/mysql"
	gorm_mysql "gorm.io/driver/mysql"
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
		// Register custom TLS config for Railway MySQL
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // For Railway, skip certificate verification
		}
		if err := mysql.RegisterTLSConfig("railway", tlsConfig); err != nil {
			log.Printf("Failed to register TLS config: %v", err)
		}
		
		// Update DSN to use custom TLS config
		dsn := config.GetDatabaseDSN()
		if config.Database.TLSMode != "" && config.Database.TLSMode != "preferred" {
			// Replace the tls parameter with our custom config
			dsn = strings.Replace(dsn, "&tls="+config.Database.TLSMode, "&tls=railway", 1)
		} else {
			// Add TLS if not present
			dsn += "&tls=railway"
		}
		
		DB, err = gorm.Open(gorm_mysql.Open(dsn), gormConfig)
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

	// Run auto migrations (can be disabled with SKIP_MIGRATIONS=true)
	if skipMigrations := os.Getenv("SKIP_MIGRATIONS"); skipMigrations != "true" {
		if err := runMigrations(); err != nil {
			log.Printf("Migration failed: %v", err)
			log.Println("You can skip migrations by setting SKIP_MIGRATIONS=true environment variable")
			return fmt.Errorf("failed to run migrations: %w", err)
		}
	} else {
		log.Println("Skipping database migrations (SKIP_MIGRATIONS=true)")
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

	// First, handle existing tables with enum columns
	if err := handleExistingEnumColumns(); err != nil {
		log.Printf("Warning: Could not handle existing enum columns: %v", err)
	}

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

	// Run migrations with enhanced error handling
	for _, model := range modelsToMigrate {
		if err := DB.AutoMigrate(model); err != nil {
			errStr := err.Error()

			// Check for various error types and handle gracefully
			if isIgnorableError(errStr, DB.Dialector.Name()) {
				log.Printf("Warning: Ignorable migration error for %T: %v", model, err)
				continue
			}

			// For enum-related errors, try alternative approach
			if strings.Contains(errStr, "enum") || strings.Contains(errStr, "syntax error") {
				log.Printf("Enum-related error for %T, attempting manual migration...", model)
				if err := handleEnumMigration(model); err != nil {
					log.Printf("Manual migration also failed for %T: %v", model, err)
					continue
				}
				continue
			}

			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// handleExistingEnumColumns modifies existing enum columns to varchar
func handleExistingEnumColumns() error {
	dialectName := DB.Dialector.Name()

	if dialectName == "mysql" {
		// Check if users table exists and has enum role column
		if DB.Migrator().HasTable("users") && DB.Migrator().HasColumn(&models.User{}, "role") {
			// Try to alter the column if it's currently an enum
			if err := DB.Exec("ALTER TABLE users MODIFY COLUMN role VARCHAR(20)").Error; err != nil {
				log.Printf("Could not alter users.role column: %v", err)
			}
		}

		// Check medical_records table
		if DB.Migrator().HasTable("medical_records") && DB.Migrator().HasColumn(&models.MedicalRecord{}, "severity") {
			if err := DB.Exec("ALTER TABLE medical_records MODIFY COLUMN severity VARCHAR(20)").Error; err != nil {
				log.Printf("Could not alter medical_records.severity column: %v", err)
			}
		}
	} else if dialectName == "sqlite" {
		// SQLite doesn't support ALTER COLUMN, so we need to recreate tables if they have enum issues
		// For now, just log and continue
		log.Println("SQLite detected - enum columns will be handled during AutoMigrate")
	}

	return nil
}

// isIgnorableError checks if a migration error can be safely ignored
func isIgnorableError(errStr, dialectName string) bool {
	ignorablePatterns := []string{
		"Error 1062:", // MySQL duplicate key
		"UNIQUE constraint failed", // SQLite unique constraint
		"Duplicate key name", // Generic duplicate key
		"already exists", // Table/column exists
		"duplicate column name", // Column already exists
	}

	for _, pattern := range ignorablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// handleEnumMigration handles enum-related migration issues
func handleEnumMigration(model interface{}) error {
	// Try to create the table without constraints first
	switch model.(type) {
	case *models.User:
		return DB.Exec(`CREATE TABLE IF NOT EXISTS users (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			email VARCHAR(191) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(20) NOT NULL,
			name VARCHAR(100) NOT NULL,
			active BOOLEAN DEFAULT TRUE,
			last_login DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			INDEX idx_users_email (email)
		)`).Error
	case *models.MedicalRecord:
		return DB.Exec(`CREATE TABLE IF NOT EXISTS medical_records (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			patient_id BIGINT UNSIGNED NOT NULL,
			doctor_id BIGINT UNSIGNED NOT NULL,
			diagnosis TEXT,
			treatment TEXT,
			notes TEXT,
			medications TEXT,
			severity VARCHAR(20),
			created_at DATETIME,
			updated_at DATETIME,
			INDEX idx_medical_records_patient_id (patient_id),
			INDEX idx_medical_records_doctor_id (doctor_id)
		)`).Error
	}

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