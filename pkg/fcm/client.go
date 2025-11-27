package fcm

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type Client struct {
	messagingClient *messaging.Client
}

func NewClient(ctx context.Context, credentialsPath string) (*Client, error) {
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %w", err)
	}

	messagingClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting messaging client: %w", err)
	}

	return &Client{
		messagingClient: messagingClient,
	}, nil
}

func (c *Client) SendNotification(ctx context.Context, token, title, body string, data map[string]string, priority string) (string, error) {
	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
	}

	if priority == "high" {
		message.Android = &messaging.AndroidConfig{
			Priority: "high",
		}
		message.APNS = &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
		}
	}

	messageID, err := c.messagingClient.Send(ctx, message)
	if err != nil {
		return "", fmt.Errorf("error sending message: %w", err)
	}

	return messageID, nil
}

func (c *Client) SendBatchNotifications(ctx context.Context, messages []*messaging.Message) (*messaging.BatchResponse, error) {
	br, err := c.messagingClient.SendEach(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("error sending batch messages: %w", err)
	}

	return br, nil
}

func (c *Client) SubscribeToTopic(ctx context.Context, tokens []string, topic string) error {
	_, err := c.messagingClient.SubscribeToTopic(ctx, tokens, topic)
	if err != nil {
		return fmt.Errorf("error subscribing to topic: %w", err)
	}
	return nil
}

func (c *Client) UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) error {
	_, err := c.messagingClient.UnsubscribeFromTopic(ctx, tokens, topic)
	if err != nil {
		return fmt.Errorf("error unsubscribing from topic: %w", err)
	}
	return nil
}
