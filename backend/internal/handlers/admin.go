package handlers

import (
	"net/http"
	"strconv"

	"healthsecure/internal/auth"
	"healthsecure/internal/models"
	"healthsecure/internal/services"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	userService  *services.UserService
	auditService *services.AuditService
	jwtService   *auth.JWTService
}

func NewAdminHandler(userService *services.UserService, auditService *services.AuditService, jwtService *auth.JWTService) *AdminHandler {
	return &AdminHandler{
		userService:  userService,
		auditService: auditService,
		jwtService:   jwtService,
	}
}

// GetAllUsers retrieves all users
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	userRole := models.UserRole(c.GetString("user_role"))
	page, limit := getPaginationParams(c)

	users, total, err := h.userService.GetAllUsers(userRole, page, limit)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"current_page": page,
			"limit":        limit,
			"total":        total,
			"total_pages":  (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// CreateUser creates a new user
func (h *AdminHandler) CreateUser(c *gin.Context) {
	createdByUserID := c.GetUint("user_id")

	var req services.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.CreateUser(&req, createdByUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user,
	})
}

// GetUser retrieves a user by ID
func (h *AdminHandler) GetUser(c *gin.Context) {
	requestedByUserID := c.GetUint("user_id")
	requestedByRole := models.UserRole(c.GetString("user_role"))

	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUser(uint(userID), requestedByUserID, requestedByRole)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// UpdateUser updates user information
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	updatedByUserID := c.GetUint("user_id")
	updatedByRole := models.UserRole(c.GetString("user_role"))

	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateUser(uint(userID), &req, updatedByUserID, updatedByRole)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    user,
	})
}

// DeactivateUser deactivates a user account
func (h *AdminHandler) DeactivateUser(c *gin.Context) {
	deactivatedByUserID := c.GetUint("user_id")
	deactivatedByRole := models.UserRole(c.GetString("user_role"))

	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.userService.DeactivateUser(uint(userID), deactivatedByUserID, deactivatedByRole); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deactivated successfully"})
}

// GetUserSessions retrieves user sessions
func (h *AdminHandler) GetUserSessions(c *gin.Context) {
	requestedByUserID := c.GetUint("user_id")
	requestedByRole := models.UserRole(c.GetString("user_role"))

	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	sessions, err := h.userService.GetUserSessions(uint(userID), requestedByUserID, requestedByRole)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

// GetDashboardStats returns dashboard statistics
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	// This would gather various statistics for the admin dashboard
	// For now, return placeholder data
	stats := gin.H{
		"total_users":    150,
		"active_users":   142,
		"total_patients": 1250,
		"recent_logins":  25,
		"security_alerts": 3,
		"system_status": "healthy",
	}

	c.JSON(http.StatusOK, gin.H{"dashboard_stats": stats})
}