package auth

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"

	"healthsecure/configs"
	"healthsecure/internal/database"
	"healthsecure/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID   uint             `json:"user_id"`
	Email    string           `json:"email"`
	Role     models.UserRole  `json:"role"`
	TokenID  string           `json:"token_id"`
	Type     TokenType        `json:"type"`
	jwt.RegisteredClaims
}

type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         *models.User `json:"user"`
}

type JWTService struct {
	config *configs.Config
}

func NewJWTService(config *configs.Config) *JWTService {
	return &JWTService{
		config: config,
	}
}

// GenerateTokens creates both access and refresh tokens for a user
func (j *JWTService) GenerateTokens(user *models.User) (*AuthResponse, error) {
	now := time.Now()
	accessTokenID := uuid.New().String()
	refreshTokenID := uuid.New().String()

	// Create access token claims
	accessClaims := &Claims{
		UserID:  user.ID,
		Email:   user.Email,
		Role:    user.Role,
		TokenID: accessTokenID,
		Type:    AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.JWT.Expires)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "healthsecure",
			Subject:   strconv.Itoa(int(user.ID)),
			ID:        accessTokenID,
		},
	}

	// Create refresh token claims
	refreshClaims := &Claims{
		UserID:  user.ID,
		Email:   user.Email,
		Role:    user.Role,
		TokenID: refreshTokenID,
		Type:    RefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.JWT.RefreshTokenExpires)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "healthsecure",
			Subject:   strconv.Itoa(int(user.ID)),
			ID:        refreshTokenID,
		},
	}

	// Generate tokens
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	// Sign tokens
	accessTokenString, err := accessToken.SignedString([]byte(j.config.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshTokenString, err := refreshToken.SignedString([]byte(j.config.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Sanitize user data for response
	userResponse := *user
	userResponse.Password = ""

	return &AuthResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessClaims.RegisteredClaims.ExpiresAt.Time,
		User:         &userResponse,
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if token is blacklisted
	if isBlacklisted, err := j.IsTokenBlacklisted(tokenString); err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	} else if isBlacklisted {
		return nil, fmt.Errorf("token is blacklisted")
	}

	// Verify token hasn't expired
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
func (j *JWTService) RefreshAccessToken(refreshTokenString string) (*AuthResponse, error) {
	// Validate refresh token
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.Type != RefreshToken {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	// Get user from database to ensure they're still active
	var user models.User
	if err := database.GetDB().Where("id = ? AND active = ?", claims.UserID, true).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found or inactive: %w", err)
	}

	// Generate new tokens (both access and refresh for security)
	tokens, err := j.GenerateTokens(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	// Blacklist the old refresh token
	if err := j.BlacklistToken(refreshTokenString); err != nil {
		// Log the error but don't fail the refresh process
		fmt.Printf("Warning: failed to blacklist old refresh token: %v", err)
	}

	return tokens, nil
}

// BlacklistToken adds a token to the blacklist
func (j *JWTService) BlacklistToken(tokenString string) error {
	// Parse token to get expiration
	claims, err := j.parseTokenWithoutValidation(tokenString)
	if err != nil {
		return fmt.Errorf("failed to parse token for blacklisting: %w", err)
	}

	// Create hash of token for storage
	hash := sha256.Sum256([]byte(tokenString))
	tokenHash := fmt.Sprintf("%x", hash)

	// Store in blacklist
	blacklistedToken := database.BlacklistedToken{
		TokenHash: tokenHash,
		UserID:    claims.UserID,
		ExpiresAt: claims.ExpiresAt.Time,
	}

	if err := database.GetDB().Create(&blacklistedToken).Error; err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

// IsTokenBlacklisted checks if a token is in the blacklist
func (j *JWTService) IsTokenBlacklisted(tokenString string) (bool, error) {
	// Create hash of token
	hash := sha256.Sum256([]byte(tokenString))
	tokenHash := fmt.Sprintf("%x", hash)

	var count int64
	err := database.GetDB().Model(&database.BlacklistedToken{}).
		Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check blacklist: %w", err)
	}

	return count > 0, nil
}

// parseTokenWithoutValidation parses a token without validating its signature or expiration
func (j *JWTService) parseTokenWithoutValidation(tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// HashPassword securely hashes a password using bcrypt
func (j *JWTService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), j.config.Security.BCryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a password with a hash
func (j *JWTService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// CreateUserSession creates a new user session record
func (j *JWTService) CreateUserSession(userID uint, sessionID, ipAddress, userAgent string) error {
	session := database.UserSession{
		UserID:    userID,
		SessionID: sessionID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(j.config.JWT.RefreshTokenExpires),
	}

	return database.GetDB().Create(&session).Error
}

// InvalidateUserSessions invalidates all sessions for a user
func (j *JWTService) InvalidateUserSessions(userID uint) error {
	return database.GetDB().Where("user_id = ?", userID).Delete(&database.UserSession{}).Error
}

// GetActiveUserSessions returns active sessions for a user
func (j *JWTService) GetActiveUserSessions(userID uint) ([]database.UserSession, error) {
	var sessions []database.UserSession
	err := database.GetDB().
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("last_activity DESC").
		Find(&sessions).Error

	return sessions, err
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}

// ValidatePasswordStrength checks if password meets security requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case char == '!' || char == '@' || char == '#' || char == '$' || char == '%' || 
		     char == '^' || char == '&' || char == '*' || char == '(' || char == ')' ||
		     char == '-' || char == '_' || char == '+' || char == '=':
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}