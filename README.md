# WhatsApp Notification Service

Layanan Gateway Notifikasi WhatsApp mandiri yang **sangat ringan**, **hemat RAM & CPU** (< 35 MB RAM saat idle), dan **mudah dirawat** berkat penerapan **Clean Architecture**.

Dibangun menggunakan **Golang** dan library native protocol **[`whatsmeow`](https://github.com/tulir/whatsmeow)** tanpa browser headless (tanpa Chromium/Puppeteer), sehingga sangat ideal untuk server rumahan, VPS mini, maupun **Raspberry Pi**.

---

## 🚀 Fitur Utama

- **Ultra-Lightweight**: Konsumsi RAM hanya ~20–35 MB, CPU idle mendekati 0%.
- **Clean Architecture**: Pemisahan jelas antara Domain, Use Cases, Repositories, dan Delivery/Adapters.
- **Web UI Pairing Ramah Pengguna**: Cukup buka `http://<ip-server>:3030/login` dari browser HP/Laptop untuk langsung scan QR code dan melihat status perangkat.
- **Terminal ASCII QR Code**: QR code juga otomatis dicetak di terminal/log container.
- **Persistent Session**: Menggunakan SQLite lokal (`wa_session.db`). Saat server restart atau reboot, koneksi otomatis tersambung kembali tanpa perlu scan ulang.
- **Bearer Token Security**: Endpoint private diproteksi dengan header `Authorization: Bearer <token>`.
- **Default Recipient Support**: Kirim pesan cepat tanpa menyertakan nomor penerima (otomatis ke `default_recipient` di `config.json`).
- **Send Media / Document**: Mendukung pengiriman file gambar, screenshot, atau dokumen (PDF/log) via URL atau Base64.
- **Safe Rate Limiter**: Membatasi jeda pengiriman antar pesan untuk mencegah nomor ditandai spam/ban.
- **OpenAPI & Swagger Ready**: Dilengkapi endpoint `/openapi.json` dan interactive docs di `/docs`.
- **Deploy Cepat via Docker Compose**.

---

## 📁 Struktur Direktori (Clean Architecture)

```text
wa-notif/
├── cmd/
│   └── server/
│       └── main.go                  # Dependency Injection & Application Entrypoint
├── config/
│   └── config.go                    # Config JSON Loader (env: CONFIG_PATH)
├── deploy/
│   └── docker-compose.yml           # Deploy di myrpi: /home/myrpi/docker/wa-notif/
├── internal/
│   ├── domain/                      # Domain Layer (Core Business Rules)
│   │   ├── entity/
│   │   │   ├── message.go           # Message entity
│   │   │   └── session.go           # Session state & device info
│   │   └── repository/
│   │       └── wa_gateway.go        # Port Interface WhatsApp Gateway
│   ├── usecase/                     # Application Layer (Use Cases)
│   │   ├── send_message_uc.go       # Logic kirim pesan teks & media
│   │   ├── session_uc.go            # Logic status, QR code, & logout
│   │   └── dto/                     # Request & Response DTOs
│   └── adapter/                     # Interface Adapters Layer
│       ├── http/                    # Inbound HTTP Delivery
│       │   ├── router.go            # Route registration & CORS
│       │   ├── middleware/auth.go   # Bearer token validator
│       │   ├── handler/             # REST Handlers (Message, Session, OpenAPI)
│       │   └── web/pairing_ui.go    # Embedded Pairing Web UI
│       └── whatsapp/                # Outbound WhatsApp Client
│           └── whatsmeow_adapter.go # Whatsmeow & SQLite store implementation
├── pkg/                             # Shared Utility Packages
│   ├── phone/                       # Smart Phone Number Normalizer
│   └── logger/                      # Slog structured logger
├── docs/
│   ├── openapi.json                 # OpenAPI 3.0 Specification
│   └── CURL_EXAMPLES.md            # Panduan cURL lengkap
├── storage/                         # Persistent database files (gitignored)
├── config.example.json              # Contoh template konfigurasi
├── Dockerfile                       # Multi-stage build (linux/arm64)
└── README.md
```

---

## ⚙️ Konfigurasi (`config.json`)

Konfigurasi disimpan dalam format JSON:

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 3030
  },
  "auth": {
    "bearer_token": "your-secret-bearer-token"
  },
  "whatsapp": {
    "default_recipient": "6281234567890",
    "session_db_path": "./storage/wa_session.db",
    "device_name": "WA Notification Gateway",
    "rate_limit_seconds": 1.5
  },
  "log": {
    "level": "info"
  }
}
```

---

## 🐳 Cara Deploy dengan Docker Compose

### Deploy di Raspberry Pi (myrpi)

File `deploy/docker-compose.yml` didesain untuk di-copy ke Pi dan langsung dijalankan.

#### 1. Build image (dari mesin dev, cross-compile ke arm64)
```bash
docker buildx build --platform linux/arm64 -t wa-notif:<tag> --load .
# Simpan dan transfer ke Pi:
docker save wa-notif:<tag> | ssh myrpi@192.168.1.222 "docker load"
```

#### 2. Siapkan folder di Pi
```bash
ssh myrpi@192.168.1.222
sudo mkdir -p /mnt/ssd/wa-notif/storage
sudo chown myrpi:65532 /mnt/ssd/wa-notif/storage
mkdir -p ~/docker/wa-notif
```

#### 3. Salin file deploy ke Pi
```bash
# Dari mesin dev:
scp deploy/docker-compose.yml myrpi@192.168.1.222:~/docker/wa-notif/
scp config.example.json myrpi@192.168.1.222:~/docker/wa-notif/config.json
```

#### 4. Edit konfigurasi di Pi
```bash
ssh myrpi@192.168.1.222
nano ~/docker/wa-notif/config.json   # isi bearer_token & default_recipient
chown myrpi:65532 ~/docker/wa-notif/config.json && chmod 640 ~/docker/wa-notif/config.json
```

#### 5. Jalankan
```bash
cd ~/docker/wa-notif
docker compose up -d
```

#### 6. Pairing Perangkat WhatsApp
Buka browser dan akses Web UI:
```text
http://<IP_RASPBERRY_PI>:3030/login
```
Scan QR code yang muncul menggunakan WhatsApp di HP (**Perangkat Tertaut** > **Tautkan Perangkat**).

---

### Deploy Lokal (Dev / Test)

#### 1. Sesuaikan Konfigurasi
Salin dan edit `config.example.json` → `config.json`:
```bash
cp config.example.json config.json
nano config.json
```

#### 2. Jalankan dengan docker-compose lama (root)
```bash
# Untuk dev lokal, kamu juga bisa pakai docker-compose.yml yang sudah ada di root
docker compose up -d --build
```

#### 3. Cek Logs
```bash
docker compose logs -f
```

---

## 📡 Daftar Endpoint API

Semua endpoint private memerlukan header `Authorization: Bearer <bearer_token>`.

| Method | Endpoint | Deskripsi | Auth |
| :--- | :--- | :--- | :--- |
| `GET` | `/login` atau `/` | Web UI untuk scan QR Code & cek status | Public |
| `GET` | `/health` | Healthcheck service | Public |
| `GET` | `/docs` | Swagger / Interactive API Documentation | Public |
| `GET` | `/openapi.json` | OpenAPI 3.0 JSON Specification | Public |
| `GET` | `/api/v1/session/status` | Cek status koneksi WA & device info | Bearer Token |
| `GET` | `/api/v1/session/qr` | Ambil QR Code data & Base64 PNG | Bearer Token |
| `POST` | `/api/v1/session/logout` | Logout / Putuskan tautan akun WhatsApp | Bearer Token |
| `POST` | `/api/v1/messages/send` | Kirim pesan teks | Bearer Token |
| `POST` | `/api/v1/messages/send-media` | Kirim file media/dokumen via Base64/URL | Bearer Token |

---

## 💻 Contoh Penggunaan cURL

### 1. Kirim Pesan Teks
```bash
curl -X POST http://localhost:3030/api/v1/messages/send \
  -H "Authorization: Bearer your-secret-bearer-token" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "081234567890",
    "message": "*[SERVER ALERT]* CPU load 92% on Raspberry Pi!"
  }'
```

### 2. Kirim ke Nomor Default (Cukup kirim `message`)
```bash
curl -X POST http://localhost:3030/api/v1/messages/send \
  -H "Authorization: Bearer your-secret-bearer-token" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Backup database mingguan selesai dengan sukses."
  }'
```

Untuk contoh lengkap pengiriman gambar, base64, file dokumen, dan integrasi lainnya, lihat **[docs/CURL_EXAMPLES.md](docs/CURL_EXAMPLES.md)**.
