package model

type PushRequest struct {
	Token    string            `json:"token" binding:"required"`
	Title    string            `json:"title" binding:"required"`
	Body     string            `json:"body" binding:"required"`
	Data     map[string]string `json:"data,omitempty"`
	Priority string            `json:"priority,omitempty"`
	ClientID string            `json:"client_id,omitempty"`
}

type PushResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

type BatchPushRequest struct {
	Notifications []PushRequest `json:"notifications" binding:"required,min=1,max=500"`
}
type BatchPushResponse struct {
	SuccessCount int            `json:"success_count"`
	FailureCount int            `json:"failure_count"`
	Results      []PushResponse `json:"results"`
}
