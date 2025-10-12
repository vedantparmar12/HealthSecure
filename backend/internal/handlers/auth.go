package handlers

import (
	"net/http"
	"strconv"

	"healthsecure/internal/auth"
	"healthsecure/internal/models"
	"healthsecure/internal/services"

	"github.com/gin-gonic/gin"
)

type SignupRequest struct {
	Email    string           `json:"email" binding:"required,email"`
	Password string           `json:"password" binding:"required,min=8"`
	Name     string           `json:"name" binding:"required"`
	Role     models.UserRole  `json:"role" binding:"required,oneof=doctor nurse"`
}

type AuthHandler struct {
	userService  *services.UserService
	oauthService *auth.OAuthService
	jwtService   *auth.JWTService
}

func NewAuthHandler(userService *services.UserService, oauthService *auth.OAuthService, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		userService:  userService,
		oauthService: oauthService,
		jwtService:   jwtService,
	}
}

// Login handles user authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	tokens, err := h.userService.Login(&req, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Login successful",
		"token":         tokens.AccessToken,
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_at":    tokens.ExpiresAt,
		"user":          tokens.User,
	})
}

// Signup handles user registration
func (h *AuthHandler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to CreateUserRequest
	createReq := &services.CreateUserRequest{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Role:     req.Role,
	}

	// Create user (using admin user ID 1 as creator, or system)
	user, err := h.userService.CreateUser(createReq, 0) // 0 indicates system creation
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Auto-login the new user
	loginReq := &services.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	tokens, err := h.userService.Login(loginReq, ipAddress, userAgent)
	if err != nil {
		// User created successfully but login failed, still return success
		c.JSON(http.StatusCreated, gin.H{
			"message": "Account created successfully. Please log in.",
			"user":    user,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Account created and logged in successfully",
		"token":         tokens.AccessToken,
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_at":    tokens.ExpiresAt,
		"user":          tokens.User,
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Token refreshed successfully",
		"token":         tokens.AccessToken,
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_at":    tokens.ExpiresAt,
		"user":          tokens.User,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := c.GetUint("user_id")
	authHeader := c.GetHeader("Authorization")
	accessToken := auth.ExtractTokenFromHeader(authHeader)
	
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.userService.Logout(userID, accessToken, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	user, err := h.userService.GetUser(userID, userID, userRole)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// UpdateProfile updates the current user's profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateUser(userID, &req, userID, userRole)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user,
	})
}

// ChangePassword handles password changes
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	var req services.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.userService.ChangePassword(userID, &req, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// GetUserSessions returns the current user's active sessions
func (h *AuthHandler) GetUserSessions(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := models.UserRole(c.GetString("user_role"))

	sessions, err := h.userService.GetUserSessions(userID, userID, userRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

// OAuthLogin initiates OAuth login flow
func (h *AuthHandler) OAuthLogin(c *gin.Context) {
	provider := c.Param("provider")
	
	authURL, err := h.oauthService.GenerateAuthURL(auth.OAuthProvider(provider))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"provider": provider,
	})
}

// OAuthCallback handles OAuth callback
func (h *AuthHandler) OAuthCallback(c *gin.Context) {
	tokens, err := h.oauthService.HandleCallback(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "OAuth login successful",
		"token":         tokens.AccessToken,
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_at":    tokens.ExpiresAt,
		"user":          tokens.User,
	})
}

// Additional utility handlers

// ValidateToken validates a token without requiring full authentication
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	token := auth.ExtractTokenFromHeader(authHeader)
	
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":    true,
		"user_id":  claims.UserID,
		"role":     claims.Role,
		"expires":  claims.ExpiresAt,
	})
}

// GetSupportedOAuthProviders returns available OAuth providers
func (h *AuthHandler) GetSupportedOAuthProviders(c *gin.Context) {
	if !h.oauthService.IsConfigured() {
		c.JSON(http.StatusOK, gin.H{
			"oauth_enabled": false,
			"providers":     []string{},
		})
		return
	}

	providers := h.oauthService.GetSupportedProviders()
	c.JSON(http.StatusOK, gin.H{
		"oauth_enabled": true,
		"providers":     providers,
	})
}

// GetDefaultUsers returns default user credentials for each role
func (h *AuthHandler) GetDefaultUsers(c *gin.Context) {
	defaults, err := h.userService.GetDefaultUsersByRole()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch default users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"defaults": defaults})
}

// Helper function to get pagination parameters
func getPaginationParams(c *gin.Context) (page int, limit int) {
	page = 1
	limit = 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	return page, limit
}