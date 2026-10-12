package usecase

import (
	"context"
	"encoding/base64"
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
	"wa-notif/internal/domain/repository"
	"wa-notif/internal/usecase/dto"
)

type SessionUseCase struct {
	waGateway repository.WhatsAppGateway
}

func NewSessionUseCase(waGateway repository.WhatsAppGateway) *SessionUseCase {
	return &SessionUseCase{
		waGateway: waGateway,
	}
}

func (uc *SessionUseCase) GetStatus() dto.SessionStatusResponse {
	status := uc.waGateway.GetSessionStatus()
	connected := uc.waGateway.IsConnected()
	devInfo := uc.waGateway.GetDeviceInfo()

	resp := dto.SessionStatusResponse{
		Status:    string(status),
		Connected: connected,
	}

	if devInfo != nil {
		resp.JID = devInfo.JID
		resp.PushName = devInfo.PushName
		resp.Platform = devInfo.Platform
	}

	return resp
}

func (uc *SessionUseCase) GetQRCode() (dto.QRCodeResponse, error) {
	if uc.waGateway.IsConnected() {
		return dto.QRCodeResponse{Connected: true}, nil
	}

	qrCodeStr, ok := uc.waGateway.GetLatestQRCode()
	if !ok || qrCodeStr == "" {
		return dto.QRCodeResponse{Connected: false}, nil
	}

	// Generate PNG QR Code bytes
	pngBytes, err := qrcode.Encode(qrCodeStr, qrcode.Medium, 256)
	if err != nil {
		return dto.QRCodeResponse{}, fmt.Errorf("failed to encode qr code image: %w", err)
	}

	b64Image := "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes)

	return dto.QRCodeResponse{
		QRCode:     qrCodeStr,
		QRCodeData: b64Image,
		Connected:  false,
	}, nil
}

func (uc *SessionUseCase) Logout(ctx context.Context) error {
	return uc.waGateway.Logout(ctx)
}
