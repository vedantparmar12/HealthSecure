package handlers

import (
	"net/http"
	"strconv"

	"healthsecure/internal/auth"
	"healthsecure/internal/models"
	"healthsecure/internal/services"

	"github.com/gin-gonic/gin"
)

type MedicalRecordHandler struct {
	recordService *services.MedicalRecordService
	jwtService    *auth.JWTService
}

func NewMedicalRecordHandler(recordService *services.MedicalRecordService, jwtService *auth.JWTService) *MedicalRecordHandler {
	return &MedicalRecordHandler{
		recordService: recordService,
		jwtService:    jwtService,
	}
}

// CreateMedicalRecord creates a new medical record
func (h *MedicalRecordHandler) CreateMedicalRecord(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	var req services.CreateMedicalRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get patient ID from URL parameter
	patientIDStr := c.Param("id")
	patientID, err := strconv.ParseUint(patientIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient ID"})
		return
	}
	req.PatientID = uint(patientID)

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	record, err := h.recordService.CreateMedicalRecord(&req, userID, userRole, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Medical record created successfully",
		"record":  record,
	})
}

// GetMedicalRecord retrieves a medical record by ID
func (h *MedicalRecordHandler) GetMedicalRecord(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	recordIDStr := c.Param("id")
	recordID, err := strconv.ParseUint(recordIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Check for emergency access
	emergencyAccess := c.GetHeader("X-Emergency-Access-Token") != ""

	record, err := h.recordService.GetMedicalRecord(uint(recordID), userID, userRole, ipAddress, userAgent, emergencyAccess)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"record": record})
}

// GetPatientMedicalRecords retrieves medical records for a patient
func (h *MedicalRecordHandler) GetPatientMedicalRecords(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	patientIDStr := c.Param("id")
	patientID, err := strconv.ParseUint(patientIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient ID"})
		return
	}

	page, limit := getPaginationParams(c)
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Check for emergency access
	emergencyAccess := c.GetHeader("X-Emergency-Access-Token") != ""

	records, total, err := h.recordService.GetPatientMedicalRecords(uint(patientID), userID, userRole, ipAddress, userAgent, emergencyAccess, page, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records": records,
		"pagination": gin.H{
			"current_page": page,
			"limit":        limit,
			"total":        total,
			"total_pages":  (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// UpdateMedicalRecord updates a medical record
func (h *MedicalRecordHandler) UpdateMedicalRecord(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	recordIDStr := c.Param("id")
	recordID, err := strconv.ParseUint(recordIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	var req services.UpdateMedicalRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	record, err := h.recordService.UpdateMedicalRecord(uint(recordID), &req, userID, userRole, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Medical record updated successfully",
		"record":  record,
	})
}