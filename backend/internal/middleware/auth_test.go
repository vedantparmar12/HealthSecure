package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"healthsecure/internal/auth"
	"healthsecure/internal/models"
)

func setupTestGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestAuthMiddleware(t *testing.T) {
	jwtService := auth.NewJWTService("test-secret", 15*time.Minute, 24*time.Hour)
	middleware := NewAuthMiddleware(jwtService, nil)

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleDoctor,
		Name:  "Test Doctor",
	}

	t.Run("ValidToken", func(t *testing.T) {
		token, err := jwtService.GenerateAccessToken(user)
		require.NoError(t, err)

		router := setupTestGin()
		router.Use(middleware.RequireAuth())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("MissingToken", func(t *testing.T) {
		router := setupTestGin()
		router.Use(middleware.RequireAuth())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		router := setupTestGin()
		router.Use(middleware.RequireAuth())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRoleMiddleware(t *testing.T) {
	jwtService := auth.NewJWTService("test-secret", 15*time.Minute, 24*time.Hour)
	middleware := NewAuthMiddleware(jwtService, nil)

	t.Run("RequireAdmin", func(t *testing.T) {
		adminUser := &models.User{
			ID:   1,
			Role: models.RoleAdmin,
		}
		doctorUser := &models.User{
			ID:   2,
			Role: models.RoleDoctor,
		}

		adminToken, err := jwtService.GenerateAccessToken(adminUser)
		require.NoError(t, err)

		doctorToken, err := jwtService.GenerateAccessToken(doctorUser)
		require.NoError(t, err)

		router := setupTestGin()
		router.Use(middleware.RequireAuth())
		router.Use(middleware.RequireRole(models.RoleAdmin))
		router.GET("/admin", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "admin access"})
		})

		// Admin should have access
		req := httptest.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Doctor should not have access
		req = httptest.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+doctorToken)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("RequireAnyRole", func(t *testing.T) {
		doctorUser := &models.User{
			ID:   1,
			Role: models.RoleDoctor,
		}
		nurseUser := &models.User{
			ID:   2,
			Role: models.RoleNurse,
		}
		adminUser := &models.User{
			ID:   3,
			Role: models.RoleAdmin,
		}

		doctorToken, _ := jwtService.GenerateAccessToken(doctorUser)
		nurseToken, _ := jwtService.GenerateAccessToken(nurseUser)
		adminToken, _ := jwtService.GenerateAccessToken(adminUser)

		router := setupTestGin()
		router.Use(middleware.RequireAuth())
		router.Use(middleware.RequireAnyRole(models.RoleDoctor, models.RoleNurse))
		router.GET("/medical", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "medical access"})
		})

		// Doctor should have access
		req := httptest.NewRequest("GET", "/medical", nil)
		req.Header.Set("Authorization", "Bearer "+doctorToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Nurse should have access
		req = httptest.NewRequest("GET", "/medical", nil)
		req.Header.Set("Authorization", "Bearer "+nurseToken)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Admin should not have access
		req = httptest.NewRequest("GET", "/medical", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestAuditMiddleware(t *testing.T) {
	jwtService := auth.NewJWTService("test-secret", 15*time.Minute, 24*time.Hour)
	
	// Mock audit service
	auditLogs := []models.AuditLog{}
	mockAuditService := &MockAuditService{
		logs: &auditLogs,
	}
	
	middleware := NewAuthMiddleware(jwtService, mockAuditService)

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleDoctor,
		Name:  "Test Doctor",
	}

	token, err := jwtService.GenerateAccessToken(user)
	require.NoError(t, err)

	router := setupTestGin()
	router.Use(middleware.RequireAuth())
	router.Use(middleware.AuditLog("TEST_ACTION"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, auditLogs, 1)
	assert.Equal(t, "TEST_ACTION", auditLogs[0].Action)
	assert.Equal(t, user.ID, auditLogs[0].UserID)
	assert.True(t, auditLogs[0].Success)
}

type MockAuditService struct {
	logs *[]models.AuditLog
}

func (m *MockAuditService) LogAction(log *models.AuditLog) error {
	*m.logs = append(*m.logs, *log)
	return nil
}