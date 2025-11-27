package handler

import (
	"net/http"

	"github.com/galyym/fcm_push/internal/model"
	"github.com/galyym/fcm_push/internal/service"
	"github.com/gin-gonic/gin"
)

type PushHandler struct {
	pushService *service.PushService
}

func NewPushHandler(pushService *service.PushService) *PushHandler {
	return &PushHandler{
		pushService: pushService,
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

	response, err := h.pushService.SendPush(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to send push",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
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

	response, err := h.pushService.SendBatchPush(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to send batch push",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
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
