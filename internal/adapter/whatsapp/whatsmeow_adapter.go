package whatsapp

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	qrcode "github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	_ "modernc.org/sqlite"

	"wa-notif/internal/domain/entity"
	"wa-notif/internal/domain/repository"
)

var _ repository.WhatsAppGateway = (*WhatsmeowAdapter)(nil)

type WhatsmeowAdapter struct {
	client    *whatsmeow.Client
	container *sqlstore.Container
	status    entity.SessionStatus
	latestQR  string
	qrMu      sync.RWMutex
	sendMu    sync.Mutex
	lastSent  time.Time
	rateLimit time.Duration
	logger    waLog.Logger
}

func NewWhatsmeowAdapter(dbPath string, rateLimitSeconds float64, logLevel string) (*WhatsmeowAdapter, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory %s: %w", dir, err)
	}

	wlog := waLog.Stdout("Whatsmeow", logLevel, true)

	// Open SQLite using pure-Go driver
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	container := sqlstore.NewWithDB(db, "sqlite3", wlog)
	if err := container.Upgrade(context.Background()); err != nil {
		_ = container.Close()
		return nil, fmt.Errorf("failed to upgrade sqlite schema: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		_ = container.Close()
		return nil, fmt.Errorf("failed to get whatsapp device store: %w", err)
	}

	client := whatsmeow.NewClient(deviceStore, wlog)

	adapter := &WhatsmeowAdapter{
		client:    client,
		container: container,
		status:    entity.StatusDisconnected,
		rateLimit: time.Duration(rateLimitSeconds * float64(time.Second)),
		logger:    wlog,
	}

	client.AddEventHandler(adapter.handleEvent)

	return adapter, nil
}

func (a *WhatsmeowAdapter) handleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Connected:
		a.qrMu.Lock()
		a.status = entity.StatusConnected
		a.latestQR = ""
		a.qrMu.Unlock()
		a.logger.Infof("WhatsApp connected successfully! JID: %s", a.client.Store.ID.String())

	case *events.LoggedOut:
		a.qrMu.Lock()
		a.status = entity.StatusDisconnected
		a.latestQR = ""
		a.qrMu.Unlock()
		a.logger.Warnf("WhatsApp device logged out. Reason: %v", v.Reason)

	case *events.StreamReplaced:
		a.qrMu.Lock()
		a.status = entity.StatusDisconnected
		a.qrMu.Unlock()
		a.logger.Warnf("WhatsApp stream replaced (logged in on another device or concurrent session)")
	}
}

func (a *WhatsmeowAdapter) Connect(ctx context.Context) error {
	if a.client.Store.ID == nil {
		// Device is not paired yet, request QR channel with background context
		qrChan, err := a.client.GetQRChannel(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get qr channel: %w", err)
		}

		err = a.client.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect whatsapp client: %w", err)
		}

		go func() {
			for evt := range qrChan {
				switch evt.Event {
				case "code":
					a.qrMu.Lock()
					a.latestQR = evt.Code
					a.status = entity.StatusPairing
					a.qrMu.Unlock()

					a.logger.Infof("New QR code generated! Scan via Web UI or Terminal.")
					// Render terminal ASCII QR
					if qrObj, err := qrcode.New(evt.Code, qrcode.Medium); err == nil {
						fmt.Println("\n" + qrObj.ToSmallString(false))
					}
				case "success":
					a.qrMu.Lock()
					a.status = entity.StatusConnected
					a.latestQR = ""
					a.qrMu.Unlock()
					a.logger.Infof("QR pairing successful!")
				case "timeout":
					a.qrMu.Lock()
					a.latestQR = ""
					a.status = entity.StatusDisconnected
					a.qrMu.Unlock()
					a.logger.Warnf("QR pairing timed out.")
				}
			}
		}()
	} else {
		// Already paired, reconnect
		err := a.client.Connect()
		if err != nil {
			return fmt.Errorf("failed to reconnect whatsapp client: %w", err)
		}
		a.qrMu.Lock()
		a.status = entity.StatusConnected
		a.qrMu.Unlock()
	}

	return nil
}

func (a *WhatsmeowAdapter) Disconnect() {
	a.client.Disconnect()
	a.qrMu.Lock()
	a.status = entity.StatusDisconnected
	a.qrMu.Unlock()
}

