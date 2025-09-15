package handlers

import (
	"net/http"
	"strconv"

	"healthsecure/internal/auth"
	"healthsecure/internal/models"
	"healthsecure/internal/services"

	"github.com/gin-gonic/gin"
)

type EmergencyHandler struct {
	emergencyService *services.EmergencyService
	jwtService       *auth.JWTService
}

func NewEmergencyHandler(emergencyService *services.EmergencyService, jwtService *auth.JWTService) *EmergencyHandler {
	return &EmergencyHandler{
		emergencyService: emergencyService,
		jwtService:       jwtService,
	}
}

// RequestEmergencyAccess handles emergency access requests
func (h *EmergencyHandler) RequestEmergencyAccess(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	var req services.EmergencyAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	response, err := h.emergencyService.RequestEmergencyAccess(&req, userID, userRole, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Emergency access requested successfully",
		"data":    response,
	})
}

// ActivateEmergencyAccess activates an emergency access token
func (h *EmergencyHandler) ActivateEmergencyAccess(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	accessIDStr := c.Param("id")
	accessID, err := strconv.ParseUint(accessIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid access ID"})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.emergencyService.ActivateEmergencyAccess(uint(accessID), userID, userRole, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Emergency access activated successfully"})
}

// RevokeEmergencyAccess revokes an emergency access token
func (h *EmergencyHandler) RevokeEmergencyAccess(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	accessIDStr := c.Param("id")
	accessID, err := strconv.ParseUint(accessIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid access ID"})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.emergencyService.RevokeEmergencyAccess(uint(accessID), userID, userRole, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Emergency access revoked successfully"})
}

// GetActiveEmergencyAccess gets all active emergency access sessions
func (h *EmergencyHandler) GetActiveEmergencyAccess(c *gin.Context) {
	userRole := models.UserRole(c.GetString("user_role"))

	records, err := h.emergencyService.GetActiveEmergencyAccess(userRole)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"active_sessions": records})
}

// GetUserEmergencyAccess gets emergency access records for a user
func (h *EmergencyHandler) GetUserEmergencyAccess(c *gin.Context) {
	requestedByUserID := c.GetUint("user_id")
	requestedByRole := models.UserRole(c.GetString("user_role"))

	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	page, limit := getPaginationParams(c)

	records, total, err := h.emergencyService.GetUserEmergencyAccess(uint(userID), requestedByUserID, requestedByRole, page, limit)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"emergency_access": records,
		"pagination": gin.H{
			"current_page": page,
			"limit":        limit,
			"total":        total,
			"total_pages":  (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetPatientEmergencyAccess gets emergency access records for a patient
func (h *EmergencyHandler) GetPatientEmergencyAccess(c *gin.Context) {
	userRole := models.UserRole(c.GetString("user_role"))

	patientIDStr := c.Param("id")
	patientID, err := strconv.ParseUint(patientIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient ID"})
		return
	}

	page, limit := getPaginationParams(c)

	records, total, err := h.emergencyService.GetPatientEmergencyAccess(uint(patientID), userRole, page, limit)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"emergency_access": records,
		"pagination": gin.H{
			"current_page": page,
			"limit":        limit,
			"total":        total,
			"total_pages":  (total + int64(limit) - 1) / int64(limit),
		},
	})
}