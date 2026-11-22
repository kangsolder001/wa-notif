package whatsapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

const (
	cmdStartSave = "/savegambar"
	cmdStopSave  = "/gambarselesai"
)

// imageCollector saves incoming images per sender between /savegambar and /gambarselesai.
type imageCollector struct {
	dir     string
	allowed map[string]bool // empty = any sender allowed
	mu      sync.Mutex
	active  map[string]int // sender user -> images saved in current session
}

func newImageCollector(dir string, allowedSenders []string) *imageCollector {
	allowed := make(map[string]bool, len(allowedSenders))
	for _, s := range allowedSenders {
		allowed[s] = true
	}
	return &imageCollector{dir: dir, allowed: allowed, active: make(map[string]int)}
}

// normalizeCommand turns "/Gambar Selesai" into "/gambarselesai".
func normalizeCommand(text string) string {
	return strings.ToLower(strings.Join(strings.Fields(text), ""))
}

func imageExt(mime string) string {
	switch strings.ToLower(strings.TrimSpace(strings.Split(mime, ";")[0])) {
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".jpg"
	}
}

func (a *WhatsmeowAdapter) handleIncomingMessage(v *events.Message) {
	if v.Info.IsFromMe || v.Info.IsGroup || a.images == nil {
		return
	}

	c := a.images
	sender := v.Info.Sender.User
	if len(c.allowed) > 0 && !c.allowed[sender] && !c.allowed[v.Info.SenderAlt.User] {
		return
	}

	text := v.Message.GetConversation()
	if text == "" {
		text = v.Message.GetExtendedTextMessage().GetText()
	}

	switch normalizeCommand(text) {
	case cmdStartSave:
		c.mu.Lock()
		c.active[sender] = 0
		c.mu.Unlock()
		a.reply(v, "Mode simpan gambar aktif. Kirim gambar sekarang, lalu ketik /gambarselesai jika sudah.")
		return
	case cmdStopSave:
		c.mu.Lock()
		n, ok := c.active[sender]
		delete(c.active, sender)
		c.mu.Unlock()
		if !ok {
			a.reply(v, "Mode simpan gambar belum aktif. Ketik /savegambar untuk mulai.")
			return
		}
		a.reply(v, fmt.Sprintf("Selesai. %d gambar tersimpan.", n))
		return
	}

	img := v.Message.GetImageMessage()
	if img == nil {
		return
	}

	c.mu.Lock()
	_, ok := c.active[sender]
	c.mu.Unlock()
	if !ok {
		return
	}

	data, err := a.client.Download(context.Background(), img)
	if err != nil {
		a.logger.Errorf("Failed to download image %s from %s: %v", v.Info.ID, sender, err)
		a.reply(v, "Gagal mengunduh gambar, silakan kirim ulang.")
		return
	}

	dir := filepath.Join(c.dir, sender)
	if err := os.MkdirAll(dir, 0755); err != nil {
		a.logger.Errorf("Failed to create image directory %s: %v", dir, err)
		a.reply(v, "Gagal menyimpan gambar.")
		return
	}

	name := v.Info.Timestamp.Format("20060102_150405") + "_" + v.Info.ID + imageExt(img.GetMimetype())
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0644); err != nil {
		a.logger.Errorf("Failed to write image %s: %v", path, err)
		a.reply(v, "Gagal menyimpan gambar.")
		return
	}

	c.mu.Lock()
	if _, ok := c.active[sender]; ok {
		c.active[sender]++
	}
	c.mu.Unlock()
	a.logger.Infof("Saved image from %s to %s", sender, path)
}

func (a *WhatsmeowAdapter) reply(v *events.Message, text string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	a.applyRateLimit()
	if _, err := a.client.SendMessage(ctx, v.Info.Chat, &waProto.Message{Conversation: proto.String(text)}); err != nil {
		a.logger.Errorf("Failed to send reply to %s: %v", v.Info.Chat, err)
	}
}