func (a *WhatsmeowAdapter) IsConnected() bool {
	return a.client.IsConnected() && a.client.IsLoggedIn()
}

func (a *WhatsmeowAdapter) GetSessionStatus() entity.SessionStatus {
	a.qrMu.RLock()
	defer a.qrMu.RUnlock()

	if a.IsConnected() {
		return entity.StatusConnected
	}
	if a.status == entity.StatusPairing {
		return entity.StatusPairing
	}
	return entity.StatusDisconnected
}

func (a *WhatsmeowAdapter) GetDeviceInfo() *entity.DeviceInfo {
	if a.client.Store.ID == nil {
		return &entity.DeviceInfo{Connected: false}
	}

	return &entity.DeviceInfo{
		JID:       a.client.Store.ID.String(),
		PushName:  a.client.Store.PushName,
		Platform:  a.client.Store.Platform,
		Connected: a.IsConnected(),
	}
}

func (a *WhatsmeowAdapter) GetLatestQRCode() (string, bool) {
	a.qrMu.Lock()
	defer a.qrMu.Unlock()

	if a.latestQR != "" {
		return a.latestQR, true
	}

	// Jika belum connected dan belum ada QR / QR expired, reconnect untuk minta QR baru
	if !a.IsConnected() && a.client.Store.ID == nil {
		go func() {
			a.client.Disconnect()
			_ = a.Connect(context.Background())
		}()
	}

	return "", false
}

func (a *WhatsmeowAdapter) applyRateLimit() {
	a.sendMu.Lock()
	defer a.sendMu.Unlock()

	if !a.lastSent.IsZero() {
		elapsed := time.Since(a.lastSent)
		if elapsed < a.rateLimit {
			time.Sleep(a.rateLimit - elapsed)
		}
	}
	a.lastSent = time.Now()
}

func (a *WhatsmeowAdapter) SendMessage(ctx context.Context, msg *entity.Message) (string, error) {
	targetJID, err := types.ParseJID(msg.To)
	if err != nil {
		return "", fmt.Errorf("invalid recipient JID %s: %w", msg.To, err)
	}

	a.applyRateLimit()

	protoMsg := &waProto.Message{
		Conversation: proto.String(msg.Body),
	}

	resp, err := a.client.SendMessage(ctx, targetJID, protoMsg)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (a *WhatsmeowAdapter) SendMedia(ctx context.Context, msg *entity.Message) (string, error) {
	targetJID, err := types.ParseJID(msg.To)
	if err != nil {
		return "", fmt.Errorf("invalid recipient JID %s: %w", msg.To, err)
	}

	a.applyRateLimit()

	var mediaType whatsmeow.MediaType
	if msg.Type == entity.MessageTypeImage {
		mediaType = whatsmeow.MediaImage
	} else {
		mediaType = whatsmeow.MediaDocument
	}

	uploaded, err := a.client.Upload(ctx, msg.MediaData, mediaType)
	if err != nil {
		return "", fmt.Errorf("failed to upload media to whatsapp servers: %w", err)
	}

	var protoMsg *waProto.Message

	if msg.Type == entity.MessageTypeImage {
		protoMsg = &waProto.Message{
			ImageMessage: &waProto.ImageMessage{
				Caption:       proto.String(msg.Body),
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				Mimetype:      proto.String(msg.MimeType),
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uploaded.FileLength),
			},
		}
	} else {
		protoMsg = &waProto.Message{
			DocumentMessage: &waProto.DocumentMessage{
				Title:         proto.String(msg.FileName),
				FileName:      proto.String(msg.FileName),
				Caption:       proto.String(msg.Body),
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				Mimetype:      proto.String(msg.MimeType),
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uploaded.FileLength),
			},
		}
	}

	resp, err := a.client.SendMessage(ctx, targetJID, protoMsg)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (a *WhatsmeowAdapter) Logout(ctx context.Context) error {
	if a.client.Store.ID != nil {
		err := a.client.Logout(ctx)
		if err != nil {
			return err
		}
	}
	a.qrMu.Lock()
	a.status = entity.StatusDisconnected
	a.latestQR = ""
	a.qrMu.Unlock()
	return nil
}
