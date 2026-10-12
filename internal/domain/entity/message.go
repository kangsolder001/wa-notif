package entity

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyMessage = errors.New("message content cannot be empty")
	ErrEmptyRecipient = errors.New("recipient phone number cannot be empty")
)

type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeDocument MessageType = "document"
)

type Message struct {
	ID        string      `json:"id"`
	To        string      `json:"to"`
	Body      string      `json:"body"`
	Type      MessageType `json:"type"`
	MediaURL  string      `json:"media_url,omitempty"`
	MediaData []byte      `json:"-"`
	FileName  string      `json:"file_name,omitempty"`
	MimeType  string      `json:"mime_type,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
}

func NewTextMessage(to, body string) (*Message, error) {
	to = strings.TrimSpace(to)
	body = strings.TrimSpace(body)

	if to == "" {
		return nil, ErrEmptyRecipient
	}
	if body == "" {
		return nil, ErrEmptyMessage
	}

	return &Message{
		To:        to,
		Body:      body,
		Type:      MessageTypeText,
		CreatedAt: time.Now(),
	}, nil
}

func NewMediaMessage(to, caption, fileName, mimeType string, data []byte, msgType MessageType) (*Message, error) {
	to = strings.TrimSpace(to)
	if to == "" {
		return nil, ErrEmptyRecipient
	}
	if len(data) == 0 {
		return nil, errors.New("media data cannot be empty")
	}

	if msgType == "" {
		msgType = MessageTypeImage
	}

	return &Message{
		To:        to,
		Body:      caption,
		Type:      msgType,
		FileName:  fileName,
		MimeType:  mimeType,
		MediaData: data,
		CreatedAt: time.Now(),
	}, nil
}
