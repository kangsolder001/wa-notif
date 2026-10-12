package phone

import (
	"errors"
	"regexp"
	"strings"
)

var (
	nonNumericRegex = regexp.MustCompile(`[^0-9]`)
	ErrInvalidPhone = errors.New("invalid phone number: must contain at least 8 digits")
)

// Normalize cleans and normalizes a phone number to standard international format (defaulting 08xx to 628xx).
func Normalize(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidPhone
	}

	// If it's already a full WhatsApp JID (e.g., group ...@g.us or user ...@s.whatsapp.net), handle properly
	if strings.Contains(raw, "@") {
		parts := strings.Split(raw, "@")
		cleanedUser := nonNumericRegex.ReplaceAllString(parts[0], "")
		if len(cleanedUser) < 5 {
			return "", ErrInvalidPhone
		}
		return cleanedUser + "@" + parts[1], nil
	}

	// Remove all non-numeric characters
	cleaned := nonNumericRegex.ReplaceAllString(raw, "")

	if len(cleaned) < 8 {
		return "", ErrInvalidPhone
	}

	// Indonesian local number conversion (08xxx -> 628xxx)
	if strings.HasPrefix(cleaned, "0") {
		cleaned = "62" + cleaned[1:]
	} else if strings.HasPrefix(cleaned, "8") {
		// e.g. 8123456789 -> 628123456789
		cleaned = "62" + cleaned
	}

	return cleaned, nil
}

// ToJID converts a phone number to a WhatsApp JID string (e.g., 628123456789@s.whatsapp.net)
func ToJID(raw string) (string, error) {
	normalized, err := Normalize(raw)
	if err != nil {
		return "", err
	}

	if strings.Contains(normalized, "@") {
		return normalized, nil
	}

	return normalized + "@s.whatsapp.net", nil
}
