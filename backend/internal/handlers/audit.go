package handlers

import (
	"net/http"
	"strconv"
	"time"

	"healthsecure/internal/auth"
	"healthsecure/internal/models"
	"healthsecure/internal/services"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	auditService *services.AuditService
	jwtService   *auth.JWTService
}

func NewAuditHandler(auditService *services.AuditService, jwtService *auth.JWTService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
		jwtService:   jwtService,
	}
}

// GetAuditLogs retrieves audit logs with filtering
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	var query services.AuditLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default pagination
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 50
	}

	logs, total, err := h.auditService.GetAuditLogs(&query, userRole, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"audit_logs": logs,
		"pagination": gin.H{
			"current_page": query.Page,
			"limit":        query.Limit,
			"total":        total,
			"total_pages":  (total + int64(query.Limit) - 1) / int64(query.Limit),
		},
	})
}

// GetUserAuditHistory gets audit history for specific user
func (h *AuditHandler) GetUserAuditHistory(c *gin.Context) {
	requestedByUserID := c.GetUint("user_id")
	requestedByRole := models.UserRole(c.GetString("user_role"))

	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 200 {
		limit = 50
	}

	logs, err := h.auditService.GetUserAuditHistory(uint(userID), requestedByRole, requestedByUserID, limit)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"audit_history": logs})
}

// GetPatientAuditHistory gets audit history for specific patient
func (h *AuditHandler) GetPatientAuditHistory(c *gin.Context) {
	requestedByUserID := c.GetUint("user_id")
	requestedByRole := models.UserRole(c.GetString("user_role"))

	patientIDStr := c.Param("id")
	patientID, err := strconv.ParseUint(patientIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient ID"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 200 {
		limit = 50
	}

	logs, err := h.auditService.GetPatientAuditHistory(uint(patientID), requestedByRole, requestedByUserID, limit)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"audit_history": logs})
}

// GetSecurityEvents retrieves security events
func (h *AuditHandler) GetSecurityEvents(c *gin.Context) {
	page, limit := getPaginationParams(c)

	var resolved *bool
	if resolvedStr := c.Query("resolved"); resolvedStr != "" {
		if r, err := strconv.ParseBool(resolvedStr); err == nil {
			resolved = &r
		}
	}

	events, total, err := h.auditService.GetSecurityEvents(page, limit, resolved, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"security_events": events,
		"pagination": gin.H{
			"current_page": page,
			"limit":        limit,
			"total":        total,
			"total_pages":  (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// ResolveSecurityEvent marks a security event as resolved
func (h *AuditHandler) ResolveSecurityEvent(c *gin.Context) {
	userID := c.GetUint("user_id")

	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err := h.auditService.ResolveSecurityEvent(uint(eventID), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Security event resolved successfully"})
}

// GetAuditStatistics returns audit statistics
func (h *AuditHandler) GetAuditStatistics(c *gin.Context) {
	userRole := models.UserRole(c.GetString("user_role"))

	// Default to last 30 days
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -30)

	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = t
		}
	}

	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = t
		}
	}

	stats, err := h.auditService.GetAuditStatistics(startTime, endTime, userRole)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
		"time_range": gin.H{
			"start": startTime,
			"end":   endTime,
		},
	})
}