package service

import (
	"context"
	"fmt"
	"log"

	"github.com/galyym/fcm_push/internal/model"
	"github.com/galyym/fcm_push/internal/repository"
	"github.com/google/uuid"
)

type QueueService struct {
	repo *repository.QueueRepository
}

func NewQueueService(repo *repository.QueueRepository) *QueueService {
	return &QueueService{
		repo: repo,
	}
}

func (s *QueueService) EnqueuePush(ctx context.Context, req *model.CreateQueueTaskRequest) (*model.QueueTaskResponse, error) {
	log.Printf("Enqueueing push notification for client: %s", req.ClientID)

	task, err := s.repo.CreateTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue push: %w", err)
	}

	log.Printf("Push notification enqueued successfully, task ID: %s", task.ID)

	return &model.QueueTaskResponse{
		ID:          task.ID,
		Status:      task.Status,
		ClientID:    task.ClientID,
		Attempts:    task.Attempts,
		MaxAttempts: task.MaxAttempts,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}, nil
}

func (s *QueueService) EnqueueBatchPush(ctx context.Context, notifications []model.CreateQueueTaskRequest) ([]model.QueueTaskResponse, error) {
	log.Printf("Enqueueing batch push notifications, count: %d", len(notifications))

	responses := make([]model.QueueTaskResponse, 0, len(notifications))

	for i, req := range notifications {
		task, err := s.repo.CreateTask(ctx, &req)
		if err != nil {
			log.Printf("Failed to enqueue notification %d: %v", i, err)
			// Continue with other notifications
			responses = append(responses, model.QueueTaskResponse{
				Status:       model.StatusFailed,
				ErrorMessage: stringPtr(err.Error()),
			})
			continue
		}

		responses = append(responses, model.QueueTaskResponse{
			ID:          task.ID,
			Status:      task.Status,
			ClientID:    task.ClientID,
			Attempts:    task.Attempts,
			MaxAttempts: task.MaxAttempts,
			CreatedAt:   task.CreatedAt,
			UpdatedAt:   task.UpdatedAt,
		})
	}

	log.Printf("Batch push notifications enqueued, total: %d", len(responses))
	return responses, nil
}

func (s *QueueService) GetTaskStatus(ctx context.Context, taskID uuid.UUID) (*model.QueueTaskResponse, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task status: %w", err)
	}

	return &model.QueueTaskResponse{
		ID:           task.ID,
		Status:       task.Status,
		Token:        task.Token,
		Title:        task.Title,
		Body:         task.Body,
		ClientID:     task.ClientID,
		Attempts:     task.Attempts,
		MaxAttempts:  task.MaxAttempts,
		ErrorMessage: task.ErrorMessage,
		FCMMessageID: task.FCMMessageID,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
	}, nil
}

func (s *QueueService) GetHistory(ctx context.Context, req *model.QueueHistoryRequest) (*model.QueueHistoryResponse, error) {
	return s.repo.GetHistory(ctx, req)
}

func (s *QueueService) GetStats(ctx context.Context) (*model.QueueStatsResponse, error) {
	return s.repo.GetStats(ctx)
}

func stringPtr(s string) *string {
	return &s
}
