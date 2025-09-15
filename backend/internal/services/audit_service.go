package services

import (
	"fmt"
	"time"

	"healthsecure/internal/database"
	"healthsecure/internal/models"

	"gorm.io/gorm"
)

type AuditService struct {
	db *gorm.DB
}

type AuditLogQuery struct {
	UserID      *uint                `form:"user_id"`
	PatientID   *uint                `form:"patient_id"`
	Action      *models.AuditAction  `form:"action"`
	Success     *bool                `form:"success"`
	Emergency   *bool                `form:"emergency"`
	StartTime   *time.Time           `form:"start_time"`
	EndTime     *time.Time           `form:"end_time"`
	IPAddress   string               `form:"ip_address"`
	Page        int                  `form:"page,default=1"`
	Limit       int                  `form:"limit,default=50"`
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{
		db: db,
	}
}

// LogUserAction logs a user action to the audit trail
func (s *AuditService) LogUserAction(userID uint, action models.AuditAction, resource, ipAddress, userAgent string, success bool, reason string) error {
	auditLog := &models.AuditLog{
		UserID:       &userID,
		Action:       action,
		Resource:     resource,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		EmergencyUse: false,
		Reason:       reason,
		Success:      success,
		Timestamp:    time.Now(),
	}

	if !success {
		auditLog.ErrorMessage = reason
	}

	return s.db.Create(auditLog).Error
}

// LogPatientAccess logs access to patient data
func (s *AuditService) LogPatientAccess(userID, patientID uint, action models.AuditAction, ipAddress, userAgent string, emergencyUse bool, reason string) error {
	auditLog := &models.AuditLog{
		UserID:       &userID,
		PatientID:    &patientID,
		Action:       action,
		Resource:     fmt.Sprintf("patient:%d", patientID),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		EmergencyUse: emergencyUse,
		Reason:       reason,
		Success:      true,
		Timestamp:    time.Now(),
	}

	if err := s.db.Create(auditLog).Error; err != nil {
		return fmt.Errorf("failed to log patient access: %w", err)
	}

	// If this is a security-sensitive event, create a security event as well
	if emergencyUse || action == models.ActionUnauthorized {
		s.createSecurityEvent(userID, ipAddress, auditLog)
	}

	return nil
}

// LogMedicalRecordAccess logs access to medical records
func (s *AuditService) LogMedicalRecordAccess(userID, patientID, recordID uint, action models.AuditAction, ipAddress, userAgent string, emergencyUse bool, reason string) error {
	auditLog := &models.AuditLog{
		UserID:       &userID,
		PatientID:    &patientID,
		RecordID:     &recordID,
		Action:       action,
		Resource:     fmt.Sprintf("medical_record:%d", recordID),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		EmergencyUse: emergencyUse,
		Reason:       reason,
		Success:      true,
		Timestamp:    time.Now(),
	}

	if err := s.db.Create(auditLog).Error; err != nil {
		return fmt.Errorf("failed to log medical record access: %w", err)
	}

	// If this is a security-sensitive event, create a security event as well
	if emergencyUse || action == models.ActionUnauthorized {
		s.createSecurityEvent(userID, ipAddress, auditLog)
	}

	return nil
}

// LogEmergencyAccess logs emergency access requests and usage
func (s *AuditService) LogEmergencyAccess(userID, patientID uint, action models.AuditAction, ipAddress, userAgent, reason string, success bool) error {
	auditLog := &models.AuditLog{
		UserID:       &userID,
		PatientID:    &patientID,
		Action:       action,
		Resource:     fmt.Sprintf("emergency_access:%d", patientID),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		EmergencyUse: true,
		Reason:       reason,
		Success:      success,
		Timestamp:    time.Now(),
	}

	if !success {
		auditLog.ErrorMessage = reason
	}

	if err := s.db.Create(auditLog).Error; err != nil {
		return fmt.Errorf("failed to log emergency access: %w", err)
	}

	// Always create a security event for emergency access
	s.createSecurityEvent(userID, ipAddress, auditLog)

	return nil
}

// LogFailedLogin logs failed login attempts
func (s *AuditService) LogFailedLogin(email, ipAddress, userAgent, reason string) error {
	// Temporarily disabled audit logging for failed logins due to database schema constraints
	// TODO: Re-enable after database schema is updated to allow NULL user_id
	
	// Log to console for debugging
	fmt.Printf("Failed login attempt: email=%s, reason=%s, ip=%s\n", email, reason, ipAddress)
	
	// Temporarily disabled security event creation as well
	// s.createSecurityEventForFailedLogin(email, ipAddress, reason)

	return nil
}

