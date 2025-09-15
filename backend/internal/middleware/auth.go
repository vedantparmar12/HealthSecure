package middleware

import (
	"net/http"

	"healthsecure/internal/auth"
	"healthsecure/internal/models"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
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
		token := auth.ExtractTokenFromHeader(authHeader)
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

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
		})
		c.Abort()
	}
}

// AdminOnly restricts access to administrators only
func AdminOnly() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin)
}

// MedicalStaffOnly restricts access to doctors and nurses only
func MedicalStaffOnly() gin.HandlerFunc {
	return RequireRole(models.RoleDoctor, models.RoleNurse)
}

// DoctorOnly restricts access to doctors only
func DoctorOnly() gin.HandlerFunc {
	return RequireRole(models.RoleDoctor)
}