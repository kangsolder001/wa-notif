# Panduan Penggunaan API (cURL Examples)

Semua endpoint private memerlukan header `Authorization: Bearer <token>` yang sesuai dengan `bearer_token` pada `config.json`.

---

## 1. Health Check (Public)
Mengecek apakah service aktif berjalan.

```bash
curl -X GET http://localhost:3030/health
```

**Contoh Response:**
```json
{
  "status": "UP"
}
```

---

## 2. Cek Status Koneksi WhatsApp
Memeriksa apakah WhatsApp sudah terhubung (`CONNECTED`), sedang proses pairing (`PAIRING`), atau terputus (`DISCONNECTED`).

```bash
curl -X GET http://localhost:3030/api/v1/session/status \
  -H "Authorization: Bearer your-secret-bearer-token"
```

**Contoh Response:**
```json
{
  "success": true,
  "data": {
    "status": "CONNECTED",
    "connected": true,
    "jid": "6281234567890:1@s.whatsapp.net",
    "push_name": "My Device",
    "platform": "Android"
  }
}
```

---

## 3. Ambil QR Code untuk Pairing
Mengambil data QR Code dalam format string dan Base64 Image (PNG Data URI).

```bash
curl -X GET http://localhost:3030/api/v1/session/qr \
  -H "Authorization: Bearer your-secret-bearer-token"
```

> **Catatan:** Anda juga bisa membuka **Web UI Pairing** langsung di browser: `http://localhost:3030/login` atau `http://<ip-rpi>:3030/login`.

---

## 4. Kirim Pesan Teks

### A. Kirim ke Nomor Tertentu:
```bash
curl -X POST http://localhost:3030/api/v1/messages/send \
  -H "Authorization: Bearer your-secret-bearer-token" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "081234567890",
    "message": "*[SERVER ALERT]* CPU load 92% on Raspberry Pi!\nPeriksa proses yang sedang berjalan."
  }'
```

### B. Kirim ke Nomor Default (Otomatis menggunakan `default_recipient` di `config.json`):
```bash
curl -X POST http://localhost:3030/api/v1/messages/send \
  -H "Authorization: Bearer your-secret-bearer-token" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Halo! Ini pesan notifikasi otomatis tanpa perlu set parameter `to`."
  }'
```

**Contoh Response Sukses:**
```json
{
  "success": true,
  "data": {
    "message_id": "3EB04F18B37A9F43862A",
    "to": "6281234567890@s.whatsapp.net",
    "status": "SENT",
    "timestamp": "2026-10-12T08:05:00Z"
  }
}
```

---

## 5. Kirim Media / Dokumen (Gambar / Foto / File)

### A. Menggunakan URL Gambar:
```bash
curl -X POST http://localhost:3030/api/v1/messages/send-media \
  -H "Authorization: Bearer your-secret-bearer-token" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "081234567890",
    "type": "image",
    "caption": "Grafik Monitoring Server Terkini",
    "media_url": "https://example.com/grafik-cctv.png"
  }'
```

### B. Menggunakan Data Base64:
```bash
curl -X POST http://localhost:3030/api/v1/messages/send-media \
  -H "Authorization: Bearer your-secret-bearer-token" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "081234567890",
    "type": "image",
    "caption": "Screenshot CCTV",
    "data_base64": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII="
  }'
```

### C. Mengirim File Dokumen / PDF:
```bash
curl -X POST http://localhost:3030/api/v1/messages/send-media \
  -H "Authorization: Bearer your-secret-bearer-token" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "081234567890",
    "type": "document",
    "file_name": "backup_db_2026-10-12.log",
    "mime_type": "text/plain",
    "caption": "Log backup database mingguan",
    "data_base64": "SGFsbywgaW5pIGlzZSBmaWxlIGxvZyBiYWNrdXAh"
  }'
```

---

## 6. Logout / Putuskan Tautan Perangkat (Unlink)
```bash
curl -X POST http://localhost:3030/api/v1/session/logout \
  -H "Authorization: Bearer your-secret-bearer-token"
```

---

## 7. Akses Dokumentasi Swagger / OpenAPI
- **OpenAPI JSON Spec**: `GET http://localhost:3030/openapi.json`
- **Swagger Interactive Docs**: Buka browser ke `http://localhost:3030/docs`
