package auth

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"healthsecure/configs"
	"healthsecure/internal/database"
	"healthsecure/internal/models"
)

func setupTestDB() (*configs.Config, error) {
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

	// Skip migrations for faster tests
	os.Setenv("SKIP_MIGRATIONS", "false")

	// Initialize test database
	if err := database.Initialize(config); err != nil {
		return nil, err
	}

	return config, nil
}

func TestJWTService(t *testing.T) {
	config, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer database.Close()

	jwtService := NewJWTService(config)

	t.Run("GenerateTokens", func(t *testing.T) {
		user := &models.User{
			ID:    1,
			Email: "test@example.com",
			Role:  models.RoleDoctor,
			Name:  "Test Doctor",
		}

		tokens, err := jwtService.GenerateTokens(user)
		require.NoError(t, err)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)
		assert.NotNil(t, tokens.User)
	})

	t.Run("GenerateRefreshToken", func(t *testing.T) {
		user := &models.User{
			ID:    1,
			Email: "test@example.com",
			Role:  models.RoleDoctor,
			Name:  "Test User",
		}

		tokens, err := jwtService.GenerateTokens(user)
		require.NoError(t, err)
		assert.NotEmpty(t, tokens.RefreshToken)
	})

	t.Run("ValidateToken", func(t *testing.T) {
		user := &models.User{
			ID:    1,
			Email: "test@example.com",
			Role:  models.RoleDoctor,
			Name:  "Test Doctor",
		}

		tokens, err := jwtService.GenerateTokens(user)
		require.NoError(t, err)

		claims, err := jwtService.ValidateToken(tokens.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Email, claims.Email)
		assert.Equal(t, user.Role, claims.Role)
	})

	t.Run("ValidateExpiredToken", func(t *testing.T) {
		shortConfig := &configs.Config{
			JWT: configs.JWTConfig{
				Secret:              "test-secret-key-for-testing-minimum-32-chars",
				Expires:             1 * time.Millisecond,
				RefreshTokenExpires: 1 * time.Millisecond,
			},
			Security: configs.SecurityConfig{
				BCryptCost: 10,
			},
		}
		shortJWT := NewJWTService(shortConfig)
		user := &models.User{
			ID:    1,
			Email: "test@example.com",
			Role:  models.RoleDoctor,
			Name:  "Test User",
		}

		tokens, err := shortJWT.GenerateTokens(user)
		require.NoError(t, err)

		time.Sleep(2 * time.Millisecond)

		_, err = shortJWT.ValidateToken(tokens.AccessToken)
		assert.Error(t, err)
	})

	t.Run("ValidateInvalidToken", func(t *testing.T) {
		_, err := jwtService.ValidateToken("invalid-token")
		assert.Error(t, err)
	})

	t.Run("BlacklistToken", func(t *testing.T) {
		user := &models.User{
			ID:    1,
			Email: "test@example.com",
			Role:  models.RoleDoctor,
			Name:  "Test User",
		}

		tokens, err := jwtService.GenerateTokens(user)
		require.NoError(t, err)

		// Token should be valid initially
		_, err = jwtService.ValidateToken(tokens.AccessToken)
		require.NoError(t, err)

		// Blacklist the token
		err = jwtService.BlacklistToken(tokens.AccessToken)
		require.NoError(t, err)

		// Token should now be invalid
		_, err = jwtService.ValidateToken(tokens.AccessToken)
		assert.Error(t, err)
	})

	t.Run("ValidatePasswordStrength", func(t *testing.T) {
		tests := []struct {
			password string
			valid    bool
		}{
			{"short", false},
			{"nouppercase1!", false},
			{"NOLOWERCASE1!", false},
			{"NoNumbers!", false},
			{"NoSpecial123", false},
			{"ValidPassword123!", true},
			{"AnotherValid1@", true},
		}

		for _, test := range tests {
			err := ValidatePasswordStrength(test.password)
			if test.valid {
				assert.NoError(t, err, "Password %s should be valid", test.password)
			} else {
				assert.Error(t, err, "Password %s should be invalid", test.password)
			}
		}
	})

	t.Run("HashAndCheckPassword", func(t *testing.T) {
		password := "TestPassword123!"

		hashedPassword, err := jwtService.HashPassword(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)
		assert.NotEqual(t, password, hashedPassword)

		// Valid password should match
		valid := jwtService.CheckPasswordHash(password, hashedPassword)
		assert.True(t, valid)

		// Invalid password should not match
		valid = jwtService.CheckPasswordHash("WrongPassword", hashedPassword)
		assert.False(t, valid)
	})
}