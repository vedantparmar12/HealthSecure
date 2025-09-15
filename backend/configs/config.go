package configs

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	// Database configuration
	Database DatabaseConfig `mapstructure:"database"`
	
	// Redis configuration
	Redis RedisConfig `mapstructure:"redis"`
	
	// JWT configuration
	JWT JWTConfig `mapstructure:"jwt"`
	
	// OAuth2 configuration
	OAuth OAuth2Config `mapstructure:"oauth"`
	
	// Security configuration
	Security SecurityConfig `mapstructure:"security"`
	
	// Emergency access configuration
	Emergency EmergencyConfig `mapstructure:"emergency"`
	
	// Application configuration
	App AppConfig `mapstructure:"app"`
	
	// SSL configuration
	SSL SSLConfig `mapstructure:"ssl"`
	
	// Monitoring configuration
	Monitoring MonitoringConfig `mapstructure:"monitoring"`

	// AI configuration
	AI AIConfig `mapstructure:"ai"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	TLSMode  string `mapstructure:"tls_mode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret             string        `mapstructure:"secret"`
	Expires            time.Duration `mapstructure:"expires"`
	RefreshTokenExpires time.Duration `mapstructure:"refresh_token_expires"`
}

type OAuth2Config struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

type SecurityConfig struct {
	BCryptCost         int           `mapstructure:"bcrypt_cost"`
	RateLimitRequests  int           `mapstructure:"rate_limit_requests"`
	RateLimitWindow    time.Duration `mapstructure:"rate_limit_window"`
}

type EmergencyConfig struct {
	AccessDuration      time.Duration `mapstructure:"access_duration"`
	NotificationEmail   string        `mapstructure:"notification_email"`
}

type AppConfig struct {
	ServerPort  int      `mapstructure:"server_port"`
	Environment string   `mapstructure:"environment"`
	LogLevel    string   `mapstructure:"log_level"`
	EnableCORS  bool     `mapstructure:"enable_cors"`
	CORSOrigins []string `mapstructure:"cors_origins"`
}

type SSLConfig struct {
	Enabled  bool
	CertPath string
	KeyPath  string
}

// AIConfig holds configuration for the AI service
type AIConfig struct {
	AuthToken string
}

type MonitoringConfig struct {
	Enabled     bool `mapstructure:"enabled"`
	MetricsPort int  `mapstructure:"metrics_port"`
}

var GlobalConfig *Config

