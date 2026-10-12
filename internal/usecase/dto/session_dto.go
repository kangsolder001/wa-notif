package dto

type SessionStatusResponse struct {
	Status    string `json:"status"`
	Connected bool   `json:"connected"`
	JID       string `json:"jid,omitempty"`
	PushName  string `json:"push_name,omitempty"`
	Platform  string `json:"platform,omitempty"`
}

type QRCodeResponse struct {
	QRCode     string `json:"qr_code,omitempty"`
	QRCodeData string `json:"qr_code_data,omitempty"` // base64 PNG data URI: data:image/png;base64,...
	Connected  bool   `json:"connected"`
}
