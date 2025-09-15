# Security Implementation Guide

## Overview

HealthSecure implements comprehensive security measures to protect sensitive medical data and ensure HIPAA compliance. This document outlines the security architecture, implementation details, and best practices.

## Table of Contents

1. [Authentication & Authorization](#authentication--authorization)
2. [Data Protection](#data-protection)
3. [Audit Logging](#audit-logging)
4. [Emergency Access](#emergency-access)
5. [Network Security](#network-security)
6. [Input Validation](#input-validation)
7. [Session Management](#session-management)
8. [Monitoring & Alerting](#monitoring--alerting)
9. [Security Best Practices](#security-best-practices)

## Authentication & Authorization

### JWT Token Implementation

- **Token Expiration**: Access tokens expire in 15 minutes
- **Refresh Tokens**: Valid for 7 days with automatic rotation
- **Token Structure**: Contains user ID, role, and expiration claims
- **Secret Management**: JWT secrets must be minimum 32 characters

```go
type Claims struct {
    UserID uint   `json:"user_id"`
    Role   string `json:"role"`
    Name   string `json:"name"`
    jwt.StandardClaims
}
```

### Role-Based Access Control (RBAC)

#### User Roles

1. **Doctor**
   - Full patient data access including SSN
   - Create and modify medical records
   - Emergency access capabilities
   - View diagnostic information

2. **Nurse** 
   - Limited patient data access (no SSN)
   - View basic medical information
   - Update care notes
   - Emergency access for critical situations

3. **Admin**
   - User management and system configuration
   - Audit log access and monitoring
   - No direct patient data access
   - Emergency access oversight

#### Implementation

```go
// Middleware checks
func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("user_role")
        for _, role := range roles {
            if userRole == role {
                c.Next()
                return
            }
        }
        c.JSON(403, gin.H{"error": "Insufficient permissions"})
        c.Abort()
    }
}

// Service-level filtering
func (s *PatientService) GetPatient(userRole string, patientID uint) (*Patient, error) {
    switch userRole {
    case "doctor":
        // Full access including sensitive data
    case "nurse":
        // Limited access, exclude SSN
    case "admin":
        return nil, errors.New("administrators cannot access patient data")
    }
}
```

### OAuth2 Integration

Support for enterprise identity providers:
- Google Workspace
- Azure Active Directory
- Custom OIDC providers

## Data Protection

### Encryption at Rest

- **Database**: MySQL with InnoDB encryption
- **Configuration**: AES-256 encryption for sensitive config values
- **File Storage**: Encrypted file system for documents

### Encryption in Transit

- **HTTPS**: TLS 1.2+ required for all connections
- **Database**: TLS connection to MySQL required in production
- **API**: All endpoints require HTTPS in production

### Password Security

```go
// Password hashing with bcrypt
func HashPassword(password string) (string, error) {
    // Minimum cost of 12 for HIPAA compliance
    hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    return string(hash), err
}
```

### Data Sanitization

- **SSN Masking**: SSN displayed as XXX-XX-XXXX for nurses
- **Audit Logs**: No sensitive data (SSN, diagnosis details) logged
- **API Responses**: Role-based data filtering

## Audit Logging

### Required Audit Events

All the following actions are automatically logged:

- User authentication (success/failure)
- Patient data access
- Medical record creation/modification
- Emergency access requests and usage
- Administrative actions
- Failed authorization attempts

### Audit Log Structure

```go
type AuditLog struct {
    ID           uint      `json:"id"`
    UserID       uint      `json:"user_id"`
    PatientID    *uint     `json:"patient_id,omitempty"`
    Action       string    `json:"action"`
    Resource     string    `json:"resource"`
    IPAddress    string    `json:"ip_address"`
    UserAgent    string    `json:"user_agent"`
    EmergencyUse bool      `json:"emergency_use"`
    Success      bool      `json:"success"`
    Timestamp    time.Time `json:"timestamp"`
}
```

### Audit Trail Integrity

- **Immutable Logs**: Audit logs cannot be modified once created
- **Retention**: Minimum 6 years for HIPAA compliance
- **Backup**: Regular encrypted backups of audit data
- **Monitoring**: Real-time alerting for suspicious patterns

## Emergency Access

### Break-Glass Access Protocol

1. **Request**: User requests emergency access with justification
2. **Token Generation**: Secure, time-limited access token created
3. **Notification**: Immediate notification to administrators
4. **Enhanced Logging**: All emergency actions logged with context
5. **Automatic Expiration**: Default 1-hour access window

### Implementation

```go
type EmergencyAccess struct {
    ID          uint      `json:"id"`
    UserID      uint      `json:"user_id"`
    PatientID   uint      `json:"patient_id"`
    Reason      string    `json:"reason"`
    AccessToken string    `json:"access_token"`
    ExpiresAt   time.Time `json:"expires_at"`
    CreatedAt   time.Time `json:"created_at"`
}

func (s *EmergencyService) RequestEmergencyAccess(userID, patientID uint, reason string) (*EmergencyAccess, error) {
    // Validate justification
    if len(reason) < 20 {
        return nil, errors.New("emergency access requires detailed justification")
    }
    
    // Generate secure token
    token := generateSecureToken(32)
    
    // Create time-limited access
    access := &EmergencyAccess{
        UserID:      userID,
        PatientID:   patientID,
        Reason:      reason,
        AccessToken: token,
        ExpiresAt:   time.Now().Add(1 * time.Hour),
    }
    
    // Enhanced audit logging
    s.auditService.LogEmergencyRequest(userID, patientID, reason)
    
    // Notify administrators
    s.notificationService.NotifyEmergencyAccess(access)
    
    return access, s.db.Create(access).Error
}
```

## Network Security

### API Security Headers

```go
func SecurityHeadersMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Header("Content-Security-Policy", "default-src 'self'")
        c.Next()
    })
}
```

### CORS Configuration

```go
func CORSMiddleware(config *configs.Config) gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins:     config.App.CORSOrigins,
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    })
}
```

### Rate Limiting

- **Authentication Endpoints**: 5 attempts per 15 minutes
- **API Endpoints**: 100 requests per hour per user
- **Emergency Access**: 3 requests per hour per user

## Input Validation

### Data Validation Rules

```go
type UserCreateRequest struct {
    Email    string   `json:"email" binding:"required,email"`
    Name     string   `json:"name" binding:"required,min=2,max=100"`
    Role     UserRole `json:"role" binding:"required,oneof=doctor nurse admin"`
    Password string   `json:"password" binding:"required,min=8"`
}

type PatientCreateRequest struct {
    FirstName        string    `json:"first_name" binding:"required,max=50"`
    LastName         string    `json:"last_name" binding:"required,max=50"`
    DateOfBirth      time.Time `json:"date_of_birth" binding:"required"`
    SSN              string    `json:"ssn" binding:"omitempty,len=11"`
    Phone            string    `json:"phone" binding:"omitempty,max=15"`
    Address          string    `json:"address" binding:"omitempty,max=200"`
    EmergencyContact string    `json:"emergency_contact" binding:"omitempty,max=100"`
}
```

### SQL Injection Prevention

- **Parameterized Queries**: All database queries use GORM parameterization
- **Input Sanitization**: HTML and SQL special characters escaped
- **Whitelist Validation**: Enum values validated against allowed lists

### XSS Prevention

```go
func sanitizeInput(input string) string {
    // Remove HTML tags
    p := bluemonday.StripTagsPolicy()
    cleaned := p.Sanitize(input)
    
    // Escape special characters
    return html.EscapeString(cleaned)
}
```

## Session Management

### Token Lifecycle

1. **Generation**: Secure random token generation
2. **Storage**: Redis-based session storage with expiration
3. **Validation**: Token validation on each request
4. **Rotation**: Refresh token rotation on use
5. **Revocation**: Immediate token invalidation on logout

### Security Measures

```go
// Token blacklisting
func (j *JWTService) IsTokenBlacklisted(token string) bool {
    result := j.redis.Get(context.Background(), "blacklist:"+token)
    return result.Err() == nil
}

func (j *JWTService) BlacklistToken(token string, expiration time.Duration) error {
    return j.redis.Set(context.Background(), "blacklist:"+token, true, expiration).Err()
}
```

## Monitoring & Alerting

### Security Events

Immediate alerts for:
- Multiple failed authentication attempts
- Emergency access requests
- Unusual data access patterns
- Administrative privilege escalation
- Database connection failures
- SSL certificate expiration

### Metrics Collection

```go
// Prometheus metrics
var (
    authFailures = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "healthsecure_auth_failures_total",
            Help: "Total number of authentication failures",
        },
        []string{"reason"},
    )
    
    emergencyAccess = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "healthsecure_emergency_access_total",
            Help: "Total number of emergency access requests",
        },
        []string{"user_role"},
    )
)
```

### Log Analysis

- **ELK Stack**: Elasticsearch, Logstash, Kibana for log analysis
- **Anomaly Detection**: Machine learning-based pattern recognition
- **Real-time Dashboards**: Security metrics visualization

## Security Best Practices

### Development Guidelines

1. **Secure Coding**
   - Input validation on all user data
   - Output encoding for web responses
   - Error handling without information disclosure
   - Secure random number generation

2. **Database Security**
   - Principle of least privilege for database users
   - Regular security updates and patches
   - Database connection encryption
   - Query performance monitoring

3. **API Security**
   - Authentication required for all endpoints
   - Authorization checks at service layer
   - Request/response size limits
   - API versioning and deprecation

### Deployment Security

1. **Environment Separation**
   - Isolated production environment
   - Secure configuration management
   - Environment-specific secrets
   - Network segmentation

2. **Container Security**
   - Minimal base images
   - Non-root container execution
   - Regular image vulnerability scanning
   - Secure registry usage

3. **Infrastructure Security**
   - Regular security assessments
   - Automated patch management
   - Backup and recovery testing
   - Incident response procedures

### Compliance Considerations

1. **HIPAA Requirements**
   - Administrative safeguards implementation
   - Physical safeguards documentation
   - Technical safeguards verification
   - Risk assessment and management

2. **Regular Security Reviews**
   - Quarterly security assessments
   - Annual penetration testing
   - Code security reviews
   - Compliance audits

## Incident Response

### Security Incident Types

1. **Data Breach**
   - Unauthorized access to patient data
   - Data exfiltration attempts
   - Database compromise

2. **Authentication Bypass**
   - JWT token compromise
   - Session hijacking
   - Privilege escalation

3. **System Compromise**
   - Malware detection
   - Unauthorized system access
   - Service disruption attacks

### Response Procedures

1. **Immediate Actions**
   - Isolate affected systems
   - Preserve evidence
   - Notify security team
   - Begin containment

2. **Investigation**
   - Analyze audit logs
   - Determine impact scope
   - Identify root cause
   - Document findings

3. **Recovery**
   - Remove threat
   - Restore services
   - Implement improvements
   - Monitor for recurrence

4. **Post-Incident**
   - Conduct lessons learned
   - Update procedures
   - Staff training
   - Compliance reporting

## Security Testing

### Automated Testing

```bash
# Security scanning
gosec ./...
npm audit
docker scan healthsecure:latest

# Dependency checking
go list -m all | nancy sleuth
npm audit --audit-level high

# Static analysis
golangci-lint run --enable-all
```

### Manual Testing

1. **Penetration Testing**
   - External network testing
   - Web application testing
   - Social engineering assessment
   - Physical security review

2. **Code Review**
   - Security-focused code review
   - Architecture security review
   - Configuration review
   - Third-party library assessment

---

**Note**: This security implementation is continuously updated based on evolving threats and compliance requirements. Regular security assessments ensure ongoing protection of sensitive medical data.
