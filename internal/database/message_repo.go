package database

import (
	"context"
	"time"

	"gorm.io/gorm"

	"case_study/internal/models"
)

type MessageRepo struct{ DB *gorm.DB }

func NewMessageRepo(db *gorm.DB) *MessageRepo { return &MessageRepo{DB: db} }

// ClaimNextPending marks up to limit messages as processed and returns them to avoid double-send
func (r *MessageRepo) ClaimNextPending(ctx context.Context, limit int) ([]models.Message, error) {
	var msgs []models.Message
	return msgs, r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Select ids first to avoid updating too many rows
		var ids []uint
		err := tx.Model(&models.Message{}).
			Where("status = ?", models.StatusPending).
			Order("id ASC").
			Limit(limit).
			Pluck("id", &ids).Error
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}

		// Mark as processing
		if err := tx.Model(&models.Message{}).
			Where("id IN ?", ids).
			Updates(map[string]any{"status": models.StatusProcessing}).Error; err != nil {
			return err
		}

		return tx.Where("id IN ?", ids).Find(&msgs).Error
	})
}

func (r *MessageRepo) MarkSent(ctx context.Context, id uint, remoteID string, at time.Time) error {
	return r.DB.WithContext(ctx).Model(&models.Message{}).Where("id = ?", id).
		Updates(map[string]any{
			"status":            models.StatusSent,
			"sent_at":           at,
			"remote_message_id": remoteID,
			"last_error":        nil,
		}).Error
}

func (r *MessageRepo) MarkFailed(ctx context.Context, id uint, errMsg string) error {
	return r.DB.WithContext(ctx).Model(&models.Message{}).Where("id = ?", id).
		Updates(map[string]any{
			"status":     models.StatusFailed,
			"last_error": errMsg,
		}).Error
}

func (r *MessageRepo) ListSent(ctx context.Context, limit int) ([]models.Message, error) {
	var msgs []models.Message
	err := r.DB.WithContext(ctx).Where("status = ?", models.StatusSent).Order("sent_at DESC").Limit(limit).Find(&msgs).Error
	return msgs, err
}

func (r *MessageRepo) Create(ctx context.Context, to, content string, status models.MessageStatus) (models.Message, error) {
	m := models.Message{To: to, Content: content, Status: status}
	err := r.DB.WithContext(ctx).Create(&m).Error
	return m, err
}
