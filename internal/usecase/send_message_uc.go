package usecase

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"wa-notif/internal/domain/entity"
	"wa-notif/internal/domain/repository"
	"wa-notif/internal/usecase/dto"
	"wa-notif/pkg/phone"
)

var (
	ErrNotConnected = errors.New("whatsapp service is not connected. please scan the qr code first")
)

type SendMessageUseCase struct {
	waGateway        repository.WhatsAppGateway
	defaultRecipient string
}

func NewSendMessageUseCase(waGateway repository.WhatsAppGateway, defaultRecipient string) *SendMessageUseCase {
	return &SendMessageUseCase{
		waGateway:        waGateway,
		defaultRecipient: defaultRecipient,
	}
}

func (uc *SendMessageUseCase) SendTextMessage(ctx context.Context, req dto.SendMessageRequest) (*dto.SendMessageResponse, error) {
	if !uc.waGateway.IsConnected() {
		return nil, ErrNotConnected
	}

	target := strings.TrimSpace(req.To)
	if target == "" {
		target = uc.defaultRecipient
	}

	if target == "" {
		return nil, errors.New("recipient phone number is required (no default recipient configured)")
	}

	jid, err := phone.ToJID(target)
	if err != nil {
		return nil, fmt.Errorf("invalid phone number %q: %w", target, err)
	}

	msg, err := entity.NewTextMessage(jid, req.Message)
	if err != nil {
		return nil, err
	}

	msgID, err := uc.waGateway.SendMessage(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send message via whatsapp: %w", err)
	}

	return &dto.SendMessageResponse{
		MessageID: msgID,
		To:        jid,
		Status:    "SENT",
		Timestamp: time.Now(),
	}, nil
}

func (uc *SendMessageUseCase) SendMediaMessage(ctx context.Context, req dto.SendMediaRequest) (*dto.SendMessageResponse, error) {
	if !uc.waGateway.IsConnected() {
		return nil, ErrNotConnected
	}

	target := strings.TrimSpace(req.To)
	if target == "" {
		target = uc.defaultRecipient
	}

	if target == "" {
		return nil, errors.New("recipient phone number is required (no default recipient configured)")
	}

	jid, err := phone.ToJID(target)
	if err != nil {
		return nil, fmt.Errorf("invalid phone number %q: %w", target, err)
	}

	var data []byte
	var mimeType = req.MimeType

	if req.DataBase64 != "" {
		// Clean potential prefix like "data:image/png;base64,"
		b64 := req.DataBase64
		if idx := strings.Index(b64, ","); idx != -1 {
			b64 = b64[idx+1:]
		}
		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 media data: %w", err)
		}
		data = decoded
	} else if req.MediaURL != "" {
		// Download from URL
		resp, err := http.Get(req.MediaURL)
		if err != nil {
			return nil, fmt.Errorf("failed to download media from URL: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("media URL returned status: %d", resp.StatusCode)
		}

		downloaded, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read media body: %w", err)
		}
		data = downloaded
		if mimeType == "" {
			mimeType = resp.Header.Get("Content-Type")
		}
	} else {
		return nil, errors.New("either data_base64 or media_url must be provided")
	}

	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}

	msgType := entity.MessageTypeImage
	if strings.ToLower(req.Type) == "document" || (!strings.HasPrefix(mimeType, "image/") && !strings.HasPrefix(mimeType, "video/")) {
		msgType = entity.MessageTypeDocument
	}

	fileName := req.FileName
	if fileName == "" {
		if msgType == entity.MessageTypeImage {
			fileName = "image.png"
		} else {
			fileName = "document.bin"
		}
	}

	msg, err := entity.NewMediaMessage(jid, req.Caption, fileName, mimeType, data, msgType)
	if err != nil {
		return nil, err
	}

	msgID, err := uc.waGateway.SendMedia(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send media via whatsapp: %w", err)
	}

	return &dto.SendMessageResponse{
		MessageID: msgID,
		To:        jid,
		Status:    "SENT",
		Timestamp: time.Now(),
	}, nil
}
