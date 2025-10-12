package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"healthsecure/configs"
	"healthsecure/internal/auth"
	"healthsecure/internal/database"
	"healthsecure/internal/models"
)

func setupTestGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func setupTestDB(t *testing.T) *configs.Config {
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

	return config
}

func TestAuthMiddleware(t *testing.T) {
	config := setupTestDB(t)
	defer database.Close()

	jwtService := auth.NewJWTService(config)

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleDoctor,
		Name:  "Test Doctor",
	}

	t.Run("ValidToken", func(t *testing.T) {
		tokens, err := jwtService.GenerateTokens(user)
		require.NoError(t, err)

		router := setupTestGin()
		router.Use(auth.AuthMiddleware(jwtService))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("MissingToken", func(t *testing.T) {
		router := setupTestGin()
		router.Use(auth.AuthMiddleware(jwtService))
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
		router.Use(auth.AuthMiddleware(jwtService))
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
	config := setupTestDB(t)
	defer database.Close()

	jwtService := auth.NewJWTService(config)

	t.Run("RequireAdmin", func(t *testing.T) {
		adminUser := &models.User{
			ID:    1,
			Email: "admin@test.com",
			Name:  "Admin User",
			Role:  models.RoleAdmin,
		}
		doctorUser := &models.User{
			ID:    2,
			Email: "doctor@test.com",
			Name:  "Doctor User",
			Role:  models.RoleDoctor,
		}

		adminTokens, err := jwtService.GenerateTokens(adminUser)
		require.NoError(t, err)

		doctorTokens, err := jwtService.GenerateTokens(doctorUser)
		require.NoError(t, err)

		router := setupTestGin()
		router.Use(auth.AuthMiddleware(jwtService))
		router.Use(auth.RequireRole(models.RoleAdmin))
		router.GET("/admin", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "admin access"})
		})

		// Admin should have access
		req := httptest.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+adminTokens.AccessToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Doctor should not have access
		req = httptest.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+doctorTokens.AccessToken)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("RequireAnyRole", func(t *testing.T) {
		doctorUser := &models.User{
			ID:    1,
			Email: "doctor@test.com",
			Name:  "Doctor User",
			Role:  models.RoleDoctor,
		}
		nurseUser := &models.User{
			ID:    2,
			Email: "nurse@test.com",
			Name:  "Nurse User",
			Role:  models.RoleNurse,
		}
		adminUser := &models.User{
			ID:    3,
			Email: "admin@test.com",
			Name:  "Admin User",
			Role:  models.RoleAdmin,
		}

		doctorTokens, _ := jwtService.GenerateTokens(doctorUser)
		nurseTokens, _ := jwtService.GenerateTokens(nurseUser)
		adminTokens, _ := jwtService.GenerateTokens(adminUser)

		router := setupTestGin()
		router.Use(auth.AuthMiddleware(jwtService))
		router.Use(auth.RequireRole(models.RoleDoctor, models.RoleNurse))
		router.GET("/medical", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "medical access"})
		})

		// Doctor should have access
		req := httptest.NewRequest("GET", "/medical", nil)
		req.Header.Set("Authorization", "Bearer "+doctorTokens.AccessToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Nurse should have access
		req = httptest.NewRequest("GET", "/medical", nil)
		req.Header.Set("Authorization", "Bearer "+nurseTokens.AccessToken)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Admin should not have access
		req = httptest.NewRequest("GET", "/medical", nil)
		req.Header.Set("Authorization", "Bearer "+adminTokens.AccessToken)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}