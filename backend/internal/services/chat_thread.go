package services

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"time"

	"healthsecure/internal/models"

	"gorm.io/gorm"
)

type ChatThreadService struct {
	db           *gorm.DB
	auditService *AuditService
}

type ThreadConfig struct {
	ThreadID string            `json:"thread_id"`
	Metadata map[string]string `json:"metadata"`
	RunName  string            `json:"run_name"`
}

type FeedbackRequest struct {
	MessageID string `json:"message_id" binding:"required"`
	Feedback  string `json:"feedback" binding:"required,oneof=thumbs_up thumbs_down"`
}

func NewChatThreadService(db *gorm.DB, auditService *AuditService) *ChatThreadService {
	service := &ChatThreadService{
		db:           db,
		auditService: auditService,
	}

	return service
}

// GenerateThreadID creates a new unique thread ID
func (s *ChatThreadService) GenerateThreadID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("thread_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("thread_%x", bytes)
}

// CreateThread creates a new chat thread
func (s *ChatThreadService) CreateThread(userID string, title string) (*models.ChatThread, error) {
	threadID := s.GenerateThreadID()
	
	thread := &models.ChatThread{
		ThreadID:  threadID,
		UserID:    userID,
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	if err := s.db.Create(thread).Error; err != nil {
		return nil, fmt.Errorf("failed to create thread: %w", err)
	}

	// Log thread creation
	userIDUint, _ := strconv.ParseUint(userID, 10, 32)
	s.auditService.LogUserAction(
		uint(userIDUint),
		models.ActionCreate,
		fmt.Sprintf("chat_thread:%s", threadID),
		"",
		"",
		true,
		fmt.Sprintf("Created new chat thread: %s", threadID),
	)

	return thread, nil
}

// GetThread retrieves a thread by ID
func (s *ChatThreadService) GetThread(threadID, userID string) (*models.ChatThread, error) {
	var thread models.ChatThread
	err := s.db.Where("thread_id = ? AND user_id = ?", threadID, userID).First(&thread).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}
	return &thread, nil
}

// GetUserThreads gets all threads for a user
func (s *ChatThreadService) GetUserThreads(userID string, limit int) ([]models.ChatThread, error) {
	var threads []models.ChatThread
	query := s.db.Where("user_id = ? AND is_active = ?", userID, true).
		Order("updated_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&threads).Error; err != nil {
		return nil, fmt.Errorf("failed to get user threads: %w", err)
	}

	return threads, nil
}

// SaveMessage saves a message to a thread
func (s *ChatThreadService) SaveMessage(threadID, role, content, runID string) (*models.ChatMessage, error) {
	// Generate unique message ID
	messageID := s.generateMessageID()

	message := &models.ChatMessage{
		ThreadID:  threadID,
		MessageID: messageID,
		Role:      role,
		Content:   content,
		RunID:     runID,
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(message).Error; err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// Update thread's updated_at timestamp
	s.db.Model(&models.ChatThread{}).Where("thread_id = ?", threadID).Update("updated_at", time.Now())

	return message, nil
}

// GetThreadMessages retrieves all messages for a thread
func (s *ChatThreadService) GetThreadMessages(threadID, userID string, limit int) ([]models.ChatMessage, error) {
	// First verify user owns the thread
	if _, err := s.GetThread(threadID, userID); err != nil {
		return nil, err
	}

	var messages []models.ChatMessage
	query := s.db.Where("thread_id = ?", threadID).Order("created_at DESC").Limit(limit)

	if err := query.Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get thread messages: %w", err)
	}

	// Reverse the order to be ascending
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetThreadMessagesPage retrieves all messages for a thread with pagination
func (s *ChatThreadService) GetThreadMessagesPage(threadID, userID string, page, pageSize int) ([]models.ChatMessage, int64, error) {
	// First verify user owns the thread
	if _, err := s.GetThread(threadID, userID); err != nil {
		return nil, 0, err
	}

	var messages []models.ChatMessage
	var total int64

	query := s.db.Model(&models.ChatMessage{}).Where("thread_id = ?", threadID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count thread messages: %w", err)
	}

	// Paginate
	offset := (page - 1) * pageSize
	if err := query.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get thread messages: %w", err)
	}

	return messages, total, nil
}

