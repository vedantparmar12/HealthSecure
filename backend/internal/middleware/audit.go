package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"healthsecure/internal/models"

	"github.com/gin-gonic/gin"
)

// AuditMiddleware automatically logs API requests for audit trail
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip audit for health check and static files
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		startTime := time.Now()

		// Capture request body (for POST/PUT requests)
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Use a custom response writer to capture response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process request
		c.Next()

		// Log the request after processing
		logAuditEvent(c, startTime, requestBody, blw.body.String())
	}
}

// bodyLogWriter captures response body for audit logging
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// logAuditEvent creates an audit log entry
func logAuditEvent(c *gin.Context, startTime time.Time, requestBody []byte, responseBody string) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		return // Skip logging for unauthenticated requests
	}

	// Determine action based on HTTP method and path
	action := getActionFromRequest(c)
	
	// Get resource being accessed
	resource := c.Request.URL.Path

	// Check if this was an emergency access request
	emergencyAccess := c.GetHeader("X-Emergency-Access-Token") != ""

	// Prepare audit log data
	auditData := map[string]interface{}{
		"user_id":        userID,
		"action":         action,
		"resource":       resource,
		"method":         c.Request.Method,
		"ip_address":     c.ClientIP(),
		"user_agent":     c.Request.UserAgent(),
		"emergency_use":  emergencyAccess,
		"success":        c.Writer.Status() < 400,
		"status_code":    c.Writer.Status(),
		"duration_ms":    time.Since(startTime).Milliseconds(),
		"timestamp":      time.Now(),
	}

	// Add error message if request failed
	if c.Writer.Status() >= 400 {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal([]byte(responseBody), &errorResponse); err == nil {
			if errorMsg, exists := errorResponse["error"]; exists {
				auditData["error_message"] = errorMsg
			}
		}
	}

	// Log sensitive operations with more detail
	if isSensitiveOperation(c) {
		auditData["details"] = getSensitiveOperationDetails(c, requestBody)
	}

	// In a real implementation, this would save to the audit_logs table
	// For now, we'll just log to console in development
	logToConsole(auditData)
}

func getActionFromRequest(c *gin.Context) models.AuditAction {
	method := c.Request.Method
	path := c.Request.URL.Path

	switch {
	case method == "GET" && path == "/api/auth/login":
		return models.ActionLogin
	case method == "POST" && path == "/api/auth/logout":
		return models.ActionLogout
	case method == "GET":
		return models.ActionView
	case method == "POST":
		return models.ActionCreate
	case method == "PUT" || method == "PATCH":
		return models.ActionUpdate
	case method == "DELETE":
		return models.ActionDelete
	default:
		return models.ActionView
	}
}

func isSensitiveOperation(c *gin.Context) bool {
	path := c.Request.URL.Path
	
	// Define sensitive operations
	sensitivePatterns := []string{
		"/api/patients",
		"/api/records",
		"/api/emergency",
		"/api/admin",
	}

	for _, pattern := range sensitivePatterns {
		if len(path) >= len(pattern) && path[:len(pattern)] == pattern {
			return true
		}
	}

	return false
}

func getSensitiveOperationDetails(c *gin.Context, requestBody []byte) map[string]interface{} {
	details := make(map[string]interface{})

	// Add query parameters (excluding sensitive data)
	queryParams := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if key != "ssn" && key != "password" {
			queryParams[key] = values[0]
		}
	}
	details["query_params"] = queryParams

	// Add request size
	details["request_size"] = len(requestBody)

	// Add any path parameters
	pathParams := make(map[string]string)
	for _, param := range c.Params {
		pathParams[param.Key] = param.Value
	}
	details["path_params"] = pathParams

	return details
}

func logToConsole(auditData map[string]interface{}) {
	// In development, log to console
	// In production, this would save to database
	if data, err := json.Marshal(auditData); err == nil {
		println("AUDIT:", string(data))
	}
}