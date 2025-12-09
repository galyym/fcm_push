package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type QueueStatus string

const (
	StatusPending    QueueStatus = "pending"
	StatusProcessing QueueStatus = "processing"
	StatusSuccess    QueueStatus = "success"
	StatusFailed     QueueStatus = "failed"
)

type PushQueueTask struct {
	ID           uuid.UUID   `db:"id" json:"id"`
	Token        string      `db:"token" json:"token"`
	Title        string      `db:"title" json:"title"`
	Body         string      `db:"body" json:"body"`
	Data         JSONMap     `db:"data" json:"data,omitempty"`
	Priority     string      `db:"priority" json:"priority"`
	ClientID     string      `db:"client_id" json:"client_id,omitempty"`
	Status       QueueStatus `db:"status" json:"status"`
	Attempts     int         `db:"attempts" json:"attempts"`
	MaxAttempts  int         `db:"max_attempts" json:"max_attempts"`
	ErrorMessage *string     `db:"error_message" json:"error_message,omitempty"`
	FCMMessageID *string     `db:"fcm_message_id" json:"fcm_message_id,omitempty"`
	ScheduledAt  time.Time   `db:"scheduled_at" json:"scheduled_at"`
	CreatedAt    time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time   `db:"updated_at" json:"updated_at"`
}

type JSONMap map[string]string

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, j)
}

type CreateQueueTaskRequest struct {
	Token       string            `json:"token" binding:"required"`
	Title       string            `json:"title" binding:"required"`
	Body        string            `json:"body" binding:"required"`
	Data        map[string]string `json:"data,omitempty"`
	Priority    string            `json:"priority,omitempty"`
	ClientID    string            `json:"client_id,omitempty"`
	MaxAttempts int               `json:"max_attempts,omitempty"`
}

type QueueTaskResponse struct {
	ID           uuid.UUID   `json:"id"`
	Status       QueueStatus `json:"status"`
	Token        string      `json:"token,omitempty"`
	Title        string      `json:"title,omitempty"`
	Body         string      `json:"body,omitempty"`
	ClientID     string      `json:"client_id,omitempty"`
	Attempts     int         `json:"attempts"`
	MaxAttempts  int         `json:"max_attempts"`
	ErrorMessage *string     `json:"error_message,omitempty"`
	FCMMessageID *string     `json:"fcm_message_id,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type QueueHistoryRequest struct {
	ClientID  string      `form:"client_id"`
	Status    QueueStatus `form:"status"`
	StartDate *time.Time  `form:"start_date"`
	EndDate   *time.Time  `form:"end_date"`
	Limit     int         `form:"limit"`
	Offset    int         `form:"offset"`
}

type QueueHistoryResponse struct {
	Tasks  []QueueTaskResponse `json:"tasks"`
	Total  int                 `json:"total"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
}

type QueueStatsResponse struct {
	PendingCount    int `json:"pending_count"`
	ProcessingCount int `json:"processing_count"`
	SuccessCount    int `json:"success_count"`
	FailedCount     int `json:"failed_count"`
	TotalCount      int `json:"total_count"`
}
