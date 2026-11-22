package whatsapp

import "testing"

func TestNormalizeCommand(t *testing.T) {
	cases := map[string]string{
		"/savegambar":      cmdStartSave,
		" /SaveGambar ":    cmdStartSave,
		"/gambar selesai":  cmdStopSave,
		"/Gambar  Selesai": cmdStopSave,
		"/gambarselesai":   cmdStopSave,
	}
	for in, want := range cases {
		if got := normalizeCommand(in); got != want {
			t.Errorf("normalizeCommand(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestImageExt(t *testing.T) {
	cases := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp", "": ".jpg"}
	for in, want := range cases {
		if got := imageExt(in); got != want {
			t.Errorf("imageExt(%q) = %q, want %q", in, got, want)
		}
	}
}
