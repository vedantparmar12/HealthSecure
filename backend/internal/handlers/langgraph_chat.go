package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"healthsecure/internal/auth"
	"healthsecure/internal/models"
	"healthsecure/internal/services"

	"github.com/gin-gonic/gin"
)

type LangGraphChatHandler struct {
	userService       *services.UserService
	patientService    *services.PatientService
	auditService      *services.AuditService
	chatThreadService *services.ChatThreadService
	jwtService        *auth.JWTService
}

type ChatRequest struct {
	Message string `json:"message" binding:"required"`
	Config  struct {
		Configurable struct {
			ThreadID string `json:"thread_id"`
		} `json:"configurable"`
		Metadata struct {
			ThreadID string `json:"thread_id"`
			UserID   string `json:"user_id"`
			UserRole string `json:"user_role"`
		} `json:"metadata"`
		RunName string `json:"run_name"`
	} `json:"config"`
	User struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Email       string   `json:"email"`
		Role        string   `json:"role"`
		Permissions []string `json:"permissions"`
	} `json:"user"`
	Context struct {
		Role        string   `json:"role"`
		Permissions []string `json:"permissions"`
		SessionID   string   `json:"session_id"`
	} `json:"context"`
}

type ThreadRequest struct {
	Title string `json:"title"`
}

type FeedbackRequest struct {
	MessageID string `json:"message_id" binding:"required"`
	Feedback  string `json:"feedback" binding:"required,oneof=thumbs_up thumbs_down"`
}

type LangSmithTrace struct {
	RunID     string            `json:"run_id"`
	ThreadID  string            `json:"thread_id"`
	StartTime time.Time         `json:"start_time"`
	EndTime   time.Time         `json:"end_time"`
	Inputs    map[string]string `json:"inputs"`
	Outputs   map[string]string `json:"outputs"`
	Metadata  map[string]string `json:"metadata"`
}

