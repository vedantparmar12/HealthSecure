package services

import (
	"fmt"
	"time"

	"healthsecure/internal/models"

	"gorm.io/gorm"
)

type PatientService struct {
	db    *gorm.DB
	audit *AuditService
}

type CreatePatientRequest struct {
	FirstName        string    `json:"first_name" binding:"required"`
	LastName         string    `json:"last_name" binding:"required"`
	DateOfBirth      time.Time `json:"date_of_birth" binding:"required"`
	SSN              string    `json:"ssn" binding:"required"`
	Phone            string    `json:"phone"`
	Address          string    `json:"address"`
	EmergencyContact string    `json:"emergency_contact"`
}

type UpdatePatientRequest struct {
	FirstName        *string    `json:"first_name,omitempty"`
	LastName         *string    `json:"last_name,omitempty"`
	DateOfBirth      *time.Time `json:"date_of_birth,omitempty"`
	Phone            *string    `json:"phone,omitempty"`
	Address          *string    `json:"address,omitempty"`
	EmergencyContact *string    `json:"emergency_contact,omitempty"`
}

type PatientSearchQuery struct {
	FirstName   string    `form:"first_name"`
	LastName    string    `form:"last_name"`
	SSN         string    `form:"ssn"`
	Phone       string    `form:"phone"`
	DateOfBirth time.Time `form:"date_of_birth"`
	Page        int       `form:"page,default=1"`
	Limit       int       `form:"limit,default=20"`
}

func NewPatientService(db *gorm.DB, audit *AuditService) *PatientService {
	return &PatientService{
		db:    db,
		audit: audit,
	}
}

// CreatePatient creates a new patient record
func (s *PatientService) CreatePatient(req *CreatePatientRequest, createdByUserID uint, createdByRole models.UserRole, ipAddress, userAgent string) (*models.Patient, error) {
	// Only doctors and admins can create patients
	if createdByRole != models.RoleDoctor && createdByRole != models.RoleAdmin {
		s.audit.LogUnauthorizedAccess(createdByUserID, "patients", ipAddress, userAgent, "insufficient_role_for_creation")
		return nil, fmt.Errorf("insufficient permissions to create patient")
	}

	// Check if patient with SSN already exists
	var existingPatient models.Patient
	if err := s.db.Where("ssn = ?", req.SSN).First(&existingPatient).Error; err == nil {
		return nil, fmt.Errorf("patient with SSN already exists")
	}

	// Create patient
	patient := models.Patient{
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		DateOfBirth:      req.DateOfBirth,
		SSN:              req.SSN,
		Phone:            req.Phone,
		Address:          req.Address,
		EmergencyContact: req.EmergencyContact,
	}

	if err := s.db.Create(&patient).Error; err != nil {
		return nil, fmt.Errorf("failed to create patient: %w", err)
	}

	// Log patient creation
	s.audit.LogPatientAccess(createdByUserID, patient.ID, models.ActionCreate, ipAddress, userAgent, false, "patient_created")

	return &patient, nil
}

// GetPatient retrieves a patient by ID with role-based data filtering
func (s *PatientService) GetPatient(patientID uint, requestedByUserID uint, requestedByRole models.UserRole, ipAddress, userAgent string, emergencyAccess bool) (*models.Patient, error) {
	// Check if user has permission to access patient data
	if !s.canAccessPatientData(requestedByRole) {
		s.audit.LogUnauthorizedAccess(requestedByUserID, fmt.Sprintf("patient:%d", patientID), ipAddress, userAgent, "insufficient_role")
		return nil, fmt.Errorf("insufficient permissions to access patient data")
	}

	var patient models.Patient
	if err := s.db.Where("id = ?", patientID).First(&patient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("patient not found")
		}
		return nil, fmt.Errorf("failed to retrieve patient: %w", err)
	}

	// Log patient access
	reason := ""
	if emergencyAccess {
		reason = "emergency_access"
	}
	s.audit.LogPatientAccess(requestedByUserID, patientID, models.ActionView, ipAddress, userAgent, emergencyAccess, reason)

	// Apply role-based filtering
	sanitizedPatient := patient.SanitizeForRole(requestedByRole)
	if sanitizedPatient == nil {
		s.audit.LogUnauthorizedAccess(requestedByUserID, fmt.Sprintf("patient:%d", patientID), ipAddress, userAgent, "role_based_filtering_denied")
		return nil, fmt.Errorf("access denied to patient data")
	}

	return sanitizedPatient, nil
}

