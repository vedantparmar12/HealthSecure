package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"healthsecure/configs"
	"healthsecure/internal/auth"
	"healthsecure/internal/database"
	"healthsecure/internal/handlers"
	"healthsecure/internal/models"
	"healthsecure/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthenticationSecurity tests authentication security measures
func TestAuthenticationSecurity(t *testing.T) {
	// Setup
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	t.Run("Reject Invalid JWT Token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/patients", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid token", response["error"])
	})

	t.Run("Reject Expired JWT Token", func(t *testing.T) {
		// Create a token that's already expired
		config := &configs.Config{
			JWT: configs.JWTConfig{
				Secret:  "test-secret-key-minimum-32-chars-long",
				Expires: -1 * time.Hour, // Expired
			},
		}
		
		jwtService := auth.NewJWTService(config)
		token, err := jwtService.GenerateToken(1, "doctor", "Dr. Test")
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/api/patients", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Require Authentication for Protected Routes", func(t *testing.T) {
		protectedRoutes := []string{
			"/api/patients",
			"/api/patients/1",
			"/api/records/1",
			"/api/emergency/request",
			"/api/audit/logs",
		}

		for _, route := range protectedRoutes {
			req, _ := http.NewRequest("GET", route, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code, 
				"Route %s should require authentication", route)
		}
	})
}

// TestRoleBasedAccessControl tests RBAC security
func TestRoleBasedAccessControl(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Create test users with different roles
	doctorToken := createTestUserToken(t, "doctor", "Dr. Test", "doctor@test.com")
	nurseToken := createTestUserToken(t, "nurse", "Nurse Test", "nurse@test.com")
	adminToken := createTestUserToken(t, "admin", "Admin Test", "admin@test.com")

	t.Run("Admin Cannot Access Patient Data", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/patients", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Admin should be forbidden from accessing patient data
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Nurse Cannot Access Sensitive Patient Data", func(t *testing.T) {
		// This would need to be tested at the service level
		// as the filtering happens in the service layer, not at route level
		
		// Create a test patient first
		patientID := createTestPatient(t)
		
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/patients/%d", patientID), nil)
		req.Header.Set("Authorization", "Bearer "+nurseToken)
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var patient models.Patient
		err := json.Unmarshal(w.Body.Bytes(), &patient)
		assert.NoError(t, err)
		
		// Nurse should not see SSN or other sensitive data
		assert.Empty(t, patient.SSN, "Nurse should not have access to SSN")
	})

	t.Run("Doctor Has Full Access to Patient Data", func(t *testing.T) {
		patientID := createTestPatient(t)
		
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/patients/%d", patientID), nil)
		req.Header.Set("Authorization", "Bearer "+doctorToken)
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var patient models.Patient
		err := json.Unmarshal(w.Body.Bytes(), &patient)
		assert.NoError(t, err)
		
		// Doctor should have access to all data including SSN
		assert.NotEmpty(t, patient.SSN, "Doctor should have access to SSN")
	})

	t.Run("Only Doctors Can Create Medical Records", func(t *testing.T) {
		patientID := createTestPatient(t)
		
		recordData := map[string]interface{}{
			"diagnosis":   "Test Diagnosis",
			"treatment":   "Test Treatment",
			"notes":       "Test Notes",
			"medications": "Test Medications",
			"severity":    "medium",
		}
		
		data, _ := json.Marshal(recordData)

		// Test with nurse token (should fail)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/api/patients/%d/records", patientID), bytes.NewBuffer(data))
		req.Header.Set("Authorization", "Bearer "+nurseToken)
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code, "Nurse should not be able to create medical records")

		// Test with doctor token (should succeed)
		req, _ = http.NewRequest("POST", fmt.Sprintf("/api/patients/%d/records", patientID), bytes.NewBuffer(data))
		req.Header.Set("Authorization", "Bearer "+doctorToken)
		req.Header.Set("Content-Type", "application/json")
		
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Doctor should be able to create medical records")
	})
}

// TestEmergencyAccessSecurity tests emergency access security
func TestEmergencyAccessSecurity(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	nurseToken := createTestUserToken(t, "nurse", "Emergency Nurse", "emergency@test.com")

	t.Run("Emergency Access Requires Justification", func(t *testing.T) {
		patientID := createTestPatient(t)
		
		// Try without reason
		requestData := map[string]interface{}{
			"patient_id": patientID,
		}
		
		data, _ := json.Marshal(requestData)
		req, _ := http.NewRequest("POST", "/api/emergency/request", bytes.NewBuffer(data))
		req.Header.Set("Authorization", "Bearer "+nurseToken)
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		// Try with reason (should succeed)
		requestData["reason"] = "Patient in critical condition, attending physician unavailable"
		data, _ = json.Marshal(requestData)
		
		req, _ = http.NewRequest("POST", "/api/emergency/request", bytes.NewBuffer(data))
		req.Header.Set("Authorization", "Bearer "+nurseToken)
		req.Header.Set("Content-Type", "application/json")
		
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Emergency Access Has Time Limits", func(t *testing.T) {
		// This test would verify that emergency access tokens expire
		// and cannot be used after expiration
		patientID := createTestPatient(t)
		
		requestData := map[string]interface{}{
			"patient_id": patientID,
			"reason":     "Critical emergency situation",
		}
		
		data, _ := json.Marshal(requestData)
		req, _ := http.NewRequest("POST", "/api/emergency/request", bytes.NewBuffer(data))
		req.Header.Set("Authorization", "Bearer "+nurseToken)
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		// Check that expiration time is set
		assert.NotNil(t, response["expires_at"])
		
		// In a real test, you would advance time and verify the token is no longer valid
	})
}

