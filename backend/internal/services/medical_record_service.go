package services

import (
	"fmt"
	"time"

	"healthsecure/internal/models"

	"gorm.io/gorm"
)

type MedicalRecordService struct {
	db    *gorm.DB
	audit *AuditService
}

type CreateMedicalRecordRequest struct {
	PatientID   uint                    `json:"patient_id" binding:"required"`
	Diagnosis   string                  `json:"diagnosis" binding:"required"`
	Treatment   string                  `json:"treatment" binding:"required"`
	Notes       string                  `json:"notes"`
	Medications string                  `json:"medications"`
	Severity    models.SeverityLevel    `json:"severity" binding:"required"`
}

type UpdateMedicalRecordRequest struct {
	Diagnosis   *string                 `json:"diagnosis,omitempty"`
	Treatment   *string                 `json:"treatment,omitempty"`
	Notes       *string                 `json:"notes,omitempty"`
	Medications *string                 `json:"medications,omitempty"`
	Severity    *models.SeverityLevel   `json:"severity,omitempty"`
}

func NewMedicalRecordService(db *gorm.DB, audit *AuditService) *MedicalRecordService {
	return &MedicalRecordService{
		db:    db,
		audit: audit,
	}
}

// CreateMedicalRecord creates a new medical record
func (s *MedicalRecordService) CreateMedicalRecord(req *CreateMedicalRecordRequest, createdByUserID uint, createdByRole models.UserRole, ipAddress, userAgent string) (*models.MedicalRecord, error) {
	// Only doctors can create medical records
	if createdByRole != models.RoleDoctor {
		s.audit.LogUnauthorizedAccess(createdByUserID, fmt.Sprintf("medical_record:create:patient_%d", req.PatientID), ipAddress, userAgent, "insufficient_role")
		return nil, fmt.Errorf("insufficient permissions to create medical record")
	}

	// Verify patient exists
	var patient models.Patient
	if err := s.db.Where("id = ?", req.PatientID).First(&patient).Error; err != nil {
		return nil, fmt.Errorf("patient not found")
	}

	// Create medical record
	record := models.MedicalRecord{
		PatientID:   req.PatientID,
		DoctorID:    createdByUserID,
		Diagnosis:   req.Diagnosis,
		Treatment:   req.Treatment,
		Notes:       req.Notes,
		Medications: req.Medications,
		Severity:    req.Severity,
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("failed to create medical record: %w", err)
	}

	// Log medical record creation
	s.audit.LogMedicalRecordAccess(createdByUserID, req.PatientID, record.ID, models.ActionCreate, ipAddress, userAgent, false, "medical_record_created")

	return &record, nil
}

// GetMedicalRecord retrieves a medical record by ID
func (s *MedicalRecordService) GetMedicalRecord(recordID uint, requestedByUserID uint, requestedByRole models.UserRole, ipAddress, userAgent string, emergencyAccess bool) (*models.MedicalRecord, error) {
	if !s.canAccessMedicalRecords(requestedByRole) {
		s.audit.LogUnauthorizedAccess(requestedByUserID, fmt.Sprintf("medical_record:%d", recordID), ipAddress, userAgent, "insufficient_role")
		return nil, fmt.Errorf("insufficient permissions to access medical records")
	}

	var record models.MedicalRecord
	if err := s.db.Where("id = ?", recordID).Preload("Patient").Preload("Doctor").First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("medical record not found")
		}
		return nil, fmt.Errorf("failed to retrieve medical record: %w", err)
	}

	// Check role-based access
	if !record.CanBeAccessedByRole(requestedByRole, requestedByUserID) && !emergencyAccess {
		s.audit.LogUnauthorizedAccess(requestedByUserID, fmt.Sprintf("medical_record:%d", recordID), ipAddress, userAgent, "role_access_denied")
		return nil, fmt.Errorf("access denied to medical record")
	}

	// Log medical record access
	reason := ""
	if emergencyAccess {
		reason = "emergency_access"
	}
	s.audit.LogMedicalRecordAccess(requestedByUserID, record.PatientID, recordID, models.ActionView, ipAddress, userAgent, emergencyAccess, reason)

	// Sanitize based on role
	sanitized := record.SanitizeForRole(requestedByRole)
	if sanitized == nil {
		return nil, fmt.Errorf("access denied to medical record")
	}

	return sanitized, nil
}

