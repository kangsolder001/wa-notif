package phone

import (
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasErr   bool
	}{
		{"081234567890", "6281234567890", false},
		{"+62 812-3456-7890", "6281234567890", false},
		{"81234567890", "6281234567890", false},
		{"6281234567890", "6281234567890", false},
		{"12036302482394@g.us", "12036302482394@g.us", false},
		{"123", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		got, err := Normalize(tt.input)
		if tt.hasErr && err == nil {
			t.Errorf("Normalize(%q) expected error, got nil", tt.input)
		}
		if !tt.hasErr && err != nil {
			t.Errorf("Normalize(%q) unexpected error: %v", tt.input, err)
		}
		if got != tt.expected {
			t.Errorf("Normalize(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestToJID(t *testing.T) {
	jid, err := ToJID("08123456789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "628123456789@s.whatsapp.net"
	if jid != expected {
		t.Errorf("ToJID = %q; want %q", jid, expected)
	}
}
