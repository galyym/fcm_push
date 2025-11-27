package model

// PushRequest представляет запрос на отправку push-уведомления
type PushRequest struct {
	// Token устройства FCM
	Token string `json:"token" binding:"required"`

	// Заголовок уведомления
	Title string `json:"title" binding:"required"`

	// Тело уведомления
	Body string `json:"body" binding:"required"`

	// Дополнительные данные
	Data map[string]string `json:"data,omitempty"`

	// Приоритет (high/normal)
	Priority string `json:"priority,omitempty"`

	// ID клиента (для логирования/аналитики)
	ClientID string `json:"client_id,omitempty"`
}

// PushResponse представляет ответ на запрос отправки push
type PushResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// BatchPushRequest для отправки нескольких уведомлений
type BatchPushRequest struct {
	Notifications []PushRequest `json:"notifications" binding:"required,min=1,max=500"`
}

// BatchPushResponse ответ на batch запрос
type BatchPushResponse struct {
	SuccessCount int            `json:"success_count"`
	FailureCount int            `json:"failure_count"`
	Results      []PushResponse `json:"results"`
}
