package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"healthsecure/configs"
	"healthsecure/internal/models"

	"gorm.io/gorm"
)

type EmergencyService struct {
	db     *gorm.DB
	audit  *AuditService
	config *configs.Config
}

type EmergencyAccessRequest struct {
	PatientID uint   `json:"patient_id" binding:"required"`
	Reason    string `json:"reason" binding:"required,min=20"`
}

type EmergencyAccessResponse struct {
	ID          uint      `json:"id"`
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	Status      string    `json:"status"`
}

func NewEmergencyService(db *gorm.DB, audit *AuditService, config *configs.Config) *EmergencyService {
	return &EmergencyService{
		db:     db,
		audit:  audit,
		config: config,
	}
}

// RequestEmergencyAccess creates a new emergency access request
func (s *EmergencyService) RequestEmergencyAccess(req *EmergencyAccessRequest, requestedByUserID uint, requestedByRole models.UserRole, ipAddress, userAgent string) (*EmergencyAccessResponse, error) {
	// Only medical staff can request emergency access
	if requestedByRole != models.RoleDoctor && requestedByRole != models.RoleNurse {
		s.audit.LogUnauthorizedAccess(requestedByUserID, fmt.Sprintf("emergency_access:patient_%d", req.PatientID), ipAddress, userAgent, "insufficient_role")
		return nil, fmt.Errorf("insufficient permissions to request emergency access")
	}

	// Verify patient exists
	var patient models.Patient
	if err := s.db.Where("id = ?", req.PatientID).First(&patient).Error; err != nil {
		return nil, fmt.Errorf("patient not found")
	}

	// Check if user already has active emergency access for this patient
	var existingAccess models.EmergencyAccess
	if err := s.db.Where("user_id = ? AND patient_id = ? AND status IN ? AND expires_at > ?", 
		requestedByUserID, req.PatientID, []string{"pending", "active"}, time.Now()).
		First(&existingAccess).Error; err == nil {
		return nil, fmt.Errorf("active emergency access already exists for this patient")
	}

	// Generate secure access token
	token, err := s.generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Create emergency access record
	emergencyAccess := models.EmergencyAccess{
		UserID:      requestedByUserID,
		PatientID:   req.PatientID,
		Reason:      req.Reason,
		AccessToken: token,
		Status:      models.EmergencyStatusPending,
		ExpiresAt:   time.Now().Add(s.config.Emergency.AccessDuration),
	}

	if err := s.db.Create(&emergencyAccess).Error; err != nil {
		return nil, fmt.Errorf("failed to create emergency access: %w", err)
	}

	// Log emergency access request
	s.audit.LogEmergencyAccess(requestedByUserID, req.PatientID, models.ActionEmergencyRequest, ipAddress, userAgent, req.Reason, true)

	// Auto-activate for now (in production, might require approval)
	emergencyAccess.Status = models.EmergencyStatusActive
	s.db.Save(&emergencyAccess)

	response := &EmergencyAccessResponse{
		ID:          emergencyAccess.ID,
		AccessToken: emergencyAccess.AccessToken,
		ExpiresAt:   emergencyAccess.ExpiresAt,
		Status:      string(emergencyAccess.Status),
	}

	return response, nil
}

// ValidateEmergencyAccess validates an emergency access token
func (s *EmergencyService) ValidateEmergencyAccess(token string, userID, patientID uint) (*models.EmergencyAccess, error) {
	var access models.EmergencyAccess
	if err := s.db.Where("access_token = ? AND user_id = ? AND patient_id = ?", token, userID, patientID).
		Preload("User").Preload("Patient").First(&access).Error; err != nil {
		return nil, fmt.Errorf("invalid emergency access token")
	}

	// Update status if needed
	access.UpdateStatus()
	s.db.Save(&access)

	if !access.IsActive() {
		return nil, fmt.Errorf("emergency access token is not active or has expired")
	}

	return &access, nil
}

// ActivateEmergencyAccess activates a pending emergency access request
func (s *EmergencyService) ActivateEmergencyAccess(accessID uint, activatedByUserID uint, activatedByRole models.UserRole, ipAddress, userAgent string) error {
	// Only the requesting user or admin can activate emergency access
	var access models.EmergencyAccess
	if err := s.db.Where("id = ?", accessID).First(&access).Error; err != nil {
		return fmt.Errorf("emergency access not found")
	}

	if access.UserID != activatedByUserID && activatedByRole != models.RoleAdmin {
		s.audit.LogUnauthorizedAccess(activatedByUserID, fmt.Sprintf("emergency_access:%d", accessID), ipAddress, userAgent, "not_authorized_to_activate")
		return fmt.Errorf("unauthorized to activate this emergency access")
	}

	if err := access.Activate(); err != nil {
		return fmt.Errorf("cannot activate emergency access: %w", err)
	}

	if err := s.db.Save(&access).Error; err != nil {
		return fmt.Errorf("failed to activate emergency access: %w", err)
	}

	// Log activation
	s.audit.LogEmergencyAccess(activatedByUserID, access.PatientID, models.ActionEmergencyAccess, ipAddress, userAgent, "emergency_access_activated", true)

	return nil
}

