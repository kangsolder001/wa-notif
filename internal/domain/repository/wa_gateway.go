package repository

import (
	"context"
	"wa-notif/internal/domain/entity"
)

type WhatsAppGateway interface {
	Connect(ctx context.Context) error
	Disconnect()
	IsConnected() bool
	GetSessionStatus() entity.SessionStatus
	GetDeviceInfo() *entity.DeviceInfo
	GetLatestQRCode() (string, bool)
	SendMessage(ctx context.Context, msg *entity.Message) (string, error)
	SendMedia(ctx context.Context, msg *entity.Message) (string, error)
	Logout(ctx context.Context) error
}
