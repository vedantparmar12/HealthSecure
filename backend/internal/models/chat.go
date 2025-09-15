package models

import (
	"time"
)

type ChatThread struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ThreadID  string    `json:"thread_id" gorm:"uniqueIndex;size:255"`
	UserID    string    `json:"user_id" gorm:"not null;index:idx_user_active,priority:1"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active" gorm:"default:true;index:idx_user_active,priority:2"`
}

func (ct *ChatThread) TableName() string {
	return "chat_threads"
}

type ChatMessage struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ThreadID   string    `json:"thread_id" gorm:"not null;index"`
	MessageID  string    `json:"message_id" gorm:"uniqueIndex;not null;size:255"`
	Role       string    `json:"role" gorm:"not null"` // "user" or "assistant"
	Content    string    `json:"content" gorm:"type:text"`
	RunID      string    `json:"run_id"` // LangSmith run ID
	CreatedAt  time.Time `json:"created_at"`
	Feedback   *string   `json:"feedback"` // "thumbs_up", "thumbs_down", or null
	FeedbackAt *time.Time `json:"feedback_at"`
}

func (cm *ChatMessage) TableName() string {
	return "chat_messages"
}