// LogUnauthorizedAccess logs unauthorized access attempts
func (s *AuditService) LogUnauthorizedAccess(userID uint, resource, ipAddress, userAgent, reason string) error {
	auditLog := &models.AuditLog{
		UserID:       &userID,
		Action:       models.ActionUnauthorized,
		Resource:     resource,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		EmergencyUse: false,
		Success:      false,
		ErrorMessage: reason,
		Timestamp:    time.Now(),
	}

	if err := s.db.Create(auditLog).Error; err != nil {
		return fmt.Errorf("failed to log unauthorized access: %w", err)
	}

	// Create security event for unauthorized access
	s.createSecurityEvent(userID, ipAddress, auditLog)

	return nil
}

// GetAuditLogs retrieves audit logs with filtering and pagination
func (s *AuditService) GetAuditLogs(query *AuditLogQuery, requestedByRole models.UserRole, requestedByUserID uint) ([]models.AuditLog, int64, error) {
	// Only admins and medical staff can view audit logs
	if requestedByRole != models.RoleAdmin && requestedByRole != models.RoleDoctor && requestedByRole != models.RoleNurse {
		return nil, 0, fmt.Errorf("insufficient permissions to view audit logs")
	}

	filter := &models.AuditLogFilter{
		UserID:    query.UserID,
		PatientID: query.PatientID,
		Action:    query.Action,
		Success:   query.Success,
		Emergency: query.Emergency,
		StartTime: query.StartTime,
		EndTime:   query.EndTime,
		IPAddress: query.IPAddress,
		Limit:     query.Limit,
		Offset:    (query.Page - 1) * query.Limit,
	}

	// Non-admins can only see their own audit logs or patient-related logs they have access to
	if requestedByRole != models.RoleAdmin {
		filter.UserID = &requestedByUserID
	}

	var auditLogs []models.AuditLog
	var total int64

	// Build query
	dbQuery := filter.Apply(s.db.Model(&models.AuditLog{}))

	// Count total records
	dbQuery.Count(&total)

	// Get paginated results with relationships
	if err := dbQuery.Preload("User").Preload("Patient").Find(&auditLogs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve audit logs: %w", err)
	}

	// Sanitize sensitive data for non-admin users
	if requestedByRole != models.RoleAdmin {
		for i := range auditLogs {
			if auditLogs[i].Patient != nil {
				auditLogs[i].Patient = auditLogs[i].Patient.SanitizeForRole(requestedByRole)
			}
		}
	}

	return auditLogs, total, nil
}

// GetUserAuditHistory gets audit history for a specific user
func (s *AuditService) GetUserAuditHistory(userID uint, requestedByRole models.UserRole, requestedByUserID uint, limit int) ([]models.AuditLog, error) {
	// Check permissions
	if requestedByRole != models.RoleAdmin && userID != requestedByUserID {
		return nil, fmt.Errorf("insufficient permissions to view user audit history")
	}

	var auditLogs []models.AuditLog
	query := s.db.Where("user_id = ?", userID).
		Order("timestamp DESC").
		Limit(limit).
		Preload("Patient")

	if err := query.Find(&auditLogs).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve user audit history: %w", err)
	}

	return auditLogs, nil
}

// GetPatientAuditHistory gets audit history for a specific patient
func (s *AuditService) GetPatientAuditHistory(patientID uint, requestedByRole models.UserRole, requestedByUserID uint, limit int) ([]models.AuditLog, error) {
	// Only medical staff can view patient audit history
	if requestedByRole != models.RoleAdmin && requestedByRole != models.RoleDoctor && requestedByRole != models.RoleNurse {
		return nil, fmt.Errorf("insufficient permissions to view patient audit history")
	}

	var auditLogs []models.AuditLog
	query := s.db.Where("patient_id = ?", patientID).
		Order("timestamp DESC").
		Limit(limit).
		Preload("User").
		Preload("Patient")

	if err := query.Find(&auditLogs).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve patient audit history: %w", err)
	}

	return auditLogs, nil
}

// GetSecurityEvents retrieves security events
func (s *AuditService) GetSecurityEvents(page, limit int, resolved *bool, severity *database.SecurityEventSeverity) ([]database.SecurityEvent, int64, error) {
	var events []database.SecurityEvent
	var total int64

	query := s.db.Model(&database.SecurityEvent{})

	if resolved != nil {
		query = query.Where("resolved = ?", *resolved)
	}
	if severity != nil {
		query = query.Where("severity = ?", *severity)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").
		Offset(offset).Limit(limit).
		Preload("User").Preload("ResolvedByUser").
		Find(&events).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve security events: %w", err)
	}

	return events, total, nil
}

