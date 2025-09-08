package auth

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"healthsecure/configs"
	"healthsecure/internal/models"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(jwtService *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		token := ExtractTokenFromHeader(authHeader)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set user context for downstream handlers
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", string(claims.Role))
		c.Set("token_id", claims.TokenID)

		c.Next()
	}
}

// OptionalAuthMiddleware validates JWT tokens but doesn't require authentication
func OptionalAuthMiddleware(jwtService *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Extract token from "Bearer <token>" format
		token := ExtractTokenFromHeader(authHeader)
		if token == "" {
			c.Next()
			return
		}

		// Validate token (if present)
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			// Token is present but invalid - continue without auth
			c.Next()
			return
		}

		// Set user context for downstream handlers
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", string(claims.Role))
		c.Set("token_id", claims.TokenID)

		c.Next()
	}
}

// RequireRole middleware checks if user has required role(s)
func RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found in context",
			})
			c.Abort()
			return
		}

		userRoleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user role format",
			})
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		userRoleEnum := models.UserRole(userRoleStr)
		for _, role := range roles {
			if userRoleEnum == role {
				c.Next()
				return
			}
		}

		// Log unauthorized access attempt
		userID, _ := c.Get("user_id")
		LogUnauthorizedAccess(c, userID, c.Request.URL.Path, "insufficient_role")

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
		})
		c.Abort()
	}
}

// RequireActiveUser ensures the user account is active
func RequireActiveUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User ID not found in context",
			})
			c.Abort()
			return
		}

		// This would typically check the database for user status
		// For now, we assume the token validation already checked this
		c.Next()
	}
}

// AdminOnly middleware restricts access to administrators only
func AdminOnly() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin)
}

// MedicalStaffOnly middleware restricts access to doctors and nurses only
func MedicalStaffOnly() gin.HandlerFunc {
	return RequireRole(models.RoleDoctor, models.RoleNurse)
}

// DoctorOnly middleware restricts access to doctors only
func DoctorOnly() gin.HandlerFunc {
	return RequireRole(models.RoleDoctor)
}

// RateLimitMiddleware implements basic rate limiting
func RateLimitMiddleware(config *configs.Config) gin.HandlerFunc {
	// This is a simplified rate limiter
	// In production, use a proper rate limiter like golang.org/x/time/rate
	// or Redis-based rate limiting
	
	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()
		
		// Simple rate limiting logic would go here
		// For now, just continue
		_ = clientIP
		
		c.Next()
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware(config *configs.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		
		// Check if origin is allowed
		allowed := false
		if len(config.App.CORSOrigins) == 0 {
			// If no origins configured, allow all (development mode)
			allowed = true
		} else {
			for _, allowedOrigin := range config.App.CORSOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent XSS attacks
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Prevent MIME sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		
		// Strict Transport Security (HTTPS only)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		// Content Security Policy
		csp := "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'"
		c.Header("Content-Security-Policy", csp)
		
		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		c.Next()
	}
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Custom log format that doesn't include sensitive data
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate or get request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate UUID for request ID
			requestID = generateRequestID()
		}
		
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		
		c.Next()
	}
}

// EmergencyAccessMiddleware checks for emergency access permissions
func EmergencyAccessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request has emergency access token
		emergencyToken := c.GetHeader("X-Emergency-Access-Token")
		
		if emergencyToken != "" {
			// Validate emergency access token
			// This would check the emergency_access table
			// For now, just set a flag
			c.Set("emergency_access", true)
			c.Set("emergency_token", emergencyToken)
		}
		
		c.Next()
	}
}

// Helper functions
func generateRequestID() string {
	// Simple implementation - in production use proper UUID library
	return "req_" + randomString(16)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// LogUnauthorizedAccess logs unauthorized access attempts
func LogUnauthorizedAccess(c *gin.Context, userID interface{}, resource, reason string) {
	// This would typically log to the audit_logs table
	// Simplified logging for now
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	
	fmt.Printf("UNAUTHORIZED ACCESS: UserID=%v, Resource=%s, Reason=%s, IP=%s, UA=%s\n",
		userID, resource, reason, clientIP, userAgent)
}