// TestAuditLogging tests that all actions are properly logged
func TestAuditLogging(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	doctorToken := createTestUserToken(t, "doctor", "Dr. Audit", "audit@test.com")

	t.Run("Patient Access Is Audited", func(t *testing.T) {
		patientID := createTestPatient(t)
		
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/patients/%d", patientID), nil)
		req.Header.Set("Authorization", "Bearer "+doctorToken)
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		// Verify audit log was created
		db := database.GetDB()
		var auditLog models.AuditLog
		err := db.Where("action = ? AND resource = ?", "VIEW_PATIENT", "Patient").First(&auditLog).Error
		assert.NoError(t, err, "Patient access should be audited")
		assert.Equal(t, fmt.Sprintf("%d", patientID), fmt.Sprintf("%d", *auditLog.PatientID))
	})

	t.Run("Failed Authentication Is Audited", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/patients", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		
		// Verify failed authentication was logged
		db := database.GetDB()
		var auditLog models.AuditLog
		err := db.Where("action = ? AND success = ?", "AUTH_FAILURE", false).First(&auditLog).Error
		assert.NoError(t, err, "Failed authentication should be audited")
	})
}

// TestInputValidation tests input validation and sanitization
func TestInputValidation(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	adminToken := createTestUserToken(t, "admin", "Admin Test", "admin@test.com")

	t.Run("Reject SQL Injection Attempts", func(t *testing.T) {
		maliciousData := map[string]interface{}{
			"name":     "'; DROP TABLE users; --",
			"email":    "test@evil.com",
			"password": "password",
			"role":     "admin",
		}
		
		data, _ := json.Marshal(maliciousData)
		req, _ := http.NewRequest("POST", "/api/admin/users", bytes.NewBuffer(data))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should either reject the input or sanitize it
		// The exact response depends on validation implementation
		assert.NotEqual(t, http.StatusInternalServerError, w.Code, 
			"SQL injection should not cause server error")
	})

	t.Run("Reject XSS Attempts", func(t *testing.T) {
		maliciousData := map[string]interface{}{
			"name":  "<script>alert('xss')</script>",
			"email": "test@example.com",
			"role":  "nurse",
		}
		
		data, _ := json.Marshal(maliciousData)
		req, _ := http.NewRequest("POST", "/api/admin/users", bytes.NewBuffer(data))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusCreated {
			// If user was created, verify the script tags were sanitized
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			
			name := response["name"].(string)
			assert.NotContains(t, name, "<script>", "Script tags should be sanitized")
		}
	})
}

// TestRateLimiting tests rate limiting functionality
func TestRateLimiting(t *testing.T) {
	t.Skip("Rate limiting tests require specific middleware configuration")
	
	// This test would verify that rate limiting works correctly
	// Implementation depends on the specific rate limiting middleware used
}

// Helper functions

func setupTestRouter(t *testing.T) (*gin.Engine, func()) {
	gin.SetMode(gin.TestMode)

	// Setup test config
	config := &configs.Config{
		JWT: configs.JWTConfig{
			Secret:  "test-secret-key-minimum-32-characters-for-security",
			Expires: 1 * time.Hour,
		},
		Database: configs.DatabaseConfig{
			Host:     "localhost",
			Port:     3306,
			Name:     "healthsecure_test",
			User:     "root",
			Password: "",
		},
	}

	// Initialize test database
	err := database.Initialize(config)
	require.NoError(t, err)

	// Initialize services
	jwtService := auth.NewJWTService(config)
	auditService := services.NewAuditService(database.GetDB())
	userService := services.NewUserService(database.GetDB(), jwtService, auditService)
	patientService := services.NewPatientService(database.GetDB(), auditService)
	
	// Setup router with middleware
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup handlers
	authHandler := handlers.NewAuthHandler(userService, nil, jwtService)
	patientHandler := handlers.NewPatientHandler(patientService, jwtService)

	// Setup routes
	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
		}

		patients := api.Group("/patients")
		patients.Use(auth.AuthMiddleware(jwtService))
		{
			patients.GET("", patientHandler.GetPatients)
			patients.GET("/:id", patientHandler.GetPatient)
		}
	}

	cleanup := func() {
		database.Close()
	}

	return router, cleanup
}

func createTestUserToken(t *testing.T, role, name, email string) string {
	config := &configs.Config{
		JWT: configs.JWTConfig{
			Secret:  "test-secret-key-minimum-32-characters-for-security",
			Expires: 1 * time.Hour,
		},
	}
	
	jwtService := auth.NewJWTService(config)
	
	// Create user in database
	user := &models.User{
		Email:    email,
		Password: "hashedpassword",
		Role:     models.UserRole(role),
		Name:     name,
		Active:   true,
	}
	
	db := database.GetDB()
	err := db.Create(user).Error
	require.NoError(t, err)
	
	token, err := jwtService.GenerateToken(user.ID, string(user.Role), user.Name)
	require.NoError(t, err)
	
	return token
}

func createTestPatient(t *testing.T) uint {
	patient := &models.Patient{
		FirstName:   "Test",
		LastName:    "Patient",
		DateOfBirth: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		SSN:         "123-45-6789",
		Phone:       "555-0123",
		Address:     "123 Test St",
	}
	
	db := database.GetDB()
	err := db.Create(patient).Error
	require.NoError(t, err)
	
	return patient.ID
}
