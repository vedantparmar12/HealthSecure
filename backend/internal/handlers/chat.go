package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"healthsecure/internal/auth"
	"healthsecure/internal/models"
	"healthsecure/internal/services"

	"github.com/gin-gonic/gin"
)

// Sanitize input to remove potential HTML tags
var tagRegex = regexp.MustCompile(`<(.|\n)*?>`)

func sanitizeInput(input string) string {
	return tagRegex.ReplaceAllString(input, "")
}

type ChatHandler struct {
	userService       *services.UserService
	patientService    *services.PatientService
	auditService      *services.AuditService
	chatThreadService *services.ChatThreadService
	jwtService        *auth.JWTService
	aiServiceClient   *AIServiceClient
}

type ChatConfig struct {
	Configurable struct {
		ThreadID string `json:"thread_id"`
	} `json:"configurable"`
}

type ChatUser struct {
	Role string `json:"role"`
	Name string `json:"name"`
}

type ChatRequest struct {
	Message string     `json:"message" binding:"required"`
	Config  ChatConfig `json:"config"`
	User    ChatUser   `json:"user"`
}

type ThreadRequest struct {
	Title string `json:"title"`
}

type FeedbackRequest struct {
	MessageID string `json:"message_id" binding:"required"`
	Feedback  string `json:"feedback" binding:"required,oneof=thumbs_up thumbs_down"`
}

