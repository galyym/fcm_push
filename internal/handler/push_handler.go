package handler

import (
	"net/http"

	"github.com/galyym/fcm_push/internal/model"
	"github.com/galyym/fcm_push/internal/service"
	"github.com/gin-gonic/gin"
)

type PushHandler struct {
	pushService  *service.PushService
	queueService *service.QueueService
}

func NewPushHandler(pushService *service.PushService, queueService *service.QueueService) *PushHandler {
	return &PushHandler{
		pushService:  pushService,
		queueService: queueService,
	}
}

// SendPush обрабатывает запрос на отправку одного push-уведомления
// @Summary Отправить push-уведомление
// @Description Отправляет push-уведомление на указанное устройство
// @Tags push
// @Accept json
// @Produce json
// @Param request body model.PushRequest true "Push request"
// @Success 200 {object} model.PushResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/push/send [post]
func (h *PushHandler) SendPush(c *gin.Context) {
	var req model.PushRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	if req.Priority == "" {
		req.Priority = "normal"
	}

	// Enqueue push notification instead of sending directly
	queueReq := &model.CreateQueueTaskRequest{
		Token:    req.Token,
		Title:    req.Title,
		Body:     req.Body,
		Data:     req.Data,
		Priority: req.Priority,
		ClientID: req.ClientID,
	}

	task, err := h.queueService.EnqueuePush(c.Request.Context(), queueReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to enqueue push",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"queue_task_id": task.ID,
		"status":        task.Status,
		"message":       "Push notification queued successfully",
	})
}

// SendBatchPush обрабатывает запрос на отправку нескольких push-уведомлений
// @Summary Отправить batch push-уведомлений
// @Description Отправляет несколько push-уведомлений за один запрос
// @Tags push
// @Accept json
// @Produce json
// @Param request body model.BatchPushRequest true "Batch push request"
// @Success 200 {object} model.BatchPushResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/push/send-batch [post]
func (h *PushHandler) SendBatchPush(c *gin.Context) {
	var req model.BatchPushRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	queueTasks := make([]model.CreateQueueTaskRequest, len(req.Notifications))
	for i, notification := range req.Notifications {
		queueTasks[i] = model.CreateQueueTaskRequest{
			Token:    notification.Token,
			Title:    notification.Title,
			Body:     notification.Body,
			Data:     notification.Data,
			Priority: notification.Priority,
			ClientID: notification.ClientID,
		}
	}

	tasks, err := h.queueService.EnqueueBatchPush(c.Request.Context(), queueTasks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to enqueue batch push",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"queued_count": len(tasks),
		"tasks":        tasks,
		"message":      "Batch push notifications queued successfully",
	})
}

// HealthCheck проверка здоровья сервиса
// @Summary Health check
// @Description Проверка работоспособности сервиса
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *PushHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "ok",
		Service: "fcm-push-service",
	})
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}
