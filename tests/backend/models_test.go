package tests

import (
	"testing"
	"time"

	"healthsecure/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// TestUserModel tests the User model functionality
func TestUserModel(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	t.Run("Create User", func(t *testing.T) {
		user := &models.User{
			Email:    "test@example.com",
			Password: "hashedpassword",
			Role:     models.RoleDoctor,
			Name:     "Dr. Test",
			Active:   true,
		}

		err := db.Create(user).Error
		assert.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.NotZero(t, user.CreatedAt)
		assert.NotZero(t, user.UpdatedAt)
	})

	t.Run("User Role Methods", func(t *testing.T) {
		doctor := &models.User{Role: models.RoleDoctor}
		nurse := &models.User{Role: models.RoleNurse}
		admin := &models.User{Role: models.RoleAdmin}

		// Test IsDoctor
		assert.True(t, doctor.IsDoctor())
		assert.False(t, nurse.IsDoctor())
		assert.False(t, admin.IsDoctor())

		// Test IsNurse
		assert.False(t, doctor.IsNurse())
		assert.True(t, nurse.IsNurse())
		assert.False(t, admin.IsNurse())

		// Test IsAdmin
		assert.False(t, doctor.IsAdmin())
		assert.False(t, nurse.IsAdmin())
		assert.True(t, admin.IsAdmin())

		// Test CanAccessPatientData
		assert.True(t, doctor.CanAccessPatientData())
		assert.True(t, nurse.CanAccessPatientData())
		assert.False(t, admin.CanAccessPatientData())

		// Test CanAccessSensitiveData
		assert.True(t, doctor.CanAccessSensitiveData())
		assert.False(t, nurse.CanAccessSensitiveData())
		assert.False(t, admin.CanAccessSensitiveData())

		// Test CanManageUsers
		assert.False(t, doctor.CanManageUsers())
		assert.False(t, nurse.CanManageUsers())
		assert.True(t, admin.CanManageUsers())
	})

	t.Run("Unique Email Constraint", func(t *testing.T) {
		user1 := &models.User{
			Email:    "duplicate@example.com",
			Password: "password1",
			Role:     models.RoleDoctor,
			Name:     "User 1",
		}

		user2 := &models.User{
			Email:    "duplicate@example.com",
			Password: "password2",
			Role:     models.RoleNurse,
			Name:     "User 2",
		}

		// First user should create successfully
		err := db.Create(user1).Error
		assert.NoError(t, err)

		// Second user with same email should fail
		err = db.Create(user2).Error
		assert.Error(t, err)
	})
}

// TestPatientModel tests the Patient model functionality
func TestPatientModel(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	t.Run("Create Patient", func(t *testing.T) {
		patient := &models.Patient{
			FirstName:        "John",
			LastName:         "Doe",
			DateOfBirth:      time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			SSN:              "123-45-6789",
			Phone:            "555-0123",
			Address:          "123 Main St",
			EmergencyContact: "Jane Doe - 555-0124",
		}

		err := db.Create(patient).Error
		assert.NoError(t, err)
		assert.NotZero(t, patient.ID)
		assert.NotZero(t, patient.CreatedAt)
		assert.NotZero(t, patient.UpdatedAt)
	})

	t.Run("Patient with Medical Records", func(t *testing.T) {
		// Create a doctor first
		doctor := &models.User{
			Email:    "doctor@example.com",
			Password: "hashedpassword",
			Role:     models.RoleDoctor,
			Name:     "Dr. Smith",
		}
		db.Create(doctor)

		// Create patient
		patient := &models.Patient{
			FirstName:   "Jane",
			LastName:    "Smith",
			DateOfBirth: time.Date(1985, 6, 15, 0, 0, 0, 0, time.UTC),
		}
		db.Create(patient)

		// Create medical record
		record := &models.MedicalRecord{
			PatientID:   patient.ID,
			DoctorID:    doctor.ID,
			Diagnosis:   "Hypertension",
			Treatment:   "Medication and lifestyle changes",
			Notes:       "Patient responds well to treatment",
			Medications: "Lisinopril 10mg daily",
			Severity:    "medium",
		}
		db.Create(record)

		// Test relationship loading
		var loadedPatient models.Patient
		err := db.Preload("MedicalRecords").First(&loadedPatient, patient.ID).Error
		assert.NoError(t, err)
		assert.Len(t, loadedPatient.MedicalRecords, 1)
		assert.Equal(t, "Hypertension", loadedPatient.MedicalRecords[0].Diagnosis)
	})
}

// TestAuditLog tests the AuditLog model functionality
func TestAuditLog(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	t.Run("Create Audit Log", func(t *testing.T) {
		// Create user first
		user := &models.User{
			Email:    "auditor@example.com",
			Password: "hashedpassword",
			Role:     models.RoleDoctor,
			Name:     "Dr. Auditor",
		}
		db.Create(user)

		auditLog := &models.AuditLog{
			UserID:       user.ID,
			Action:       "VIEW_PATIENT",
			Resource:     "Patient",
			IPAddress:    "192.168.1.1",
			UserAgent:    "Mozilla/5.0 Test Browser",
			EmergencyUse: false,
			Success:      true,
		}

		err := db.Create(auditLog).Error
		assert.NoError(t, err)
		assert.NotZero(t, auditLog.ID)
		assert.NotZero(t, auditLog.Timestamp)
	})

	t.Run("Audit Log with Emergency Use", func(t *testing.T) {
		user := &models.User{
			Email:    "emergency@example.com",
			Password: "hashedpassword",
			Role:     models.RoleNurse,
			Name:     "Emergency Nurse",
		}
		db.Create(user)

		patient := &models.Patient{
			FirstName: "Emergency",
			LastName:  "Patient",
		}
		db.Create(patient)

		auditLog := &models.AuditLog{
			UserID:       user.ID,
			PatientID:    &patient.ID,
			Action:       "EMERGENCY_ACCESS",
			Resource:     "Patient",
			IPAddress:    "192.168.1.100",
			UserAgent:    "Emergency Access App",
			EmergencyUse: true,
			Reason:       "Life-threatening emergency requiring immediate access",
			Success:      true,
		}

		err := db.Create(auditLog).Error
		assert.NoError(t, err)
		assert.True(t, auditLog.EmergencyUse)
		assert.NotEmpty(t, auditLog.Reason)
	})
}

// TestEmergencyAccess tests the EmergencyAccess model functionality
func TestEmergencyAccess(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	t.Run("Create Emergency Access", func(t *testing.T) {
		user := &models.User{
			Email:    "emergency@example.com",
			Password: "hashedpassword",
			Role:     models.RoleNurse,
			Name:     "Emergency Nurse",
		}
		db.Create(user)

		patient := &models.Patient{
			FirstName: "Critical",
			LastName:  "Patient",
		}
		db.Create(patient)

		emergencyAccess := &models.EmergencyAccess{
			UserID:      user.ID,
			PatientID:   patient.ID,
			Reason:      "Patient in critical condition, regular physician unavailable",
			AccessToken: "emergency-token-123456",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		}

		err := db.Create(emergencyAccess).Error
		assert.NoError(t, err)
		assert.NotZero(t, emergencyAccess.ID)
		assert.NotZero(t, emergencyAccess.CreatedAt)
	})

	t.Run("Emergency Access Expiration", func(t *testing.T) {
		user := &models.User{
			Email:    "temp@example.com",
			Password: "hashedpassword",
			Role:     models.RoleDoctor,
			Name:     "Temp Doctor",
		}
		db.Create(user)

		patient := &models.Patient{
			FirstName: "Test",
			LastName:  "Patient",
		}
		db.Create(patient)

		// Create expired emergency access
		expiredAccess := &models.EmergencyAccess{
			UserID:      user.ID,
			PatientID:   patient.ID,
			Reason:      "Test emergency access",
			AccessToken: "expired-token",
			ExpiresAt:   time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		}
		db.Create(expiredAccess)

		// Create valid emergency access
		validAccess := &models.EmergencyAccess{
			UserID:      user.ID,
			PatientID:   patient.ID,
			Reason:      "Valid emergency access",
			AccessToken: "valid-token",
			ExpiresAt:   time.Now().Add(1 * time.Hour), // Expires in 1 hour
		}
		db.Create(validAccess)

		// Query only non-expired access
		var activeAccess []models.EmergencyAccess
		err := db.Where("expires_at > ? AND revoked_at IS NULL", time.Now()).Find(&activeAccess).Error
		assert.NoError(t, err)
		assert.Len(t, activeAccess, 1)
		assert.Equal(t, "valid-token", activeAccess[0].AccessToken)
	})
}

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *gorm.DB {
	// Use an in-memory SQLite database for tests
	// In production, you might want to use a dedicated test MySQL database
	db, err := gorm.Open(mysql.Open("root:@tcp(localhost:3306)/healthsecure_test?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		t.Skipf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate tables for testing
	err = db.AutoMigrate(
		&models.User{},
		&models.Patient{},
		&models.MedicalRecord{},
		&models.AuditLog{},
		&models.EmergencyAccess{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// cleanupTestDB cleans up the test database
func cleanupTestDB(db *gorm.DB) {
	// Clean up test data
	db.Exec("DELETE FROM emergency_accesses")
	db.Exec("DELETE FROM audit_logs")
	db.Exec("DELETE FROM medical_records")
	db.Exec("DELETE FROM patients")
	db.Exec("DELETE FROM users")
}