// GetPatients retrieves patients with filtering, pagination, and role-based access control
func (s *PatientService) GetPatients(query *PatientSearchQuery, requestedByUserID uint, requestedByRole models.UserRole, ipAddress, userAgent string) ([]models.Patient, int64, error) {
	// Check if user has permission to access patient data
	if !s.canAccessPatientData(requestedByRole) {
		s.audit.LogUnauthorizedAccess(requestedByUserID, "patients", ipAddress, userAgent, "insufficient_role")
		return nil, 0, fmt.Errorf("insufficient permissions to access patient data")
	}

	var patients []models.Patient
	var total int64

	// Build query
	dbQuery := s.db.Model(&models.Patient{})

	// Apply search filters
	if query.FirstName != "" {
		dbQuery = dbQuery.Where("first_name LIKE ?", "%"+query.FirstName+"%")
	}
	if query.LastName != "" {
		dbQuery = dbQuery.Where("last_name LIKE ?", "%"+query.LastName+"%")
	}
	if query.Phone != "" {
		dbQuery = dbQuery.Where("phone LIKE ?", "%"+query.Phone+"%")
	}
	if !query.DateOfBirth.IsZero() {
		dbQuery = dbQuery.Where("date_of_birth = ?", query.DateOfBirth)
	}

	// Only doctors can search by SSN
	if query.SSN != "" && requestedByRole == models.RoleDoctor {
		dbQuery = dbQuery.Where("ssn = ?", query.SSN)
	} else if query.SSN != "" {
		s.audit.LogUnauthorizedAccess(requestedByUserID, "patients", ipAddress, userAgent, "ssn_search_denied")
		return nil, 0, fmt.Errorf("insufficient permissions to search by SSN")
	}

	// Count total results
	dbQuery.Count(&total)

	// Apply pagination
	offset := (query.Page - 1) * query.Limit
	if err := dbQuery.Offset(offset).Limit(query.Limit).Find(&patients).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve patients: %w", err)
	}

	// Apply role-based filtering to each patient
	var filteredPatients []models.Patient
	for _, patient := range patients {
		sanitized := patient.SanitizeForRole(requestedByRole)
		if sanitized != nil {
			filteredPatients = append(filteredPatients, *sanitized)
		}
	}

	// Log patients list access
	s.audit.LogUserAction(requestedByUserID, models.ActionView, "patients_list", ipAddress, userAgent, true, fmt.Sprintf("returned_%d_patients", len(filteredPatients)))

	return filteredPatients, total, nil
}

// UpdatePatient updates patient information with role-based access control
func (s *PatientService) UpdatePatient(patientID uint, req *UpdatePatientRequest, updatedByUserID uint, updatedByRole models.UserRole, ipAddress, userAgent string) (*models.Patient, error) {
	// Only doctors and nurses can update patients (different permissions)
	if !s.canUpdatePatientData(updatedByRole) {
		s.audit.LogUnauthorizedAccess(updatedByUserID, fmt.Sprintf("patient:%d", patientID), ipAddress, userAgent, "insufficient_role_for_update")
		return nil, fmt.Errorf("insufficient permissions to update patient")
	}

	var patient models.Patient
	if err := s.db.Where("id = ?", patientID).First(&patient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("patient not found")
		}
		return nil, fmt.Errorf("failed to retrieve patient: %w", err)
	}

	// Build update map
	updates := make(map[string]interface{})
	
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.DateOfBirth != nil {
		updates["date_of_birth"] = *req.DateOfBirth
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.EmergencyContact != nil {
		updates["emergency_contact"] = *req.EmergencyContact
	}

	// Apply updates if any
	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.db.Model(&patient).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update patient: %w", err)
		}
	}

	// Log patient update
	s.audit.LogPatientAccess(updatedByUserID, patientID, models.ActionUpdate, ipAddress, userAgent, false, "patient_updated")

	// Reload patient data
	s.db.Where("id = ?", patientID).First(&patient)

	// Apply role-based filtering before returning
	sanitizedPatient := patient.SanitizeForRole(updatedByRole)
	if sanitizedPatient == nil {
		return nil, fmt.Errorf("access denied to updated patient data")
	}

	return sanitizedPatient, nil
}