// CreateThreadConfig creates a LangSmith-compatible thread config
func (s *ChatThreadService) CreateThreadConfig(threadID, userID string) *ThreadConfig {
	return &ThreadConfig{
		ThreadID: threadID,
		Metadata: map[string]string{
			"thread_id": threadID,
			"user_id":   userID,
			"app":       "healthsecure",
		},
		RunName: "chat_turn",
	}
}

// SubmitFeedback records feedback for a message
func (s *ChatThreadService) SubmitFeedback(messageID, userID, feedback string) error {
	// Verify the message belongs to the user's thread
	var message models.ChatMessage
	if err := s.db.Joins("JOIN chat_threads ON chat_messages.thread_id = chat_threads.thread_id").
		Where("chat_messages.message_id = ? AND chat_threads.user_id = ?", messageID, userID).
		First(&message).Error; err != nil {
		return fmt.Errorf("message not found or access denied: %w", err)
	}

	// Update feedback
	now := time.Now()
	if err := s.db.Model(&message).Updates(map[string]interface{}{
		"feedback":    feedback,
		"feedback_at": &now,
	}).Error; err != nil {
		return fmt.Errorf("failed to save feedback: %w", err)
	}

	// Log feedback submission
	userIDUint, _ := strconv.ParseUint(userID, 10, 32)
	s.auditService.LogUserAction(
		uint(userIDUint),
		models.ActionUpdate,
		fmt.Sprintf("chat_message:%s", messageID),
		"",
		"",
		true,
		fmt.Sprintf("Submitted %s feedback for message %s", feedback, messageID),
	)

	return nil
}

// GetMessageFeedback retrieves feedback statistics
func (s *ChatThreadService) GetMessageFeedback(userID string, days int) (map[string]int, error) {
	var results []struct {
		Feedback string
		Count    int
	}

	query := s.db.Table("chat_messages").
		Select("feedback, COUNT(*) as count").
		Joins("JOIN chat_threads ON chat_messages.thread_id = chat_threads.thread_id").
		Where("chat_threads.user_id = ? AND feedback IS NOT NULL", userID).
		Group("feedback")

	if days > 0 {
		query = query.Where("chat_messages.feedback_at > ?", time.Now().AddDate(0, 0, -days))
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get feedback statistics: %w", err)
	}

	stats := make(map[string]int)
	for _, result := range results {
		stats[result.Feedback] = result.Count
	}

	return stats, nil
}

// ArchiveThread marks a thread as inactive
func (s *ChatThreadService) ArchiveThread(threadID, userID string) error {
	result := s.db.Model(&models.ChatThread{}).
		Where("thread_id = ? AND user_id = ?", threadID, userID).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to archive thread: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("thread not found or access denied")
	}

	// Log thread archival
	userIDUint, _ := strconv.ParseUint(userID, 10, 32)
	s.auditService.LogUserAction(
		uint(userIDUint),
		models.ActionUpdate,
		fmt.Sprintf("chat_thread:%s", threadID),
		"",
		"",
		true,
		fmt.Sprintf("Archived chat thread: %s", threadID),
	)

	return nil
}

// CleanupOldThreads removes inactive threads older than specified days
func (s *ChatThreadService) CleanupOldThreads(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	
	// Delete messages first (foreign key constraint)
	if err := s.db.Where("thread_id IN (SELECT thread_id FROM chat_threads WHERE is_active = ? AND updated_at < ?)", false, cutoff).
		Delete(&models.ChatMessage{}).Error; err != nil {
		return fmt.Errorf("failed to delete old messages: %w", err)
	}

	// Delete threads
	if err := s.db.Where("is_active = ? AND updated_at < ?", false, cutoff).
		Delete(&models.ChatThread{}).Error; err != nil {
		return fmt.Errorf("failed to delete old threads: %w", err)
	}

	return nil
}

// generateMessageID creates a unique message ID
func (s *ChatThreadService) generateMessageID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("msg_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("msg_%x", bytes)
}

// UpdateThreadTitle updates the title of a chat thread
func (s *ChatThreadService) UpdateThreadTitle(threadID, userID, newTitle string) error {
	// First verify user owns the thread
	if _, err := s.GetThread(threadID, userID); err != nil {
		return err
	}

	result := s.db.Model(&models.ChatThread{}).
		Where("thread_id = ? AND user_id = ?", threadID, userID).
		Update("title", newTitle)

	if result.Error != nil {
		return fmt.Errorf("failed to update thread title: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("thread not found or access denied")
	}

	return nil
}
