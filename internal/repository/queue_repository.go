package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/galyym/fcm_push/internal/database"
	"github.com/galyym/fcm_push/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// QueueRepository handles database operations for push queue
type QueueRepository struct {
	db *database.DB
}

// NewQueueRepository creates a new queue repository
func NewQueueRepository(db *database.DB) *QueueRepository {
	return &QueueRepository{db: db}
}

// CreateTask creates a new push notification task in the queue
func (r *QueueRepository) CreateTask(ctx context.Context, req *model.CreateQueueTaskRequest) (*model.PushQueueTask, error) {
	task := &model.PushQueueTask{
		ID:          uuid.New(),
		Token:       req.Token,
		Title:       req.Title,
		Body:        req.Body,
		Data:        req.Data,
		Priority:    req.Priority,
		ClientID:    req.ClientID,
		Status:      model.StatusPending,
		Attempts:    0,
		MaxAttempts: req.MaxAttempts,
		ScheduledAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if task.Priority == "" {
		task.Priority = "normal"
	}
	if task.MaxAttempts == 0 {
		task.MaxAttempts = 3
	}

	query := `
		INSERT INTO push_queue (
			id, token, title, body, data, priority, client_id,
			status, attempts, max_attempts, scheduled_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		RETURNING id, created_at, updated_at
	`

	err := r.db.Pool.QueryRow(
		ctx, query,
		task.ID, task.Token, task.Title, task.Body, task.Data, task.Priority, task.ClientID,
		task.Status, task.Attempts, task.MaxAttempts, task.ScheduledAt, task.CreatedAt, task.UpdatedAt,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

// GetTaskByID retrieves a task by its ID
func (r *QueueRepository) GetTaskByID(ctx context.Context, id uuid.UUID) (*model.PushQueueTask, error) {
	query := `
		SELECT id, token, title, body, data, priority, client_id,
		       status, attempts, max_attempts, error_message, fcm_message_id,
		       scheduled_at, created_at, updated_at
		FROM push_queue
		WHERE id = $1
	`

	task := &model.PushQueueTask{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&task.ID, &task.Token, &task.Title, &task.Body, &task.Data, &task.Priority, &task.ClientID,
		&task.Status, &task.Attempts, &task.MaxAttempts, &task.ErrorMessage, &task.FCMMessageID,
		&task.ScheduledAt, &task.CreatedAt, &task.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

// GetPendingTasks retrieves tasks that are ready to be processed
func (r *QueueRepository) GetPendingTasks(ctx context.Context, limit int) ([]*model.PushQueueTask, error) {
	query := `
		UPDATE push_queue
		SET status = $1, updated_at = NOW()
		WHERE id IN (
			SELECT id FROM push_queue
			WHERE status = $2
			  AND scheduled_at <= NOW()
			  AND attempts < max_attempts
			ORDER BY scheduled_at ASC
			LIMIT $3
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, token, title, body, data, priority, client_id,
		          status, attempts, max_attempts, error_message, fcm_message_id,
		          scheduled_at, created_at, updated_at
	`

	rows, err := r.db.Pool.Query(ctx, query, model.StatusProcessing, model.StatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*model.PushQueueTask
	for rows.Next() {
		task := &model.PushQueueTask{}
		err := rows.Scan(
			&task.ID, &task.Token, &task.Title, &task.Body, &task.Data, &task.Priority, &task.ClientID,
			&task.Status, &task.Attempts, &task.MaxAttempts, &task.ErrorMessage, &task.FCMMessageID,
			&task.ScheduledAt, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTaskSuccess updates a task as successfully completed
func (r *QueueRepository) UpdateTaskSuccess(ctx context.Context, id uuid.UUID, messageID string) error {
	query := `
		UPDATE push_queue
		SET status = $1, fcm_message_id = $2, updated_at = NOW()
		WHERE id = $3
	`

	_, err := r.db.Pool.Exec(ctx, query, model.StatusSuccess, messageID, id)
	if err != nil {
		return fmt.Errorf("failed to update task success: %w", err)
	}

	return nil
}

// UpdateTaskFailure updates a task as failed and schedules retry if attempts remain
func (r *QueueRepository) UpdateTaskFailure(ctx context.Context, id uuid.UUID, errorMsg string, nextRetry *time.Time) error {
	var query string
	var args []interface{}

	if nextRetry != nil {
		// Schedule retry
		query = `
			UPDATE push_queue
			SET attempts = attempts + 1,
			    error_message = $1,
			    scheduled_at = $2,
			    status = $3,
			    updated_at = NOW()
			WHERE id = $4
		`
		args = []interface{}{errorMsg, *nextRetry, model.StatusPending, id}
	} else {
		// Mark as permanently failed
		query = `
			UPDATE push_queue
			SET attempts = attempts + 1,
			    error_message = $1,
			    status = $2,
			    updated_at = NOW()
			WHERE id = $3
		`
		args = []interface{}{errorMsg, model.StatusFailed, id}
	}

	_, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update task failure: %w", err)
	}

	return nil
}

// GetHistory retrieves task history with filtering and pagination
func (r *QueueRepository) GetHistory(ctx context.Context, req *model.QueueHistoryRequest) (*model.QueueHistoryResponse, error) {
	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argPos := 1

	if req.ClientID != "" {
		conditions = append(conditions, fmt.Sprintf("client_id = $%d", argPos))
		args = append(args, req.ClientID)
		argPos++
	}

	if req.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argPos))
		args = append(args, req.Status)
		argPos++
	}

	if req.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argPos))
		args = append(args, *req.StartDate)
		argPos++
	}

	if req.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argPos))
		args = append(args, *req.EndDate)
		argPos++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM push_queue %s", whereClause)
	var total int
	err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Set defaults for pagination
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Get tasks
	query := fmt.Sprintf(`
		SELECT id, token, title, body, client_id, status, attempts, max_attempts,
		       error_message, fcm_message_id, created_at, updated_at
		FROM push_queue
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argPos, argPos+1)

	args = append(args, req.Limit, req.Offset)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	defer rows.Close()

	var tasks []model.QueueTaskResponse
	for rows.Next() {
		task := model.QueueTaskResponse{}
		err := rows.Scan(
			&task.ID, &task.Token, &task.Title, &task.Body, &task.ClientID,
			&task.Status, &task.Attempts, &task.MaxAttempts,
			&task.ErrorMessage, &task.FCMMessageID, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return &model.QueueHistoryResponse{
		Tasks:  tasks,
		Total:  total,
		Limit:  req.Limit,
		Offset: req.Offset,
	}, nil
}

// GetStats retrieves queue statistics
func (r *QueueRepository) GetStats(ctx context.Context) (*model.QueueStatsResponse, error) {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'pending') as pending_count,
			COUNT(*) FILTER (WHERE status = 'processing') as processing_count,
			COUNT(*) FILTER (WHERE status = 'success') as success_count,
			COUNT(*) FILTER (WHERE status = 'failed') as failed_count,
			COUNT(*) as total_count
		FROM push_queue
	`

	stats := &model.QueueStatsResponse{}
	err := r.db.Pool.QueryRow(ctx, query).Scan(
		&stats.PendingCount,
		&stats.ProcessingCount,
		&stats.SuccessCount,
		&stats.FailedCount,
		&stats.TotalCount,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return stats, nil
}

// CleanupOldTasks deletes tasks older than the specified duration
func (r *QueueRepository) CleanupOldTasks(ctx context.Context, olderThan time.Duration) (int64, error) {
	query := `
		DELETE FROM push_queue
		WHERE created_at < $1
		  AND status IN ('success', 'failed')
	`

	cutoffTime := time.Now().Add(-olderThan)
	result, err := r.db.Pool.Exec(ctx, query, cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old tasks: %w", err)
	}

	return result.RowsAffected(), nil
}
