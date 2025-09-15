package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"healthsecure/internal/models"
)

func TestJWTService(t *testing.T) {
	jwtService := NewJWTService("test-secret-key-for-testing", 15*time.Minute, 24*time.Hour)

	t.Run("GenerateAccessToken", func(t *testing.T) {
		user := &models.User{
			ID:    1,
			Email: "test@example.com",
			Role:  models.RoleDoctor,
			Name:  "Test Doctor",
		}

		token, err := jwtService.GenerateAccessToken(user)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("GenerateRefreshToken", func(t *testing.T) {
		user := &models.User{
			ID:    1,
			Email: "test@example.com",
		}

		token, err := jwtService.GenerateRefreshToken(user)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("ValidateToken", func(t *testing.T) {
		user := &models.User{
			ID:    1,
			Email: "test@example.com",
			Role:  models.RoleDoctor,
			Name:  "Test Doctor",
		}

		token, err := jwtService.GenerateAccessToken(user)
		require.NoError(t, err)

		claims, err := jwtService.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Email, claims.Email)
		assert.Equal(t, string(user.Role), claims.Role)
	})

	t.Run("ValidateExpiredToken", func(t *testing.T) {
		shortJWT := NewJWTService("test-secret", 1*time.Millisecond, 1*time.Millisecond)
		user := &models.User{
			ID:    1,
			Email: "test@example.com",
			Role:  models.RoleDoctor,
		}

		token, err := shortJWT.GenerateAccessToken(user)
		require.NoError(t, err)

		time.Sleep(2 * time.Millisecond)

		_, err = shortJWT.ValidateToken(token)
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
		}

		token, err := jwtService.GenerateAccessToken(user)
		require.NoError(t, err)

		// Token should be valid initially
		_, err = jwtService.ValidateToken(token)
		require.NoError(t, err)

		// Blacklist the token
		err = jwtService.BlacklistToken(token)
		require.NoError(t, err)

		// Token should now be invalid
		_, err = jwtService.ValidateToken(token)
		assert.Error(t, err)
	})

	t.Run("ValidatePassword", func(t *testing.T) {
		tests := []struct {
			password string
			valid    bool
		}{
			{"short", false},
			{"nouppercase1", false},
			{"NOLOWERCASE1", false},
			{"NoNumbers", false},
			{"ValidPassword123", true},
			{"AnotherValid1", true},
		}

		for _, test := range tests {
			err := ValidatePassword(test.password)
			if test.valid {
				assert.NoError(t, err, "Password %s should be valid", test.password)
			} else {
				assert.Error(t, err, "Password %s should be invalid", test.password)
			}
		}
	})

	t.Run("HashAndCheckPassword", func(t *testing.T) {
		password := "TestPassword123"
		
		hashedPassword, err := HashPassword(password)
		require.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)
		assert.NotEqual(t, password, hashedPassword)

		// Valid password should match
		valid := CheckPasswordHash(password, hashedPassword)
		assert.True(t, valid)

		// Invalid password should not match
		valid = CheckPasswordHash("WrongPassword", hashedPassword)
		assert.False(t, valid)
	})
}