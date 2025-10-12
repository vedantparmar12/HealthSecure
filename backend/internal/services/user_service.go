package services

import (
	"fmt"
	"time"

	"healthsecure/internal/auth"
	"healthsecure/internal/database"
	"healthsecure/internal/models"

	"gorm.io/gorm"
)

type UserService struct {
	db         *gorm.DB
	jwtService *auth.JWTService
	audit      *AuditService
}

type CreateUserRequest struct {
	Email    string           `json:"email" binding:"required,email"`
	Password string           `json:"password" binding:"required,min=8"`
	Name     string           `json:"name" binding:"required"`
	Role     models.UserRole  `json:"role" binding:"required"`
}

type UpdateUserRequest struct {
	Name   *string          `json:"name,omitempty"`
	Role   *models.UserRole `json:"role,omitempty"`
	Active *bool            `json:"active,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func NewUserService(db *gorm.DB, jwtService *auth.JWTService, audit *AuditService) *UserService {
	return &UserService{
		db:         db,
		jwtService: jwtService,
		audit:      audit,
	}
}

// Login authenticates a user and returns JWT tokens
func (s *UserService) Login(req *LoginRequest, ipAddress, userAgent string) (*auth.AuthResponse, error) {
	var user models.User
	
	// Find user by email
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Log failed login attempt
		s.audit.LogFailedLogin(req.Email, ipAddress, userAgent, "user_not_found")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.Active {
		s.audit.LogFailedLogin(req.Email, ipAddress, userAgent, "account_inactive")
		return nil, fmt.Errorf("account is inactive")
	}

	// Verify password
	if !s.jwtService.CheckPasswordHash(req.Password, user.Password) {
		s.audit.LogFailedLogin(req.Email, ipAddress, userAgent, "invalid_password")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate tokens
	tokens, err := s.jwtService.GenerateTokens(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update last login time
	now := time.Now()
	user.LastLogin = &now
	s.db.Save(&user)

	// Log successful login
	s.audit.LogUserAction(user.ID, models.ActionLogin, "authentication", ipAddress, userAgent, true, "")

	// Create user session
	sessionID := fmt.Sprintf("session_%d_%d", user.ID, time.Now().Unix())
	s.jwtService.CreateUserSession(user.ID, sessionID, ipAddress, userAgent)

	return tokens, nil
}

// RefreshToken generates new tokens using refresh token
func (s *UserService) RefreshToken(refreshToken string) (*auth.AuthResponse, error) {
	return s.jwtService.RefreshAccessToken(refreshToken)
}

// Logout invalidates user tokens and session
func (s *UserService) Logout(userID uint, accessToken string, ipAddress, userAgent string) error {
	// Blacklist the access token
	if err := s.jwtService.BlacklistToken(accessToken); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	// Invalidate all user sessions (optional - could be just current session)
	s.jwtService.InvalidateUserSessions(userID)

	// Log logout
	s.audit.LogUserAction(userID, models.ActionLogout, "authentication", ipAddress, userAgent, true, "")

	return nil
}

// CreateUser creates a new user account
func (s *UserService) CreateUser(req *CreateUserRequest, createdByUserID uint) (*models.User, error) {
	// Validate password strength
	if err := auth.ValidatePasswordStrength(req.Password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash password
	hashedPassword, err := s.jwtService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := models.User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
		Role:     req.Role,
		Active:   true,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Log user creation
	s.audit.LogUserAction(createdByUserID, models.ActionCreate, fmt.Sprintf("user:%d", user.ID), "", "", true, "")

	// Return user without password
	user.Password = ""
	return &user, nil
}

// GetUser retrieves user by ID
func (s *UserService) GetUser(userID uint, requestedByUserID uint, requestedByRole models.UserRole) (*models.User, error) {
	var user models.User
	
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check permissions - users can view their own profile, admins can view all
	if userID != requestedByUserID && requestedByRole != models.RoleAdmin {
		return nil, fmt.Errorf("insufficient permissions to view user")
	}

	// Remove sensitive data
	user.Password = ""
	return &user, nil
}

// GetAllUsers retrieves all users (admin only)
func (s *UserService) GetAllUsers(requestedByRole models.UserRole, page, limit int) ([]models.User, int64, error) {
	if requestedByRole != models.RoleAdmin {
		return nil, 0, fmt.Errorf("insufficient permissions to list users")
	}

	var users []models.User
	var total int64

	// Count total users
	s.db.Model(&models.User{}).Count(&total)

	// Get paginated users
	offset := (page - 1) * limit
	if err := s.db.Select("id, email, name, role, active, last_login, created_at, updated_at").
		Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve users: %w", err)
	}

	return users, total, nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(userID uint, req *UpdateUserRequest, updatedByUserID uint, updatedByRole models.UserRole) (*models.User, error) {
	var user models.User
	
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check permissions
	canUpdate := false
	if updatedByRole == models.RoleAdmin {
		canUpdate = true // Admins can update any user
	} else if userID == updatedByUserID {
		canUpdate = true // Users can update their own profile (limited fields)
		// Non-admins can only update their name
		if req.Role != nil || req.Active != nil {
			return nil, fmt.Errorf("insufficient permissions to modify role or active status")
		}
	}

	if !canUpdate {
		return nil, fmt.Errorf("insufficient permissions to update user")
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Role != nil && updatedByRole == models.RoleAdmin {
		updates["role"] = *req.Role
	}
	if req.Active != nil && updatedByRole == models.RoleAdmin {
		updates["active"] = *req.Active
		
		// If deactivating user, invalidate their sessions
		if !*req.Active {
			s.jwtService.InvalidateUserSessions(userID)
		}
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Log user update
	s.audit.LogUserAction(updatedByUserID, models.ActionUpdate, fmt.Sprintf("user:%d", user.ID), "", "", true, "")

	// Reload user data
	s.db.Where("id = ?", userID).First(&user)
	user.Password = ""
	return &user, nil
}

// ChangePassword changes user's password
func (s *UserService) ChangePassword(userID uint, req *ChangePasswordRequest, ipAddress, userAgent string) error {
	var user models.User
	
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify current password
	if !s.jwtService.CheckPasswordHash(req.CurrentPassword, user.Password) {
		s.audit.LogUserAction(userID, models.ActionUpdate, "password_change", ipAddress, userAgent, false, "invalid_current_password")
		return fmt.Errorf("current password is incorrect")
	}

	// Validate new password strength
	if err := auth.ValidatePasswordStrength(req.NewPassword); err != nil {
		return fmt.Errorf("new password validation failed: %w", err)
	}

	// Hash new password
	hashedPassword, err := s.jwtService.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	if err := s.db.Model(&user).Update("password", hashedPassword).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate all existing sessions to force re-login
	s.jwtService.InvalidateUserSessions(userID)

	// Log password change
	s.audit.LogUserAction(userID, models.ActionUpdate, "password_change", ipAddress, userAgent, true, "")

	return nil
}

// DeactivateUser deactivates a user account
func (s *UserService) DeactivateUser(userID uint, deactivatedByUserID uint, deactivatedByRole models.UserRole) error {
	if deactivatedByRole != models.RoleAdmin {
		return fmt.Errorf("insufficient permissions to deactivate user")
	}

	// Cannot deactivate yourself
	if userID == deactivatedByUserID {
		return fmt.Errorf("cannot deactivate your own account")
	}

	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update user status
	if err := s.db.Model(&user).Update("active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	// Invalidate all user sessions
	s.jwtService.InvalidateUserSessions(userID)

	// Log deactivation
	s.audit.LogUserAction(deactivatedByUserID, models.ActionUpdate, fmt.Sprintf("user:%d", user.ID), "", "", true, "account_deactivated")

	return nil
}

// GetUserSessions retrieves active sessions for a user
func (s *UserService) GetUserSessions(userID uint, requestedByUserID uint, requestedByRole models.UserRole) ([]database.UserSession, error) {
	// Check permissions - users can view their own sessions, admins can view all
	if userID != requestedByUserID && requestedByRole != models.RoleAdmin {
		return nil, fmt.Errorf("insufficient permissions to view user sessions")
	}

	sessions, err := s.jwtService.GetActiveUserSessions(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user sessions: %w", err)
	}

	return sessions, nil
}

// ValidateUserPermissions checks if user has permission for a specific action
func (s *UserService) ValidateUserPermissions(userID uint, action string, resource string) (bool, error) {
	var user models.User
	if err := s.db.Where("id = ? AND active = ?", userID, true).First(&user).Error; err != nil {
		return false, fmt.Errorf("user not found or inactive")
	}

	// Define permission matrix
	permissions := map[models.UserRole]map[string][]string{
		models.RoleAdmin: {
			"users":    {"create", "read", "update", "delete"},
			"patients": {"read", "create", "update"},
			"records":  {"read"},
			"audit":    {"read"},
			"system":   {"read", "update"},
		},
		models.RoleDoctor: {
			"patients": {"read", "create", "update"},
			"records":  {"read", "create", "update"},
			"audit":    {"read_own"},
		},
		models.RoleNurse: {
			"patients": {"read", "update"},
			"records":  {"read", "update"},
			"audit":    {"read_own"},
		},
	}

	rolePermissions, roleExists := permissions[user.Role]
	if !roleExists {
		return false, nil
	}

	resourcePermissions, resourceExists := rolePermissions[resource]
	if !resourceExists {
		return false, nil
	}

	for _, permission := range resourcePermissions {
		if permission == action {
			return true, nil
		}
	}

	return false, nil
}

// GetDefaultUsersByRole returns the first user for each role for login defaults
func (s *UserService) GetDefaultUsersByRole() (map[string]map[string]string, error) {
	defaults := make(map[string]map[string]string)

	roles := []models.UserRole{models.RoleDoctor, models.RoleNurse, models.RoleAdmin}

	for _, role := range roles {
		var user models.User
		if err := s.db.Where("role = ? AND active = ?", role, true).First(&user).Error; err != nil {
			// If no user found for this role, skip it
			continue
		}

		roleStr := string(role)
		password := "Doctor123" // Default password for doctor

		// Set password based on role (matching database seed)
		switch role {
		case models.RoleNurse:
			password = "Nurse123"
		case models.RoleAdmin:
			password = "admin123"
		case models.RoleDoctor:
			password = "Doctor123"
		}

		// Special case for custom users
		if user.Email == "vedantparmar@gmail.com" {
			password = "Vedant1234"
		}

		defaults[roleStr] = map[string]string{
			"email":    user.Email,
			"password": password,
		}
	}

	return defaults, nil
}