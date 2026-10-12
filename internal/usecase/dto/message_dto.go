package dto

import "time"

type SendMessageRequest struct {
	To      string `json:"to,omitempty"`
	Message string `json:"message"`
}

type SendMediaRequest struct {
	To       string `json:"to,omitempty"`
	Caption  string `json:"caption,omitempty"`
	Type     string `json:"type,omitempty"` // image or document
	FileName string `json:"file_name,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	MediaURL string `json:"media_url,omitempty"`  // optional remote URL
	DataBase64 string `json:"data_base64,omitempty"` // optional base64 encoded
}

type SendMessageResponse struct {
	MessageID string    `json:"message_id"`
	To        string    `json:"to"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
