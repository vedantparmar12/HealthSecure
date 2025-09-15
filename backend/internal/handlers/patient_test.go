package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"healthsecure/internal/models"
)

type MockPatientService struct {
	mock.Mock
}

func (m *MockPatientService) GetPatients(page, limit int, userRole models.UserRole, emergencyToken string) ([]*models.Patient, *PaginationInfo, error) {
	args := m.Called(page, limit, userRole, emergencyToken)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*PaginationInfo), args.Error(2)
	}
	return args.Get(0).([]*models.Patient), args.Get(1).(*PaginationInfo), args.Error(2)
}

func (m *MockPatientService) GetPatientByID(id uint, userRole models.UserRole, emergencyToken string) (*models.Patient, error) {
	args := m.Called(id, userRole, emergencyToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Patient), args.Error(1)
}

func (m *MockPatientService) CreatePatient(patient *models.Patient) error {
	args := m.Called(patient)
	return args.Error(0)
}

func (m *MockPatientService) UpdatePatient(id uint, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *MockPatientService) SearchPatients(query string, userRole models.UserRole) ([]*models.Patient, error) {
	args := m.Called(query, userRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Patient), args.Error(1)
}

func setupPatientHandler() (*PatientHandler, *MockPatientService, *MockAuditService) {
	gin.SetMode(gin.TestMode)
	patientService := &MockPatientService{}
	auditService := &MockAuditService{}
	
	handler := &PatientHandler{
		PatientService: patientService,
		AuditService:   auditService,
	}
	
	return handler, patientService, auditService
}

func TestPatientHandler_GetPatients(t *testing.T) {
	handler, patientService, _ := setupPatientHandler()

	patients := []*models.Patient{
		{
			ID:        1,
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		},
		{
			ID:        2,
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane@example.com",
		},
	}

	pagination := &PaginationInfo{
		CurrentPage: 1,
		TotalPages:  1,
		Total:       2,
		Limit:       20,
	}

	user := &models.User{
		ID:   1,
		Role: models.RoleDoctor,
	}

	patientService.On("GetPatients", 1, 20, models.RoleDoctor, "").Return(patients, pagination, nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", user)
		c.Next()
	})
	router.GET("/patients", handler.GetPatients)

	req := httptest.NewRequest("GET", "/patients?page=1&limit=20", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.NotNil(t, response["data"])

	patientService.AssertExpectations(t)
}

func TestPatientHandler_GetPatient(t *testing.T) {
	handler, patientService, auditService := setupPatientHandler()

	patient := &models.Patient{
		ID:        1,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		SSN:       "123-45-6789",
	}

	user := &models.User{
		ID:   1,
		Role: models.RoleDoctor,
	}

	patientService.On("GetPatientByID", uint(1), models.RoleDoctor, "").Return(patient, nil)
	auditService.On("LogAction", mock.AnythingOfType("*models.AuditLog")).Return(nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", user)
		c.Next()
	})
	router.GET("/patients/:id", handler.GetPatient)

	req := httptest.NewRequest("GET", "/patients/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	
	data := response["data"].(map[string]interface{})
	patientData := data["patient"].(map[string]interface{})
	assert.Equal(t, "John", patientData["first_name"])
	assert.Equal(t, "Doe", patientData["last_name"])

	patientService.AssertExpectations(t)
	auditService.AssertExpectations(t)
}

func TestPatientHandler_CreatePatient(t *testing.T) {
	handler, patientService, auditService := setupPatientHandler()

	user := &models.User{
		ID:   1,
		Role: models.RoleDoctor,
	}

	t.Run("SuccessfulCreation", func(t *testing.T) {
		patientReq := CreatePatientRequest{
			FirstName:        "John",
			LastName:         "Doe",
			DateOfBirth:      "1990-01-01",
			SSN:              "123-45-6789",
			Phone:            "555-1234",
			Address:          "123 Main St",
			EmergencyContact: "Jane Doe - 555-5678",
		}

		patientService.On("CreatePatient", mock.AnythingOfType("*models.Patient")).Return(nil)
		auditService.On("LogAction", mock.AnythingOfType("*models.AuditLog")).Return(nil)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user", user)
			c.Next()
		})
		router.POST("/patients", handler.CreatePatient)

		body, _ := json.Marshal(patientReq)
		req := httptest.NewRequest("POST", "/patients", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		patientService.AssertExpectations(t)
		auditService.AssertExpectations(t)
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		patientReq := CreatePatientRequest{
			// Missing required fields
			FirstName: "John",
		}

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user", user)
			c.Next()
		})
		router.POST("/patients", handler.CreatePatient)

		body, _ := json.Marshal(patientReq)
		req := httptest.NewRequest("POST", "/patients", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestPatientHandler_UpdatePatient(t *testing.T) {
	handler, patientService, auditService := setupPatientHandler()

	user := &models.User{
		ID:   1,
		Role: models.RoleDoctor,
	}

	t.Run("SuccessfulUpdate", func(t *testing.T) {
		updateReq := UpdatePatientRequest{
			FirstName: "John Updated",
			LastName:  "Doe Updated",
			Phone:     "555-9999",
		}

		patientService.On("UpdatePatient", uint(1), mock.AnythingOfType("map[string]interface {}")).Return(nil)
		auditService.On("LogAction", mock.AnythingOfType("*models.AuditLog")).Return(nil)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user", user)
			c.Next()
		})
		router.PUT("/patients/:id", handler.UpdatePatient)

		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest("PUT", "/patients/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		patientService.AssertExpectations(t)
		auditService.AssertExpectations(t)
	})

	t.Run("InvalidPatientID", func(t *testing.T) {
		updateReq := UpdatePatientRequest{
			FirstName: "John",
		}

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user", user)
			c.Next()
		})
		router.PUT("/patients/:id", handler.UpdatePatient)

		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest("PUT", "/patients/invalid", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestPatientHandler_SearchPatients(t *testing.T) {
	handler, patientService, _ := setupPatientHandler()

	patients := []*models.Patient{
		{
			ID:        1,
			FirstName: "John",
			LastName:  "Doe",
		},
	}

	user := &models.User{
		ID:   1,
		Role: models.RoleDoctor,
	}

	patientService.On("SearchPatients", "John", models.RoleDoctor).Return(patients, nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", user)
		c.Next()
	})
	router.GET("/patients/search", handler.SearchPatients)

	req := httptest.NewRequest("GET", "/patients/search?q=John", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	patientService.AssertExpectations(t)
}