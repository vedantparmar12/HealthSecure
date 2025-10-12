package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"healthsecure/configs"
	"healthsecure/internal/auth"
	"healthsecure/internal/database"
	"healthsecure/internal/models"
	"healthsecure/internal/services"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) ValidateCredentials(email, password string) (*models.User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateLastLogin(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserService) ChangePassword(userID uint, currentPassword, newPassword string) error {
	args := m.Called(userID, currentPassword, newPassword)
	return args.Error(0)
}

type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) LogAction(log *models.AuditLog) error {
	args := m.Called(log)
	return args.Error(0)
}

func setupAuthHandler(t *testing.T) (*AuthHandler, *MockUserService, *MockAuditService, *auth.JWTService) {
	gin.SetMode(gin.TestMode)

	config := &configs.Config{
		Database: configs.DatabaseConfig{
			Host:     "localhost",
			Port:     3306,
			Name:     "test_db",
			User:     "test",
			Password: "",
			TLSMode:  "preferred",
		},
		JWT: configs.JWTConfig{
			Secret:              "test-secret-key-for-testing-minimum-32-chars",
			Expires:             15 * time.Minute,
			RefreshTokenExpires: 24 * time.Hour,
		},
		Security: configs.SecurityConfig{
			BCryptCost: 10,
		},
		App: configs.AppConfig{
			Environment: "test",
		},
	}

	os.Setenv("SKIP_MIGRATIONS", "false")
	if err := database.Initialize(config); err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	jwtService := auth.NewJWTService(config)
	mockUserService := &MockUserService{}
	mockAuditService := &MockAuditService{}
	oauthService := auth.NewOAuthService(config)

	realUserService := services.NewUserService(database.GetDB(), jwtService, nil)

	handler := NewAuthHandler(realUserService, oauthService, jwtService)

	return handler, mockUserService, mockAuditService, jwtService
}

func TestAuthHandler_Login(t *testing.T) {
	handler, _, _, _ := setupAuthHandler(t)
	defer database.Close()

	t.Run("SuccessfulLogin", func(t *testing.T) {
		loginReq := services.LoginRequest{
			Email:    "dr.smith@hospital.local",
			Password: "Doctor123",
		}

		router := gin.New()
		router.POST("/login", handler.Login)

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response["access_token"])
		assert.NotEmpty(t, response["refresh_token"])
		assert.NotNil(t, response["user"])
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		loginReq := services.LoginRequest{
			Email:    "doctor@example.com",
			Password: "wrongpassword",
		}

		router := gin.New()
		router.POST("/login", handler.Login)

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		router := gin.New()
		router.POST("/login", handler.Login)

		req := httptest.NewRequest("POST", "/login", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("MissingFields", func(t *testing.T) {
		loginReq := services.LoginRequest{
			Email: "doctor@example.com",
			// Missing password
		}

		router := gin.New()
		router.POST("/login", handler.Login)

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	handler, _, _, jwtService := setupAuthHandler(t)
	defer database.Close()

	t.Run("SuccessfulRefresh", func(t *testing.T) {
		user := &models.User{
			ID:     1,
			Email:  "doctor@example.com",
			Role:   models.RoleDoctor,
			Name:   "Test Doctor",
			Active: true,
		}

		// Create user in database
		database.GetDB().Create(user)

		// Generate valid tokens
		tokens, err := jwtService.GenerateTokens(user)
		require.NoError(t, err)

		refreshReq := map[string]string{
			"refresh_token": tokens.RefreshToken,
		}

		router := gin.New()
		router.POST("/refresh", handler.RefreshToken)

		body, _ := json.Marshal(refreshReq)
		req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response["access_token"])
		assert.NotEmpty(t, response["refresh_token"])
	})

	t.Run("InvalidRefreshToken", func(t *testing.T) {
		refreshReq := map[string]string{
			"refresh_token": "invalid-token",
		}

		router := gin.New()
		router.POST("/refresh", handler.RefreshToken)

		body, _ := json.Marshal(refreshReq)
		req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	handler, _, _, jwtService := setupAuthHandler(t)
	defer database.Close()

	user := &models.User{
		ID:     1,
		Email:  "doctor@example.com",
		Role:   models.RoleDoctor,
		Name:   "Test Doctor",
		Active: true,
	}

	tokens, err := jwtService.GenerateTokens(user)
	require.NoError(t, err)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Mock auth middleware - set user in context
		c.Set("user_id", user.ID)
		c.Set("user_role", string(user.Role))
		c.Next()
	})
	router.POST("/logout", handler.Logout)

	req := httptest.NewRequest("POST", "/logout", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_ChangePassword(t *testing.T) {
	handler, _, _, jwtService := setupAuthHandler(t)
	defer database.Close()

	user := &models.User{
		Email:  "testuser@example.com",
		Role:   models.RoleDoctor,
		Name:   "Test Doctor",
		Active: true,
	}

	// Hash the current password
	hashedPassword, _ := jwtService.HashPassword("CurrentPassword123!")
	user.Password = hashedPassword
	database.GetDB().Create(user)

	t.Run("SuccessfulPasswordChange", func(t *testing.T) {
		changeReq := services.ChangePasswordRequest{
			CurrentPassword: "CurrentPassword123!",
			NewPassword:     "NewPassword123!",
		}

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", user.ID)
			c.Next()
		})
		router.POST("/change-password", handler.ChangePassword)

		body, _ := json.Marshal(changeReq)
		req := httptest.NewRequest("POST", "/change-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("WeakPassword", func(t *testing.T) {
		changeReq := services.ChangePasswordRequest{
			CurrentPassword: "CurrentPassword123!",
			NewPassword:     "weak",
		}

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", user.ID)
			c.Next()
		})
		router.POST("/change-password", handler.ChangePassword)

		body, _ := json.Marshal(changeReq)
		req := httptest.NewRequest("POST", "/change-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}