// RevokeEmergencyAccess revokes an active emergency access
func (s *EmergencyService) RevokeEmergencyAccess(accessID uint, revokedByUserID uint, revokedByRole models.UserRole, ipAddress, userAgent string) error {
	var access models.EmergencyAccess
	if err := s.db.Where("id = ?", accessID).First(&access).Error; err != nil {
		return fmt.Errorf("emergency access not found")
	}

	// Only the requesting user, admin, or system can revoke
	if access.UserID != revokedByUserID && revokedByRole != models.RoleAdmin {
		s.audit.LogUnauthorizedAccess(revokedByUserID, fmt.Sprintf("emergency_access:%d", accessID), ipAddress, userAgent, "not_authorized_to_revoke")
		return fmt.Errorf("unauthorized to revoke this emergency access")
	}

	if err := access.Revoke(revokedByUserID); err != nil {
		return fmt.Errorf("cannot revoke emergency access: %w", err)
	}

	if err := s.db.Save(&access).Error; err != nil {
		return fmt.Errorf("failed to revoke emergency access: %w", err)
	}

	// Log revocation
	s.audit.LogEmergencyAccess(revokedByUserID, access.PatientID, models.ActionUpdate, ipAddress, userAgent, "emergency_access_revoked", true)

	return nil
}

// GetUserEmergencyAccess retrieves emergency access records for a user
func (s *EmergencyService) GetUserEmergencyAccess(userID uint, requestedByUserID uint, requestedByRole models.UserRole, page, limit int) ([]models.EmergencyAccess, int64, error) {
	// Users can view their own emergency access, admins can view all
	if userID != requestedByUserID && requestedByRole != models.RoleAdmin {
		return nil, 0, fmt.Errorf("insufficient permissions to view emergency access records")
	}

	var records []models.EmergencyAccess
	var total int64

	query := s.db.Where("user_id = ?", userID)
	query.Model(&models.EmergencyAccess{}).Count(&total)

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).
		Preload("User").Preload("Patient").Preload("RevokedByUser").
		Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve emergency access records: %w", err)
	}

	// Update statuses
	for i := range records {
		records[i].UpdateStatus()
	}

	return records, total, nil
}

// GetActiveEmergencyAccess retrieves all currently active emergency access sessions
func (s *EmergencyService) GetActiveEmergencyAccess(requestedByRole models.UserRole) ([]models.EmergencyAccess, error) {
	if requestedByRole != models.RoleAdmin {
		return nil, fmt.Errorf("insufficient permissions to view active emergency access sessions")
	}

	var records []models.EmergencyAccess
	if err := s.db.Where("status = ? AND expires_at > ? AND revoked_at IS NULL", 
		models.EmergencyStatusActive, time.Now()).
		Preload("User").Preload("Patient").
		Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve active emergency access: %w", err)
	}

	return records, nil
}

// GetPatientEmergencyAccess retrieves emergency access records for a specific patient
func (s *EmergencyService) GetPatientEmergencyAccess(patientID uint, requestedByRole models.UserRole, page, limit int) ([]models.EmergencyAccess, int64, error) {
	if requestedByRole != models.RoleAdmin {
		return nil, 0, fmt.Errorf("insufficient permissions to view patient emergency access records")
	}

	var records []models.EmergencyAccess
	var total int64

	query := s.db.Where("patient_id = ?", patientID)
	query.Model(&models.EmergencyAccess{}).Count(&total)

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).
		Preload("User").Preload("Patient").Preload("RevokedByUser").
		Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve emergency access records: %w", err)
	}

	// Update statuses
	for i := range records {
		records[i].UpdateStatus()
	}

	return records, total, nil
}

// CleanupExpiredEmergencyAccess updates expired emergency access records
func (s *EmergencyService) CleanupExpiredEmergencyAccess() error {
	result := s.db.Model(&models.EmergencyAccess{}).
		Where("expires_at < ? AND status NOT IN ?", time.Now(), []string{"expired", "revoked"}).
		Update("status", models.EmergencyStatusExpired)

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired emergency access: %w", result.Error)
	}

	return nil
}

// GetEmergencyAccessStatistics returns statistics about emergency access usage
func (s *EmergencyService) GetEmergencyAccessStatistics(startTime, endTime time.Time, requestedByRole models.UserRole) (map[string]interface{}, error) {
	if requestedByRole != models.RoleAdmin {
		return nil, fmt.Errorf("insufficient permissions to view emergency access statistics")
	}

	stats := make(map[string]interface{})

	// Total emergency access requests in time range
	var totalRequests int64
	s.db.Model(&models.EmergencyAccess{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Count(&totalRequests)
	stats["total_requests"] = totalRequests

	// Active sessions
	var activeSessions int64
	s.db.Model(&models.EmergencyAccess{}).
		Where("status = ? AND expires_at > ?", models.EmergencyStatusActive, time.Now()).
		Count(&activeSessions)
	stats["active_sessions"] = activeSessions

	// Requests by status
	var statusStats []struct {
		Status models.EmergencyAccessStatus `json:"status"`
		Count  int64                        `json:"count"`
	}
	s.db.Model(&models.EmergencyAccess{}).
		Select("status, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Group("status").Scan(&statusStats)
	stats["requests_by_status"] = statusStats

	// Average duration of access
	var avgDuration float64
	s.db.Model(&models.EmergencyAccess{}).
		Select("AVG(TIMESTAMPDIFF(MINUTE, created_at, COALESCE(revoked_at, expires_at))) as avg_duration").
		Where("created_at BETWEEN ? AND ? AND status IN ?", startTime, endTime, []string{"used", "expired", "revoked"}).
		Scan(&avgDuration)
	stats["average_duration_minutes"] = avgDuration

	return stats, nil
}

// generateSecureToken generates a cryptographically secure token
func (s *EmergencyService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// IsEmergencyAccessActive checks if a user has active emergency access for a patient
func (s *EmergencyService) IsEmergencyAccessActive(userID, patientID uint) bool {
	var count int64
	s.db.Model(&models.EmergencyAccess{}).
		Where("user_id = ? AND patient_id = ? AND status = ? AND expires_at > ?", 
			userID, patientID, models.EmergencyStatusActive, time.Now()).
		Count(&count)
	return count > 0
}