func LoadConfig() (*Config, error) {
	config := &Config{}

	// Set default values
	setDefaults()

	// Load from .env file if it exists
	viper.SetConfigFile("../configs/.env")
	viper.SetConfigType("env")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Read environment variables
	viper.AutomaticEnv()

	// Set config values from environment variables
	config.Database = DatabaseConfig{
		Host:     getViperString("DB_HOST", "localhost"),
		Port:     getViperInt("DB_PORT", 3306),
		Name:     getViperString("DB_NAME", "healthsecure"),
		User:     getViperString("DB_USER", "root"),
		Password: getViperString("DB_PASSWORD", ""),
		TLSMode:  getViperString("DB_TLS_MODE", "preferred"),
	}

	config.Redis = RedisConfig{
		Host:     getViperString("REDIS_HOST", "localhost"),
		Port:     getViperInt("REDIS_PORT", 6379),
		Password: getViperString("REDIS_PASSWORD", ""),
		DB:       getViperInt("REDIS_DB", 0),
	}

	config.JWT = JWTConfig{
		Secret:              getViperString("JWT_SECRET", ""),
		Expires:             getEnvAsDuration("JWT_EXPIRES", "15m"),
		RefreshTokenExpires: getEnvAsDuration("REFRESH_TOKEN_EXPIRES", "7d"),
	}

	config.OAuth = OAuth2Config{
		ClientID:     getEnv("OAUTH_CLIENT_ID", ""),
		ClientSecret: getEnv("OAUTH_CLIENT_SECRET", ""),
		RedirectURL:  getEnv("OAUTH_REDIRECT_URL", ""),
	}

	config.Security = SecurityConfig{
		BCryptCost:        getEnvAsInt("BCRYPT_COST", 12),
		RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnvAsDuration("RATE_LIMIT_WINDOW", "1h"),
	}

	config.Emergency = EmergencyConfig{
		AccessDuration:    getEnvAsDuration("EMERGENCY_ACCESS_DURATION", "1h"),
		NotificationEmail: getEnv("EMERGENCY_NOTIFICATION_EMAIL", "security@example.com"),
	}

	config.App = AppConfig{
		ServerPort:  getEnvAsInt("SERVER_PORT", 8080),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		EnableCORS:  getEnvAsBool("ENABLE_CORS", true),
		CORSOrigins: getEnvAsSlice("CORS_ORIGINS", []string{"http://localhost:3000"}),
	}

	config.SSL = SSLConfig{
		CertPath: getEnv("SSL_CERT_PATH", ""),
		KeyPath:  getEnv("SSL_KEY_PATH", ""),
	}

	config.Monitoring = MonitoringConfig{
		Enabled:     getEnvAsBool("METRICS_ENABLED", true),
		MetricsPort: getEnvAsInt("METRICS_PORT", 9090),
	}

	config.AI = AIConfig{
		AuthToken: getEnv("AI_AUTH_TOKEN", ""),
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	GlobalConfig = config
	return config, nil
}

func setDefaults() {
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.name", "healthsecure")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("app.server_port", 8080)
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.log_level", "info")
	viper.SetDefault("security.bcrypt_cost", 12)
	viper.SetDefault("monitoring.enabled", true)
	viper.SetDefault("monitoring.metrics_port", 9090)
}

func validateConfig(config *Config) error {
	// Database validation
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if config.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if config.Database.Password == "" {
		log.Println("WARNING: Database password is empty")
	}

	// JWT validation
	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if len(config.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}
	if config.JWT.Expires <= 0 {
		return fmt.Errorf("JWT expiration time must be positive")
	}

	// Security validation
	if config.Security.BCryptCost < 10 {
		return fmt.Errorf("BCrypt cost must be at least 10 for security")
	}
	if config.Security.BCryptCost > 15 {
		log.Printf("WARNING: BCrypt cost %d is very high, may impact performance", config.Security.BCryptCost)
	}

	// Production environment validation
	if config.App.Environment == "production" {
		if config.Database.TLSMode != "required" {
			log.Println("WARNING: TLS not required for database in production")
		}
		if config.SSL.CertPath == "" || config.SSL.KeyPath == "" {
			log.Println("WARNING: SSL certificates not configured for production")
		}
	}

	return nil
}

// Helper functions for environment variable parsing
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getViperString(key, defaultVal string) string {
	if value := viper.GetString(key); value != "" {
		return value
	}
	return defaultVal
}

func getViperInt(key string, defaultVal int) int {
	if viper.IsSet(key) {
		return viper.GetInt(key)
	}
	return defaultVal
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valueStr := getEnv(name, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsDuration(name string, defaultVal string) time.Duration {
	valueStr := getEnv(name, defaultVal)
	if duration, err := time.ParseDuration(valueStr); err == nil {
		return duration
	}
	// If parsing fails, try to parse as hours (for emergency access duration)
	if duration, err := time.ParseDuration(valueStr + "h"); err == nil {
		return duration
	}
	// Fallback to default
	if duration, err := time.ParseDuration(defaultVal); err == nil {
		return duration
	}
	return 15 * time.Minute // Safe fallback
}

func getEnvAsSlice(name string, defaultVal []string) []string {
	valueStr := getEnv(name, "")
	if valueStr == "" {
		return defaultVal
	}
	
	// Simple comma-separated parsing
	var result []string
	for _, v := range viper.GetStringSlice(name) {
		if v != "" {
			result = append(result, v)
		}
	}
	
	if len(result) == 0 {
		return defaultVal
	}
	return result
}

// IsProduction returns true if the application is running in production mode
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsDevelopment returns true if the application is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
	)
	
	if c.Database.TLSMode != "" {
		dsn += "&tls=" + c.Database.TLSMode
	}
	
	return dsn
}

// GetRedisAddress returns the Redis connection address
func (c *Config) GetRedisAddress() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// RequiresSSL returns true if SSL certificates are configured
func (c *Config) RequiresSSL() bool {
	return c.SSL.CertPath != "" && c.SSL.KeyPath != ""
}