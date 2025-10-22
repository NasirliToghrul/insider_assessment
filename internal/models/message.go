package models

import (
	"time"
)

type MessageStatus string

const (
	StatusPending    MessageStatus = "pending"
	StatusProcessing MessageStatus = "processing"
	StatusSent       MessageStatus = "sent"
	StatusFailed     MessageStatus = "failed"
)

type Message struct {
	ID        uint          `gorm:"primaryKey" json:"id"`
	To        string        `gorm:"size:32;not null" json:"to"`
	Content   string        `gorm:"type:text;not null" json:"content"`
	Status    MessageStatus `gorm:"size:16;not null;index" json:"status"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
	SentAt    *time.Time    `json:"sentAt"`
	// Tracking from webhook response
	RemoteMessageID *string `gorm:"size:64" json:"remoteMessageId"`
	LastError       *string `gorm:"type:text" json:"lastError"`
}