type ChatResponse struct {
	Response  string                 `json:"response"`
	MessageID string                 `json:"message_id"`
	ThreadID  string                 `json:"thread_id"`
	RunID     string                 `json:"run_id,omitempty"`
	Actions   []string               `json:"actions,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

type OpenRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func NewLangGraphChatHandler(userService *services.UserService, patientService *services.PatientService, auditService *services.AuditService, chatThreadService *services.ChatThreadService, jwtService *auth.JWTService) *LangGraphChatHandler {
	return &LangGraphChatHandler{
		userService:       userService,
		patientService:    patientService,
		auditService:      auditService,
		chatThreadService: chatThreadService,
		jwtService:        jwtService,
	}
}

func (h *LangGraphChatHandler) ProcessChatMessage(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Get current user from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userIDStr := fmt.Sprintf("%v", userID)

	// Get thread ID from CONFIG structure (like Streamlit example)
	var threadID string
	if req.Config.Configurable.ThreadID != "" {
		threadID = req.Config.Configurable.ThreadID
		// Verify thread exists and belongs to user, or create if it doesn't exist
		_, err := h.chatThreadService.GetThread(threadID, userIDStr)
		if err != nil {
			// Thread doesn't exist, create it
			_, err := h.chatThreadService.CreateThread(userIDStr, "Chat Session")
			if err != nil {
				log.Printf("Failed to create thread: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to create chat thread",
				})
				return
			}
		}
	} else {
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

	// Process the message through LangGraph agent with thread context
	response, err := h.processWithLangGraph(req, userIDStr, threadID)
	if err != nil {
		log.Printf("LangGraph processing error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process message with AI agent",
		})
		return
	}

	// Save assistant response to thread
	assistantMessage, err := h.chatThreadService.SaveMessage(threadID, "assistant", response.Response, response.RunID)
	if err != nil {
		log.Printf("Failed to save assistant message: %v", err)
	} else {
		response.MessageID = assistantMessage.MessageID
	}

	response.ThreadID = threadID

	c.JSON(http.StatusOK, response)
}

func (h *LangGraphChatHandler) processWithLangGraph(req ChatRequest, userID, threadID string) (*ChatResponse, error) {
	// Generate run ID for LangSmith tracking
	runID := h.generateRunID()
	
	// Get conversation history for context
	messages, err := h.chatThreadService.GetThreadMessages(threadID, userID, 50) // Last 50 messages
	if err != nil {
		log.Printf("Failed to get thread history: %v", err)
		messages = []services.ChatMessage{} // Continue without history
	}

	// Create system prompt based on user role and permissions
	systemPrompt := h.createSystemPrompt(req.User.Role, req.User.Permissions)
	
	// Send to LangSmith tracing
	traceStart := time.Now()
	h.sendLangSmithTrace(runID, threadID, userID, req.Message, traceStart, nil)

	// Call OpenRouter API with LangChain integration and conversation history
	aiResponse, err := h.callOpenRouterWithHistory(systemPrompt, req.Message, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	// Parse the AI response and determine actions
	response := &ChatResponse{
		Response:  aiResponse,
		RunID:     runID,
		Actions:   []string{},
		Data:      make(map[string]interface{}),
	}

	// Process special commands or queries
	if err := h.processSpecialCommands(req.Message, req.User.Role, response); err != nil {
		log.Printf("Error processing special commands: %v", err)
	}

	// Send completion trace to LangSmith
	traceEnd := time.Now()
	h.sendLangSmithTrace(runID, threadID, userID, req.Message, traceStart, &traceEnd)

	return response, nil
}

func (h *LangGraphChatHandler) createSystemPrompt(role string, permissions []string) string {
	basePrompt := `You are an AI medical assistant for HealthSecure, a HIPAA-compliant medical data management system. 

IMPORTANT CONTEXT:
- User Role: %s
- Available Permissions: %s
- Current Date: %s

CAPABILITIES YOU CAN HELP WITH:
1. **Patient Management**: Help with adding, viewing, and managing patient records
2. **Medical Records**: Assist with accessing and updating medical information  
3. **Emergency Protocols**: Guide through emergency access procedures
4. **System Navigation**: Help users find features and navigate the system
5. **Audit & Compliance**: Explain audit logs and compliance requirements

ROLE-SPECIFIC PERMISSIONS:
- **Admin**: Full system access, user management, audit logs, system configuration
- **Doctor**: Full patient access, medical records, prescriptions, emergency access
- **Nurse**: Patient viewing, basic medical records, emergency protocols

RESPONSE GUIDELINES:
- Always maintain HIPAA compliance in responses
- Be professional and medical-context appropriate
- Provide clear, actionable guidance
- If a user asks about something outside their permissions, explain the limitation politely
- For database operations, provide clear instructions but emphasize data privacy
- Keep responses professional and medical-context appropriate

SECURITY NOTES:
- Never display actual patient data in examples
- Always remind users about audit logging for sensitive operations
- Emphasize data privacy and security best practices

Remember: You are helping healthcare professionals efficiently and securely manage medical data while maintaining strict compliance standards.`

	return fmt.Sprintf(basePrompt, role, strings.Join(permissions, ", "), "2024")
}

func (h *LangGraphChatHandler) callOpenRouter(systemPrompt, userMessage string) (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OpenRouter API key not configured")
	}

	// Prepare the request
	reqBody := OpenRouterRequest{
		Model: "openai/gpt-3.5-turbo", // Using GPT-3.5 Turbo
		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userMessage,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make the API request
	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://healthsecure.local")
	req.Header.Set("X-Title", "HealthSecure Medical Assistant")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var response OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("OpenRouter API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "I'm sorry, I couldn't process your request at this time. Please try again.", nil
	}

	return response.Choices[0].Message.Content, nil
}

func (h *LangGraphChatHandler) callOpenRouterWithHistory(systemPrompt, userMessage string, history []services.ChatMessage) (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OpenRouter API key not configured")
	}

	// Build conversation history
	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	// Add conversation history (limit to prevent token overflow)
	historyLimit := 20 // Last 20 messages
	startIdx := 0
	if len(history) > historyLimit {
		startIdx = len(history) - historyLimit
	}

	for i := startIdx; i < len(history); i++ {
		msg := history[i]
		messages = append(messages, Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Add current user message
	messages = append(messages, Message{
		Role:    "user",
		Content: userMessage,
	})

	// Prepare the request
	reqBody := OpenRouterRequest{
		Model:    "openai/gpt-3.5-turbo", // Using GPT-3.5 Turbo
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make the API request
	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://healthsecure.local")
	req.Header.Set("X-Title", "HealthSecure Medical Assistant")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var response OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("OpenRouter API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "I'm sorry, I couldn't process your request at this time. Please try again.", nil
	}

	return response.Choices[0].Message.Content, nil
}

func (h *LangGraphChatHandler) generateRunID() string {
	return fmt.Sprintf("run_%d", time.Now().UnixNano())
}

func (h *LangGraphChatHandler) sendLangSmithTrace(runID, threadID, userID, input string, startTime time.Time, endTime *time.Time) {
	langsmithURL := os.Getenv("LANGCHAIN_ENDPOINT")
	langsmithKey := os.Getenv("LANGCHAIN_API_KEY")
	
	if langsmithURL == "" || langsmithKey == "" {
		return // Skip if not configured
	}

	// Create trace similar to Streamlit CONFIG structure
	trace := LangSmithTrace{
		RunID:     runID,
		ThreadID:  threadID,
		StartTime: startTime,
		Inputs: map[string]string{
			"user_input": input,
		},
		Metadata: map[string]string{
			"thread_id":    threadID,
			"user_id":      userID,
			"app":          "healthsecure",
			"version":      "1.0.0",
			"run_name":     "chat_turn",
			"configurable": fmt.Sprintf(`{"thread_id": "%s"}`, threadID),
		},
	}

	if endTime != nil {
		trace.EndTime = *endTime
		trace.Outputs = map[string]string{
			"status": "completed",
		}
	}

	// Send trace to LangSmith (async)
	go func() {
		jsonData, err := json.Marshal(trace)
		if err != nil {
			log.Printf("Failed to marshal LangSmith trace: %v", err)
			return
		}

		req, err := http.NewRequest("POST", langsmithURL+"/runs", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Failed to create LangSmith request: %v", err)
			return
		}

		req.Header.Set("Authorization", "Bearer "+langsmithKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to send LangSmith trace: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			log.Printf("LangSmith trace failed with status: %d", resp.StatusCode)
		}
	}()
}

// New endpoint methods
func (h *LangGraphChatHandler) CreateThread(c *gin.Context) {
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

func (h *LangGraphChatHandler) GetThreads(c *gin.Context) {
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

func (h *LangGraphChatHandler) GetThreadMessages(c *gin.Context) {
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

	messages, err := h.chatThreadService.GetThreadMessages(threadID, fmt.Sprintf("%v", userID), 100)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Thread not found or access denied",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
	})
}

func (h *LangGraphChatHandler) SubmitFeedback(c *gin.Context) {
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

func (h *LangGraphChatHandler) GetFeedbackStats(c *gin.Context) {
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

func (h *LangGraphChatHandler) processSpecialCommands(message, role string, response *ChatResponse) error {
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