// GetPatientMedicalRecords retrieves all medical records for a patient
func (s *MedicalRecordService) GetPatientMedicalRecords(patientID uint, requestedByUserID uint, requestedByRole models.UserRole, ipAddress, userAgent string, emergencyAccess bool, page, limit int) ([]models.MedicalRecord, int64, error) {
	if !s.canAccessMedicalRecords(requestedByRole) {
		s.audit.LogUnauthorizedAccess(requestedByUserID, fmt.Sprintf("medical_records:patient_%d", patientID), ipAddress, userAgent, "insufficient_role")
		return nil, 0, fmt.Errorf("insufficient permissions to access medical records")
	}

	var records []models.MedicalRecord
	var total int64

	query := s.db.Where("patient_id = ?", patientID)

	// Apply role-based filtering
	if requestedByRole == models.RoleNurse && !emergencyAccess {
		// Nurses can't see critical records without emergency access
		query = query.Where("severity != ?", models.SeverityCritical)
	}

	query.Model(&models.MedicalRecord{}).Count(&total)

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).
		Preload("Patient").Preload("Doctor").
		Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve medical records: %w", err)
	}

	// Sanitize records based on role
	var sanitizedRecords []models.MedicalRecord
	for _, record := range records {
		if sanitized := record.SanitizeForRole(requestedByRole); sanitized != nil {
			sanitizedRecords = append(sanitizedRecords, *sanitized)
		}
	}

	// Log access
	reason := ""
	if emergencyAccess {
		reason = "emergency_access"
	}
	s.audit.LogPatientAccess(requestedByUserID, patientID, models.ActionView, ipAddress, userAgent, emergencyAccess, reason)

	return sanitizedRecords, total, nil
}

// UpdateMedicalRecord updates a medical record
func (s *MedicalRecordService) UpdateMedicalRecord(recordID uint, req *UpdateMedicalRecordRequest, updatedByUserID uint, updatedByRole models.UserRole, ipAddress, userAgent string) (*models.MedicalRecord, error) {
	var record models.MedicalRecord
	if err := s.db.Where("id = ?", recordID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("medical record not found")
		}
		return nil, fmt.Errorf("failed to retrieve medical record: %w", err)
	}

	// Check permissions - only doctors can update, and typically only the creating doctor
	if updatedByRole != models.RoleDoctor {
		s.audit.LogUnauthorizedAccess(updatedByUserID, fmt.Sprintf("medical_record:%d", recordID), ipAddress, userAgent, "insufficient_role")
		return nil, fmt.Errorf("insufficient permissions to update medical record")
	}

	// Additional check: only the creating doctor or any doctor with emergency access can update
	if record.DoctorID != updatedByUserID {
		s.audit.LogUnauthorizedAccess(updatedByUserID, fmt.Sprintf("medical_record:%d", recordID), ipAddress, userAgent, "not_creating_doctor")
		return nil, fmt.Errorf("only the creating doctor can update this medical record")
	}

	// Build update map
	updates := make(map[string]interface{})
	
	if req.Diagnosis != nil {
		updates["diagnosis"] = *req.Diagnosis
	}
	if req.Treatment != nil {
		updates["treatment"] = *req.Treatment
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}
	if req.Medications != nil {
		updates["medications"] = *req.Medications
	}
	if req.Severity != nil {
		updates["severity"] = *req.Severity
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.db.Model(&record).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update medical record: %w", err)
		}
	}

	// Log update
	s.audit.LogMedicalRecordAccess(updatedByUserID, record.PatientID, recordID, models.ActionUpdate, ipAddress, userAgent, false, "medical_record_updated")

	// Reload record
	s.db.Where("id = ?", recordID).Preload("Patient").Preload("Doctor").First(&record)

	return &record, nil
}

// Helper methods
func (s *MedicalRecordService) canAccessMedicalRecords(role models.UserRole) bool {
	return role == models.RoleDoctor || role == models.RoleNurse
}