// ResolveSecurityEvent marks a security event as resolved
func (s *AuditService) ResolveSecurityEvent(eventID, resolvedByUserID uint) error {
	now := time.Now()
	
	result := s.db.Model(&database.SecurityEvent{}).
		Where("id = ?", eventID).
		Updates(map[string]interface{}{
			"resolved":    true,
			"resolved_by": resolvedByUserID,
			"resolved_at": &now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to resolve security event: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("security event not found")
	}

	return nil
}

// GetAuditStatistics returns audit statistics
func (s *AuditService) GetAuditStatistics(startTime, endTime time.Time, requestedByRole models.UserRole) (map[string]interface{}, error) {
	if requestedByRole != models.RoleAdmin {
		return nil, fmt.Errorf("insufficient permissions to view audit statistics")
	}

	stats := make(map[string]interface{})

	// Total audit logs in time range
	var totalLogs int64
	s.db.Model(&models.AuditLog{}).
		Where("timestamp BETWEEN ? AND ?", startTime, endTime).
		Count(&totalLogs)
	stats["total_logs"] = totalLogs

	// Failed actions
	var failedActions int64
	s.db.Model(&models.AuditLog{}).
		Where("timestamp BETWEEN ? AND ? AND success = ?", startTime, endTime, false).
		Count(&failedActions)
	stats["failed_actions"] = failedActions

	// Emergency access events
	var emergencyEvents int64
	s.db.Model(&models.AuditLog{}).
		Where("timestamp BETWEEN ? AND ? AND emergency_use = ?", startTime, endTime, true).
		Count(&emergencyEvents)
	stats["emergency_events"] = emergencyEvents

	// Unique users
	var uniqueUsers int64
	s.db.Model(&models.AuditLog{}).
		Where("timestamp BETWEEN ? AND ?", startTime, endTime).
		Distinct("user_id").Count(&uniqueUsers)
	stats["unique_users"] = uniqueUsers

	// Actions by type
	var actionStats []struct {
		Action models.AuditAction `json:"action"`
		Count  int64              `json:"count"`
	}
	s.db.Model(&models.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("timestamp BETWEEN ?", startTime, endTime).
		Group("action").
		Scan(&actionStats)
	stats["actions_by_type"] = actionStats

	return stats, nil
}

// createSecurityEvent creates a security event based on audit log
func (s *AuditService) createSecurityEvent(userID uint, ipAddress string, auditLog *models.AuditLog) {
	var eventType database.SecurityEventType
	var severity database.SecurityEventSeverity
	var description string

	switch auditLog.Action {
	case models.ActionUnauthorized:
		eventType = database.SecurityEventUnauthorizedAccess
		severity = database.SecuritySeverityHigh
		description = fmt.Sprintf("Unauthorized access attempt to %s", auditLog.Resource)
	case models.ActionEmergencyAccess:
		eventType = database.SecurityEventEmergencyAccess
		severity = database.SecuritySeverityMedium
		description = fmt.Sprintf("Emergency access activated for %s", auditLog.Resource)
	case models.ActionEmergencyRequest:
		eventType = database.SecurityEventEmergencyAccess
		severity = database.SecuritySeverityMedium
		description = fmt.Sprintf("Emergency access requested for %s", auditLog.Resource)
	default:
		if !auditLog.Success {
			eventType = database.SecurityEventSuspiciousActivity
			severity = database.SecuritySeverityMedium
			description = fmt.Sprintf("Failed action: %s on %s", auditLog.Action, auditLog.Resource)
		} else {
			return // No security event needed for successful regular actions
		}
	}

	securityEvent := &database.SecurityEvent{
		EventType:   eventType,
		Severity:    severity,
		UserID:      &userID,
		IPAddress:   ipAddress,
		Description: description,
		Details:     fmt.Sprintf(`{"audit_log_id": %d, "resource": "%s", "reason": "%s"}`, auditLog.ID, auditLog.Resource, auditLog.Reason),
		Resolved:    false,
	}

	s.db.Create(securityEvent)
}

// createSecurityEventForFailedLogin creates a security event for failed login
func (s *AuditService) createSecurityEventForFailedLogin(email, ipAddress, reason string) {
	securityEvent := &database.SecurityEvent{
		EventType:   database.SecurityEventFailedLogin,
		Severity:    database.SecuritySeverityMedium,
		IPAddress:   ipAddress,
		Description: fmt.Sprintf("Failed login attempt for email: %s", email),
		Details:     fmt.Sprintf(`{"email": "%s", "reason": "%s"}`, email, reason),
		Resolved:    false,
	}

	s.db.Create(securityEvent)
}