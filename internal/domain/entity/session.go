package entity

type SessionStatus string

const (
	StatusConnected    SessionStatus = "CONNECTED"
	StatusPairing      SessionStatus = "PAIRING"
	StatusDisconnected SessionStatus = "DISCONNECTED"
)

type DeviceInfo struct {
	JID       string `json:"jid,omitempty"`
	PushName  string `json:"push_name,omitempty"`
	Platform  string `json:"platform,omitempty"`
	Connected bool   `json:"connected"`
}