// DeletePatient soft deletes a patient (admin only)
func (s *PatientService) DeletePatient(patientID uint, deletedByUserID uint, deletedByRole models.UserRole, ipAddress, userAgent string) error {
	// Only admins can delete patients
	if deletedByRole != models.RoleAdmin {
		s.audit.LogUnauthorizedAccess(deletedByUserID, fmt.Sprintf("patient:%d", patientID), ipAddress, userAgent, "insufficient_role_for_deletion")
		return fmt.Errorf("insufficient permissions to delete patient")
	}

	var patient models.Patient
	if err := s.db.Where("id = ?", patientID).First(&patient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("patient not found")
		}
		return fmt.Errorf("failed to retrieve patient: %w", err)
	}

	// Soft delete the patient
	if err := s.db.Delete(&patient).Error; err != nil {
		return fmt.Errorf("failed to delete patient: %w", err)
	}

	// Log patient deletion
	s.audit.LogPatientAccess(deletedByUserID, patientID, models.ActionDelete, ipAddress, userAgent, false, "patient_deleted")

	return nil
}

// GetPatientWithMedicalRecords retrieves a patient with their medical records
func (s *PatientService) GetPatientWithMedicalRecords(patientID uint, requestedByUserID uint, requestedByRole models.UserRole, ipAddress, userAgent string, emergencyAccess bool) (*models.Patient, error) {
	// Check permissions
	if !s.canAccessPatientData(requestedByRole) {
		s.audit.LogUnauthorizedAccess(requestedByUserID, fmt.Sprintf("patient:%d", patientID), ipAddress, userAgent, "insufficient_role")
		return nil, fmt.Errorf("insufficient permissions to access patient data")
	}

	var patient models.Patient
	query := s.db.Where("id = ?", patientID).Preload("MedicalRecords")

	// Nurses can't see critical medical records unless emergency access
	if requestedByRole == models.RoleNurse && !emergencyAccess {
		query = query.Preload("MedicalRecords", "severity != ?", models.SeverityCritical)
	}

	if err := query.First(&patient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("patient not found")
		}
		return nil, fmt.Errorf("failed to retrieve patient with records: %w", err)
	}

	// Log access to patient and medical records
	reason := ""
	if emergencyAccess {
		reason = "emergency_access"
	}
	s.audit.LogPatientAccess(requestedByUserID, patientID, models.ActionView, ipAddress, userAgent, emergencyAccess, reason)

	// Sanitize medical records based on role
	var sanitizedRecords []models.MedicalRecord
	for _, record := range patient.MedicalRecords {
		if sanitized := record.SanitizeForRole(requestedByRole); sanitized != nil {
			sanitizedRecords = append(sanitizedRecords, *sanitized)
		}
	}
	patient.MedicalRecords = sanitizedRecords

	// Apply role-based filtering to patient data
	sanitizedPatient := patient.SanitizeForRole(requestedByRole)
	if sanitizedPatient == nil {
		s.audit.LogUnauthorizedAccess(requestedByUserID, fmt.Sprintf("patient:%d", patientID), ipAddress, userAgent, "role_based_filtering_denied")
		return nil, fmt.Errorf("access denied to patient data")
	}

	return sanitizedPatient, nil
}

