package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"healthsecure/internal/models"
)

// AIServiceClient handles communication with the Python AI service
type AIServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// AIServiceRequest represents a request to the AI service
type AIServiceRequest struct {
	Message   string `json:"message"`
	ThreadID  string `json:"thread_id"`
	UserID    string `json:"user_id"`
	UserRole  string `json:"user_role"`
	UserName  string `json:"user_name"`
	History   []models.ChatMessage `json:"history,omitempty"`
	MaxTokens *int   `json:"max_tokens,omitempty"`
}

// AIServiceResponse represents a response from the AI service
type AIServiceResponse struct {
	Response   string `json:"response"`
	ThreadID   string `json:"thread_id"`
	RunID      string `json:"run_id"`
	ModelUsed  string `json:"model_used"`
	NewTitle   string `json:"new_title,omitempty"`
	TokensUsed *int   `json:"tokens_used,omitempty"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	Details    string `json:"details,omitempty"`
}

// NewAIServiceClient creates a new AI service client
func NewAIServiceClient() *AIServiceClient {
	baseURL := os.Getenv("AI_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:5000" // Default to local development
	}

	return &AIServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Second * 30, // 30 second timeout
		},
	}
}

// Chat sends a chat message to the AI service
func (c *AIServiceClient) Chat(req AIServiceRequest) (*AIServiceResponse, error) {
	// Prepare request payload
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", c.baseURL+"/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "HealthSecure-Go-Backend/1.0")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var aiResp AIServiceResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return &aiResp, fmt.Errorf("AI service error (status %d): %s", resp.StatusCode, aiResp.Error)
	}

	// Check for application errors
	if !aiResp.Success {
		return &aiResp, fmt.Errorf("AI service failed: %s - %s", aiResp.Error, aiResp.Details)
	}

	return &aiResp, nil
}

// HealthCheck checks if the AI service is healthy
func (c *AIServiceClient) HealthCheck() error {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AI service unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

// GetChatHistory retrieves chat history for a thread
func (c *AIServiceClient) GetChatHistory(threadID string) (map[string]interface{}, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/history/%s", c.baseURL, threadID))
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read history response: %w", err)
	}

	var history map[string]interface{}
	if err := json.Unmarshal(body, &history); err != nil {
		return nil, fmt.Errorf("failed to parse history: %w", err)
	}

	return history, nil
}