type ChatResponse struct {
	Response  string                 `json:"response"`
	MessageID string                 `json:"message_id"`
	ThreadID  string                 `json:"thread_id"`
	RunID     string                 `json:"run_id,omitempty"`
	NewTitle  string                 `json:"new_title,omitempty"`
	Actions   []string               `json:"actions,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

func NewChatHandler(userService *services.UserService, patientService *services.PatientService, auditService *services.AuditService, chatThreadService *services.ChatThreadService, jwtService *auth.JWTService) *ChatHandler {
	return &ChatHandler{
		userService:       userService,
		patientService:    patientService,
		auditService:      auditService,
		chatThreadService: chatThreadService,
		jwtService:        jwtService,
		aiServiceClient:   NewAIServiceClient(),
	}
}

func (h *ChatHandler) ProcessChatMessage(c *gin.Context) {
	log.Printf("=== ProcessChatMessage START ===")
	log.Printf("Request Headers: %+v", c.Request.Header)

	// Read the raw body for logging
	bodyBytes, _ := c.GetRawData()
	log.Printf("Raw request body: %s", string(bodyBytes))

	// Reset the body so it can be read again
	c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ERROR: JSON binding failed: %v", err)
		log.Printf("Request body that failed: %s", string(bodyBytes))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request format: %v", err),
		})
		return
	}

	log.Printf("Parsed ChatRequest: %+v", req)

	// Sanitize user message
	req.Message = sanitizeInput(req.Message)

	// Get current user from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userIDStr := fmt.Sprintf("%v", userID)

	var threadID string
	if strings.HasPrefix(req.Config.Configurable.ThreadID, "temp_") || req.Config.Configurable.ThreadID == "" {
		// Create new thread if no thread_id in config
		thread, err := h.chatThreadService.CreateThread(userIDStr, "New Chat")
		if err != nil {
			log.Printf("Failed to create thread: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create chat thread",
			})
			return
		}
		threadID = thread.ThreadID
	} else {
		threadID = req.Config.Configurable.ThreadID
		// Verify thread exists and belongs to user
		_, err := h.chatThreadService.GetThread(threadID, userIDStr)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid or inaccessible thread ID",
			})
			return
		}
	}

	// Save user message to thread
	_, err := h.chatThreadService.SaveMessage(threadID, "user", req.Message, "")
	if err != nil {
		log.Printf("Failed to save user message: %v", err)
	}

	// Log the chat interaction for audit
	userIDUint, _ := strconv.ParseUint(userIDStr, 10, 32)
	h.auditService.LogUserAction(
		uint(userIDUint),
		models.ActionCreate,
		fmt.Sprintf("chat_message:%s", threadID),
		c.ClientIP(),
		c.Request.UserAgent(),
		true,
		fmt.Sprintf("User sent message in thread %s", threadID),
	)

	// Process the message through AI service with thread context
	response, err := h.processChatMessageWithAIService(req, userIDStr, threadID)
	if err != nil {
		log.Printf("AI service processing error: %v", err)
		// Check for specific AI service configuration error
		if strings.Contains(err.Error(), "AI service is not configured") {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "The AI assistant is not configured on the server. Please contact the administrator.",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get response from AI assistant",
			})
		}
		return
	}

	// Save assistant response to thread
	assistantMessage, err := h.chatThreadService.SaveMessage(threadID, "assistant", response.Response, response.RunID)
	if err != nil {
		log.Printf("Failed to save assistant message: %v", err)
	} else {
		response.MessageID = assistantMessage.MessageID
	}

	// Update thread title if a new one was generated
	if response.NewTitle != "" {
		err := h.chatThreadService.UpdateThreadTitle(threadID, userIDStr, response.NewTitle)
		if err != nil {
			log.Printf("Warning: Failed to update thread title: %v", err)
		}
	}

	response.ThreadID = threadID

	c.JSON(http.StatusOK, response)
}

func (h *ChatHandler) processChatMessageWithAIService(req ChatRequest, userID, threadID string) (*ChatResponse, error) {
	log.Printf("processChatMessageWithAIService called - UserID: %s, ThreadID: %s, Message: %s", userID, threadID, req.Message)
	
	// First, test AI service health
	if err := h.aiServiceClient.HealthCheck(); err != nil {
		log.Printf("AI service health check failed: %v", err)
		return nil, fmt.Errorf("AI service unavailable: %w", err)
	}
	
	log.Printf("AI service health check passed")
	
	// Get user info from database if not provided in request
	userRole := req.User.Role
	userName := req.User.Name
	
	if userRole == "" || userName == "" {
		// Fallback: get user info from database
		userIDUint, _ := strconv.ParseUint(userID, 10, 32)
		// Use the user's own role or admin to avoid permission issues
		var roleForQuery models.UserRole = models.RoleAdmin
		if userRole != "" {
			roleForQuery = models.UserRole(userRole)
		}
		
		user, err := h.userService.GetUser(uint(userIDUint), uint(userIDUint), roleForQuery)
		if err == nil {
			if userRole == "" {
				userRole = string(user.Role)
			}
			if userName == "" {
				userName = user.Name
			}
		} else {
			log.Printf("Warning: Could not get user info from database: %v", err)
			// Set fallback values if database query fails
			if userRole == "" {
				userRole = "doctor"
			}
			if userName == "" {
				userName = "Unknown User"
			}
		}
	}

	// Get chat history
	history, err := h.chatThreadService.GetThreadMessages(threadID, userID, 10) // Get last 10 messages
	if err != nil {
		log.Printf("Warning: Could not get chat history: %v", err)
	}
	
	// Prepare request for Python AI service
	aiReq := AIServiceRequest{
		Message:  req.Message,
		ThreadID: threadID,
		UserID:   userID,
		UserRole: userRole,
		UserName: userName,
		History:  history,
	}

	log.Printf("Calling Python AI service with request: %+v", aiReq)

	// Call Python AI service
	aiResp, err := h.aiServiceClient.Chat(aiReq)
	if err != nil {
		log.Printf("AI service error: %v", err)
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}
	
	log.Printf("Python AI service response received: %+v", aiResp)

	// Create response object
	response := &ChatResponse{
		Response: aiResp.Response,
		RunID:    aiResp.RunID,
		Actions:  []string{},
		Data:     make(map[string]interface{}),
		ThreadID: aiResp.ThreadID,
		NewTitle: aiResp.NewTitle,
	}

	// Add metadata
	response.Data["model_used"] = aiResp.ModelUsed
	if aiResp.TokensUsed != nil {
		response.Data["tokens_used"] = *aiResp.TokensUsed
	}

	// Handle special commands or data queries (optional - could move to Python service)
	if err := h.processSpecialCommands(req.Message, req.User.Role, response); err != nil {
		log.Printf("Error processing special commands: %v", err)
	}
	
	// Note: Message history is now handled by the Python AI service
	
	log.Printf("AI response generated successfully - Run ID: %s", aiResp.RunID)

	return response, nil
}

// New endpoint methods
func (h *ChatHandler) CreateThread(c *gin.Context) {
	var req ThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	title := req.Title
	if title == "" {
		title = "New Chat"
	}

	thread, err := h.chatThreadService.CreateThread(fmt.Sprintf("%v", userID), title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create thread",
		})
		return
	}

	c.JSON(http.StatusCreated, thread)
}

func (h *ChatHandler) GetThreads(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	threads, err := h.chatThreadService.GetUserThreads(fmt.Sprintf("%v", userID), 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get threads",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"threads": threads,
	})
}

func (h *ChatHandler) GetThreadMessages(c *gin.Context) {
	threadID := c.Param("thread_id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Thread ID required",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	messages, total, err := h.chatThreadService.GetThreadMessagesPage(threadID, fmt.Sprintf("%v", userID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Thread not found or access denied",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (h *ChatHandler) SubmitFeedback(c *gin.Context) {
	var req FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	err := h.chatThreadService.SubmitFeedback(req.MessageID, fmt.Sprintf("%v", userID), req.Feedback)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Message not found or access denied",
		})
		return
	}

	// Log feedback to LangSmith
	userIDUint, _ := strconv.ParseUint(fmt.Sprintf("%v", userID), 10, 32)
	h.auditService.LogUserAction(
		uint(userIDUint),
		models.ActionUpdate,
		fmt.Sprintf("chat_feedback:%s", req.MessageID),
		c.ClientIP(),
		c.Request.UserAgent(),
		true,
		fmt.Sprintf("User submitted %s feedback for message %s", req.Feedback, req.MessageID),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Feedback submitted successfully",
	})
}

func (h *ChatHandler) GetFeedbackStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	stats, err := h.chatThreadService.GetMessageFeedback(fmt.Sprintf("%v", userID), 30) // Last 30 days
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get feedback statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feedback_stats": stats,
	})
}

func (h *ChatHandler) processSpecialCommands(message, role string, response *ChatResponse) error {
	lowerMsg := strings.ToLower(message)

	// Handle patient search queries
	if strings.Contains(lowerMsg, "find patient") || strings.Contains(lowerMsg, "search patient") {
		response.Actions = append(response.Actions, "patient_search")
		response.Data["search_context"] = "patient_lookup"
	}

	// Handle emergency access requests
	if strings.Contains(lowerMsg, "emergency") && (role == "doctor" || role == "nurse") {
		response.Actions = append(response.Actions, "emergency_access")
		response.Data["emergency_context"] = "access_request"
	}

	// Handle audit log requests
	if strings.Contains(lowerMsg, "audit") || strings.Contains(lowerMsg, "logs") {
		response.Actions = append(response.Actions, "audit_logs")
		response.Data["audit_context"] = "log_access"
	}

	// Handle patient creation requests
	if strings.Contains(lowerMsg, "add patient") || strings.Contains(lowerMsg, "new patient") {
		if role == "doctor" || role == "nurse" {
			response.Actions = append(response.Actions, "create_patient")
			response.Data["creation_context"] = "patient_form"
		}
	}

	return nil
}