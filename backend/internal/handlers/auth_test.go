package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"healthsecure/internal/auth"
	"healthsecure/internal/models"
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

func setupAuthHandler() (*AuthHandler, *MockUserService, *MockAuditService) {
	gin.SetMode(gin.TestMode)
	jwtService := auth.NewJWTService("test-secret", 15*time.Minute, 24*time.Hour)
	userService := &MockUserService{}
	auditService := &MockAuditService{}
	
	handler := &AuthHandler{
		UserService:  userService,
		JWTService:   jwtService,
		AuditService: auditService,
	}
	
	return handler, userService, auditService
}

func TestAuthHandler_Login(t *testing.T) {
	handler, userService, auditService := setupAuthHandler()

	t.Run("SuccessfulLogin", func(t *testing.T) {
		user := &models.User{
			ID:     1,
			Email:  "doctor@example.com",
			Role:   models.RoleDoctor,
			Name:   "Test Doctor",
			Active: true,
		}

		loginReq := LoginRequest{
			Email:    "doctor@example.com",
			Password: "password123",
		}

		userService.On("ValidateCredentials", loginReq.Email, loginReq.Password).Return(user, nil)
		userService.On("UpdateLastLogin", user.ID).Return(nil)
		auditService.On("LogAction", mock.AnythingOfType("*models.AuditLog")).Return(nil)

		router := gin.New()
		router.POST("/login", handler.Login)

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, user.ID, response.User.ID)
		assert.Equal(t, user.Email, response.User.Email)

		userService.AssertExpectations(t)
		auditService.AssertExpectations(t)
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		loginReq := LoginRequest{
			Email:    "doctor@example.com",
			Password: "wrongpassword",
		}

		userService.On("ValidateCredentials", loginReq.Email, loginReq.Password).Return(nil, auth.ErrInvalidCredentials)
		auditService.On("LogAction", mock.AnythingOfType("*models.AuditLog")).Return(nil)

		router := gin.New()
		router.POST("/login", handler.Login)

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		userService.AssertExpectations(t)
		auditService.AssertExpectations(t)
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
		loginReq := LoginRequest{
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
	handler, userService, auditService := setupAuthHandler()

	t.Run("SuccessfulRefresh", func(t *testing.T) {
		user := &models.User{
			ID:     1,
			Email:  "doctor@example.com",
			Role:   models.RoleDoctor,
			Name:   "Test Doctor",
			Active: true,
		}

		// Generate a valid refresh token
		refreshToken, err := handler.JWTService.GenerateRefreshToken(user)
		require.NoError(t, err)

		refreshReq := RefreshTokenRequest{
			RefreshToken: refreshToken,
		}

		userService.On("GetByEmail", user.Email).Return(user, nil)

		router := gin.New()
		router.POST("/refresh", handler.RefreshToken)

		body, _ := json.Marshal(refreshReq)
		req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response RefreshTokenResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)

		userService.AssertExpectations(t)
	})

	t.Run("InvalidRefreshToken", func(t *testing.T) {
		refreshReq := RefreshTokenRequest{
			RefreshToken: "invalid-token",
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
	handler, _, auditService := setupAuthHandler()

	user := &models.User{
		ID:     1,
		Email:  "doctor@example.com",
		Role:   models.RoleDoctor,
		Name:   "Test Doctor",
		Active: true,
	}

	token, err := handler.JWTService.GenerateAccessToken(user)
	require.NoError(t, err)

	auditService.On("LogAction", mock.AnythingOfType("*models.AuditLog")).Return(nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Mock auth middleware - set user in context
		c.Set("user", user)
		c.Set("token", token)
		c.Next()
	})
	router.POST("/logout", handler.Logout)

	req := httptest.NewRequest("POST", "/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	auditService.AssertExpectations(t)
}

func TestAuthHandler_ChangePassword(t *testing.T) {
	handler, userService, auditService := setupAuthHandler()

	user := &models.User{
		ID:     1,
		Email:  "doctor@example.com",
		Role:   models.RoleDoctor,
		Name:   "Test Doctor",
		Active: true,
	}

	t.Run("SuccessfulPasswordChange", func(t *testing.T) {
		changeReq := ChangePasswordRequest{
			CurrentPassword: "oldpassword",
			NewPassword:     "NewPassword123",
		}

		userService.On("ChangePassword", user.ID, changeReq.CurrentPassword, changeReq.NewPassword).Return(nil)
		auditService.On("LogAction", mock.AnythingOfType("*models.AuditLog")).Return(nil)

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user", user)
			c.Next()
		})
		router.POST("/change-password", handler.ChangePassword)

		body, _ := json.Marshal(changeReq)
		req := httptest.NewRequest("POST", "/change-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		userService.AssertExpectations(t)
		auditService.AssertExpectations(t)
	})

	t.Run("WeakPassword", func(t *testing.T) {
		changeReq := ChangePasswordRequest{
			CurrentPassword: "oldpassword",
			NewPassword:     "weak",
		}

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user", user)
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