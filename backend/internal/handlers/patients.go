package handlers

import (
	"net/http"
	"strconv"

	"healthsecure/internal/auth"
	"healthsecure/internal/models"
	"healthsecure/internal/services"

	"github.com/gin-gonic/gin"
)

type PatientHandler struct {
	patientService *services.PatientService
	jwtService     *auth.JWTService
}

func NewPatientHandler(patientService *services.PatientService, jwtService *auth.JWTService) *PatientHandler {
	return &PatientHandler{
		patientService: patientService,
		jwtService:     jwtService,
	}
}

// CreatePatient creates a new patient
func (h *PatientHandler) CreatePatient(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	var req services.CreatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	patient, err := h.patientService.CreatePatient(&req, userID, userRole, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Patient created successfully",
		"patient": patient,
	})
}

// GetPatient retrieves a patient by ID
func (h *PatientHandler) GetPatient(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	patientIDStr := c.Param("id")
	patientID, err := strconv.ParseUint(patientIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient ID"})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Check for emergency access
	emergencyAccess := h.checkEmergencyAccess(c, userID, uint(patientID))

	patient, err := h.patientService.GetPatient(uint(patientID), userID, userRole, ipAddress, userAgent, emergencyAccess)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"patient": patient})
}

// GetPatients retrieves patients with filtering and pagination
func (h *PatientHandler) GetPatients(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	var query services.PatientSearchQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default pagination
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 || query.Limit > 50 {
		query.Limit = 20
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	patients, total, err := h.patientService.GetPatients(&query, userID, userRole, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"patients": patients,
		"pagination": gin.H{
			"current_page": query.Page,
			"limit":        query.Limit,
			"total":        total,
			"total_pages":  (total + int64(query.Limit) - 1) / int64(query.Limit),
		},
	})
}

// UpdatePatient updates a patient's information
func (h *PatientHandler) UpdatePatient(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	patientIDStr := c.Param("id")
	patientID, err := strconv.ParseUint(patientIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient ID"})
		return
	}

	var req services.UpdatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	patient, err := h.patientService.UpdatePatient(uint(patientID), &req, userID, userRole, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Patient updated successfully",
		"patient": patient,
	})
}

// DeletePatient deletes a patient (admin only)
func (h *PatientHandler) DeletePatient(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	patientIDStr := c.Param("id")
	patientID, err := strconv.ParseUint(patientIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient ID"})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.patientService.DeletePatient(uint(patientID), userID, userRole, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Patient deleted successfully"})
}

// SearchPatients searches patients by name
func (h *PatientHandler) SearchPatients(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search name is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 50 {
		limit = 10
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	patients, err := h.patientService.SearchPatientsByName(name, userID, userRole, ipAddress, userAgent, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"patients": patients,
		"search":   name,
	})
}

// GetPatientWithRecords retrieves a patient with their medical records
func (h *PatientHandler) GetPatientWithRecords(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	patientIDStr := c.Param("id")
	patientID, err := strconv.ParseUint(patientIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient ID"})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Check for emergency access
	emergencyAccess := h.checkEmergencyAccess(c, userID, uint(patientID))

	patient, err := h.patientService.GetPatientWithMedicalRecords(uint(patientID), userID, userRole, ipAddress, userAgent, emergencyAccess)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"patient": patient})
}

// GetPatientStatistics returns patient statistics (admin only)
func (h *PatientHandler) GetPatientStatistics(c *gin.Context) {
	userRole := models.UserRole(c.GetString("user_role"))

	if userRole != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	stats, err := h.patientService.GetPatientStatistics(userRole)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statistics": stats})
}

// Helper function to check for emergency access
func (h *PatientHandler) checkEmergencyAccess(c *gin.Context, userID, patientID uint) bool {
	emergencyToken := c.GetHeader("X-Emergency-Access-Token")
	if emergencyToken == "" {
		return false
	}

	// This would validate the emergency access token
	// For now, we'll just check if the header exists
	// In a real implementation, you would validate the token against the database
	return true
}