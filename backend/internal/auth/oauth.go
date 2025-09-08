package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"healthsecure/configs"
	"healthsecure/internal/database"
	"healthsecure/internal/models"

	"github.com/gin-gonic/gin"
)

const (
	OAuthStateTimeout = 15 * time.Minute
)

type OAuthProvider string

const (
	GoogleProvider    OAuthProvider = "google"
	MicrosoftProvider OAuthProvider = "microsoft"
	GitHubProvider    OAuthProvider = "github"
)

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
}

type OAuthUserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	Provider string `json:"provider"`
}

type OAuthState struct {
	State     string    `json:"state"`
	Provider  string    `json:"provider"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type OAuthService struct {
	config    *configs.Config
	providers map[OAuthProvider]*OAuthConfig
	states    map[string]*OAuthState // In production, use Redis or database
}

func NewOAuthService(config *configs.Config) *OAuthService {
	service := &OAuthService{
		config: config,
		providers: make(map[OAuthProvider]*OAuthConfig),
		states: make(map[string]*OAuthState),
	}

	// Initialize OAuth providers
	service.initializeProviders()

	// Start cleanup routine for expired states
	go service.cleanupExpiredStates()

	return service
}

func (o *OAuthService) initializeProviders() {
	// Google OAuth configuration
	o.providers[GoogleProvider] = &OAuthConfig{
		ClientID:     o.config.OAuth.ClientID,
		ClientSecret: o.config.OAuth.ClientSecret,
		RedirectURL:  o.config.OAuth.RedirectURL,
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
		Scopes:       []string{"openid", "email", "profile"},
	}

	// Microsoft OAuth configuration
	o.providers[MicrosoftProvider] = &OAuthConfig{
		ClientID:     o.config.OAuth.ClientID,
		ClientSecret: o.config.OAuth.ClientSecret,
		RedirectURL:  o.config.OAuth.RedirectURL,
		AuthURL:      "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
		TokenURL:     "https://login.microsoftonline.com/common/oauth2/v2.0/token",
		UserInfoURL:  "https://graph.microsoft.com/v1.0/me",
		Scopes:       []string{"openid", "email", "profile"},
	}
}

// GenerateAuthURL creates an OAuth authorization URL with state parameter
func (o *OAuthService) GenerateAuthURL(provider OAuthProvider) (string, error) {
	config, exists := o.providers[provider]
	if !exists {
		return "", fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	// Generate secure random state
	state, err := o.generateSecureState()
	if err != nil {
		return "", fmt.Errorf("failed to generate OAuth state: %w", err)
	}

	// Store state for validation
	o.states[state] = &OAuthState{
		State:     state,
		Provider:  string(provider),
		ExpiresAt: time.Now().Add(OAuthStateTimeout),
		CreatedAt: time.Now(),
	}

	// Build authorization URL
	params := url.Values{
		"client_id":     {config.ClientID},
		"redirect_uri":  {config.RedirectURL},
		"response_type": {"code"},
		"scope":         {fmt.Sprintf("%v", config.Scopes)},
		"state":         {state},
		"access_type":   {"offline"}, // For Google to get refresh token
		"prompt":        {"consent"},  // Force consent screen
	}

	authURL := config.AuthURL + "?" + params.Encode()
	return authURL, nil
}

// HandleCallback processes the OAuth callback and returns user tokens
func (o *OAuthService) HandleCallback(c *gin.Context) (*AuthResponse, error) {
	// Extract parameters from callback
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")

	// Check for OAuth errors
	if errorParam != "" {
		return nil, fmt.Errorf("OAuth error: %s", errorParam)
	}

	if code == "" {
		return nil, fmt.Errorf("authorization code is required")
	}

	if state == "" {
		return nil, fmt.Errorf("state parameter is required")
	}

	// Validate state parameter
	storedState, exists := o.states[state]
	if !exists {
		return nil, fmt.Errorf("invalid or expired state parameter")
	}

	// Check if state has expired
	if storedState.ExpiresAt.Before(time.Now()) {
		delete(o.states, state)
		return nil, fmt.Errorf("state parameter has expired")
	}

	// Clean up used state
	delete(o.states, state)

	// Get provider configuration
	provider := OAuthProvider(storedState.Provider)
	config, exists := o.providers[provider]
	if !exists {
		return nil, fmt.Errorf("unsupported provider in state: %s", provider)
	}

	// Exchange authorization code for access token
	accessToken, err := o.exchangeCodeForToken(config, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Get user information
	userInfo, err := o.getUserInfo(config, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	userInfo.Provider = string(provider)

	// Find or create user in database
	user, err := o.findOrCreateUser(userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create user: %w", err)
	}

	// Generate JWT tokens
	jwtService := NewJWTService(o.config)
	tokens, err := jwtService.GenerateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update last login time
	now := time.Now()
	user.LastLogin = &now
	database.GetDB().Save(user)

	return tokens, nil
}

// exchangeCodeForToken exchanges authorization code for access token
func (o *OAuthService) exchangeCodeForToken(config *OAuthConfig, code string) (string, error) {
	// Prepare token exchange request
	params := url.Values{
		"client_id":     {config.ClientID},
		"client_secret": {config.ClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {config.RedirectURL},
	}

	// Make request to token endpoint
	resp, err := http.PostForm(config.TokenURL, params)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}

	// Parse token response
	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}

	if err := parseJSONResponse(resp, &tokenResponse); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResponse.AccessToken == "" {
		return "", fmt.Errorf("no access token in response")
	}

	return tokenResponse.AccessToken, nil
}

// getUserInfo retrieves user information using access token
func (o *OAuthService) getUserInfo(config *OAuthConfig, accessToken string) (*OAuthUserInfo, error) {
	// Create request with authorization header
	req, err := http.NewRequest("GET", config.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	// Make request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("user info request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status: %d", resp.StatusCode)
	}

	// Parse user info response
	var userInfo OAuthUserInfo
	if err := parseJSONResponse(resp, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info response: %w", err)
	}

	if userInfo.Email == "" {
		return nil, fmt.Errorf("no email in user info response")
	}

	return &userInfo, nil
}

// findOrCreateUser finds existing user or creates new one based on OAuth info
func (o *OAuthService) findOrCreateUser(userInfo *OAuthUserInfo) (*models.User, error) {
	var user models.User

	// Try to find existing user by email
	err := database.GetDB().Where("email = ?", userInfo.Email).First(&user).Error
	if err == nil {
		// User exists, update last login and return
		if !user.Active {
			return nil, fmt.Errorf("user account is deactivated")
		}
		return &user, nil
	}

	// User doesn't exist, create new one
	// In a real application, you might want to require admin approval for new OAuth users
	user = models.User{
		Email:     userInfo.Email,
		Name:      userInfo.Name,
		Role:      models.RoleNurse, // Default role, should be configured per organization
		Active:    true,
		Password:  "", // OAuth users don't need password
		LastLogin: nil,
	}

	if err := database.GetDB().Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// generateSecureState generates a cryptographically secure random state parameter
func (o *OAuthService) generateSecureState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// cleanupExpiredStates removes expired OAuth states
func (o *OAuthService) cleanupExpiredStates() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			for state, stateInfo := range o.states {
				if stateInfo.ExpiresAt.Before(now) {
					delete(o.states, state)
				}
			}
		}
	}
}

// GetSupportedProviders returns list of configured OAuth providers
func (o *OAuthService) GetSupportedProviders() []string {
	var providers []string
	for provider := range o.providers {
		providers = append(providers, string(provider))
	}
	return providers
}

// IsConfigured checks if OAuth is properly configured
func (o *OAuthService) IsConfigured() bool {
	return o.config.OAuth.ClientID != "" && 
		   o.config.OAuth.ClientSecret != "" && 
		   o.config.OAuth.RedirectURL != ""
}

// Helper function to parse JSON responses
func parseJSONResponse(resp *http.Response, v interface{}) error {
	// This would typically use json.NewDecoder(resp.Body).Decode(v)
	// Simplified for this example
	return nil
}