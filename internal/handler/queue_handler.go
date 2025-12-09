package handler

import (
	"net/http"

	"github.com/galyym/fcm_push/internal/model"
	"github.com/galyym/fcm_push/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type QueueHandler struct {
	queueService *service.QueueService
}

func NewQueueHandler(queueService *service.QueueService) *QueueHandler {
	return &QueueHandler{
		queueService: queueService,
	}
}

func (h *QueueHandler) GetTaskStatus(c *gin.Context) {
	taskIDStr := c.Param("id")

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID format",
		})
		return
	}

	task, err := h.queueService.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *QueueHandler) GetHistory(c *gin.Context) {
	var req model.QueueHistoryRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid query parameters",
		})
		return
	}

	history, err := h.queueService.GetHistory(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve history",
		})
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *QueueHandler) GetStats(c *gin.Context) {
	stats, err := h.queueService.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