// SearchPatientsByName searches patients by name with fuzzy matching
func (s *PatientService) SearchPatientsByName(name string, requestedByUserID uint, requestedByRole models.UserRole, ipAddress, userAgent string, limit int) ([]models.Patient, error) {
	if !s.canAccessPatientData(requestedByRole) {
		s.audit.LogUnauthorizedAccess(requestedByUserID, "patients_search", ipAddress, userAgent, "insufficient_role")
		return nil, fmt.Errorf("insufficient permissions to search patients")
	}

	var patients []models.Patient
	searchTerm := "%" + name + "%"

	err := s.db.Where("first_name LIKE ? OR last_name LIKE ? OR CONCAT(first_name, ' ', last_name) LIKE ?", 
		searchTerm, searchTerm, searchTerm).
		Limit(limit).Find(&patients).Error

	if err != nil {
		return nil, fmt.Errorf("failed to search patients: %w", err)
	}

	// Apply role-based filtering
	var filteredPatients []models.Patient
	for _, patient := range patients {
		if sanitized := patient.SanitizeForRole(requestedByRole); sanitized != nil {
			filteredPatients = append(filteredPatients, *sanitized)
		}
	}

	// Log search
	s.audit.LogUserAction(requestedByUserID, models.ActionView, "patients_search", ipAddress, userAgent, true, fmt.Sprintf("searched_name:%s", name))

	return filteredPatients, nil
}

// GetPatientStatistics returns patient statistics (admin only)
func (s *PatientService) GetPatientStatistics(requestedByRole models.UserRole) (map[string]interface{}, error) {
	if requestedByRole != models.RoleAdmin {
		return nil, fmt.Errorf("insufficient permissions to view patient statistics")
	}

	stats := make(map[string]interface{})

	// Total patients
	var totalPatients int64
	s.db.Model(&models.Patient{}).Count(&totalPatients)
	stats["total_patients"] = totalPatients

	// Patients added this month
	monthStart := time.Now().AddDate(0, -1, 0)
	var patientsThisMonth int64
	s.db.Model(&models.Patient{}).Where("created_at >= ?", monthStart).Count(&patientsThisMonth)
	stats["patients_this_month"] = patientsThisMonth

	// Age distribution
	var ageDistribution []struct {
		AgeGroup string `json:"age_group"`
		Count    int64  `json:"count"`
	}

	// This is a simplified age distribution query
	// In a real implementation, you would calculate ages more accurately
	s.db.Raw(`
		SELECT 
			CASE 
				WHEN YEAR(CURDATE()) - YEAR(date_of_birth) < 18 THEN 'Under 18'
				WHEN YEAR(CURDATE()) - YEAR(date_of_birth) BETWEEN 18 AND 30 THEN '18-30'
				WHEN YEAR(CURDATE()) - YEAR(date_of_birth) BETWEEN 31 AND 50 THEN '31-50'
				WHEN YEAR(CURDATE()) - YEAR(date_of_birth) BETWEEN 51 AND 70 THEN '51-70'
				ELSE 'Over 70'
			END as age_group,
			COUNT(*) as count
		FROM patients 
		GROUP BY age_group
	`).Scan(&ageDistribution)

	stats["age_distribution"] = ageDistribution

	return stats, nil
}

// Helper methods for role-based access control
func (s *PatientService) canAccessPatientData(role models.UserRole) bool {
	return role == models.RoleDoctor || role == models.RoleNurse
}

func (s *PatientService) canUpdatePatientData(role models.UserRole) bool {
	return role == models.RoleDoctor || role == models.RoleNurse
}

func (s *PatientService) canCreatePatientData(role models.UserRole) bool {
	return role == models.RoleDoctor || role == models.RoleAdmin
}

func (s *PatientService) canDeletePatientData(role models.UserRole) bool {
	return role == models.RoleAdmin
}

func (s *PatientService) canAccessSensitiveData(role models.UserRole) bool {
	return role == models.RoleDoctor
}

// GetAllPatients retrieves all patients from the database
func (s *PatientService) GetAllPatients() ([]models.Patient, error) {
	var patients []models.Patient
	if err := s.db.Find(&patients).Error; err != nil {
		return nil, err
	}
	return patients, nil
}