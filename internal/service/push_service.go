package service

import (
	"context"
	"fmt"
	"log"

	"firebase.google.com/go/v4/messaging"
	"github.com/galyym/fcm_push/internal/model"
	"github.com/galyym/fcm_push/pkg/fcm"
)

type PushService struct {
	fcmClient *fcm.Client
}

func NewPushService(fcmClient *fcm.Client) *PushService {
	return &PushService{
		fcmClient: fcmClient,
	}
}

func (s *PushService) SendPush(ctx context.Context, req *model.PushRequest) (*model.PushResponse, error) {
	log.Printf("Sending push to client: %s, token: %s...", req.ClientID, maskToken(req.Token))
	messageID, err := s.fcmClient.SendNotification(
		ctx,
		req.Token,
		req.Title,
		req.Body,
		req.Data,
		req.Priority,
	)

	if err != nil {
		log.Printf("Failed to send push: %v", err)
		return &model.PushResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	log.Printf("Push sent successfully, message ID: %s", messageID)
	return &model.PushResponse{
		Success:   true,
		MessageID: messageID,
	}, nil
}

func (s *PushService) SendBatchPush(ctx context.Context, req *model.BatchPushRequest) (*model.BatchPushResponse, error) {
	log.Printf("Sending batch push, count: %d", len(req.Notifications))

	messages := make([]*messaging.Message, 0, len(req.Notifications))
	for _, notification := range req.Notifications {
		msg := &messaging.Message{
			Token: notification.Token,
			Notification: &messaging.Notification{
				Title: notification.Title,
				Body:  notification.Body,
			},
			Data: notification.Data,
		}

		if notification.Priority == "high" {
			msg.Android = &messaging.AndroidConfig{
				Priority: "high",
			}
			msg.APNS = &messaging.APNSConfig{
				Headers: map[string]string{
					"apns-priority": "10",
				},
			}
		}

		messages = append(messages, msg)
	}

	batchResponse, err := s.fcmClient.SendBatchNotifications(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to send batch: %w", err)
	}
	response := &model.BatchPushResponse{
		SuccessCount: batchResponse.SuccessCount,
		FailureCount: batchResponse.FailureCount,
		Results:      make([]model.PushResponse, len(batchResponse.Responses)),
	}

	for i, resp := range batchResponse.Responses {
		if resp.Success {
			response.Results[i] = model.PushResponse{
				Success:   true,
				MessageID: resp.MessageID,
			}
		} else {
			response.Results[i] = model.PushResponse{
				Success: false,
				Error:   resp.Error.Error(),
			}
		}
	}

	log.Printf("Batch push completed. Success: %d, Failed: %d", response.SuccessCount, response.FailureCount)
	return response, nil
}

func maskToken(token string) string {
	if len(token) <= 10 {
		return "***"
	}
	return token[:5] + "..." + token[len(token)-5:]
}
