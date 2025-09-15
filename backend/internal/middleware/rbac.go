package middleware

import (
	"net/http"

	"healthsecure/internal/models"

	"github.com/gin-gonic/gin"
)

// RBACMiddleware provides role-based access control
func RBACMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")
		if userRole == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role required",
			})
			c.Abort()
			return
		}

		// Store role as enum for easy comparison
		c.Set("user_role_enum", models.UserRole(userRole))
		c.Next()
	}
}

// CheckPatientAccess middleware checks if user can access patient data
func CheckPatientAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := models.UserRole(c.GetString("user_role"))
		
		// Only medical staff can access patient data
		if userRole != models.RoleDoctor && userRole != models.RoleNurse {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Medical staff access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckMedicalRecordAccess middleware checks access to medical records
func CheckMedicalRecordAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := models.UserRole(c.GetString("user_role"))
		
		// Only medical staff can access medical records
		if userRole != models.RoleDoctor && userRole != models.RoleNurse {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Medical staff access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}