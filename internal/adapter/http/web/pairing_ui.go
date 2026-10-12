package web

import (
	"net/http"
)

const PairingHTML = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>WhatsApp Notification Gateway - Pairing & Status</title>
  <style>
    :root {
      --bg: #0f172a;
      --card-bg: #1e293b;
      --text-main: #f8fafc;
      --text-muted: #94a3b8;
      --primary: #22c55e;
      --primary-hover: #16a34a;
      --border: #334155;
      --badge-green: #065f46;
      --badge-yellow: #854d0e;
      --badge-red: #991b1b;
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      background: var(--bg);
      color: var(--text-main);
      display: flex;
      justify-content: center;
      align-items: center;
      min-height: 100vh;
      padding: 1.5rem;
    }
    .container {
      background: var(--card-bg);
      border: 1px solid var(--border);
      border-radius: 1rem;
      max-width: 520px;
      width: 100%;
      padding: 2rem;
      box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.4);
    }
    .header { text-align: center; margin-bottom: 1.5rem; }
    .header h1 { font-size: 1.5rem; font-weight: 700; color: #fff; margin-bottom: 0.25rem; }
    .header p { color: var(--text-muted); font-size: 0.875rem; }
    .status-badge {
      display: inline-flex;
      align-items: center;
      gap: 0.5rem;
      padding: 0.35rem 0.85rem;
      border-radius: 9999px;
      font-size: 0.8rem;
      font-weight: 600;
      margin-top: 0.75rem;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .badge-dot { width: 8px; height: 8px; border-radius: 50%; }
    .status-connected { background: var(--badge-green); color: #a7f3d0; }
    .status-connected .badge-dot { background: #34d399; }
    .status-pairing { background: var(--badge-yellow); color: #fef08a; }
    .status-pairing .badge-dot { background: #facc15; }
    .status-disconnected { background: var(--badge-red); color: #fecaca; }
    .status-disconnected .badge-dot { background: #f87171; }
    .qr-box {
      background: #ffffff;
      border-radius: 0.75rem;
      padding: 1rem;
      display: flex;
      justify-content: center;
      align-items: center;
      margin: 1.5rem auto;
      width: 260px;
      height: 260px;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.2);
    }
    .qr-box img { width: 100%; height: 100%; display: block; border-radius: 0.5rem; }
    .spinner {
      border: 3px solid rgba(0,0,0,0.1);
      border-top-color: #22c55e;
      border-radius: 50%;
      width: 36px;
      height: 36px;
      animation: spin 1s linear infinite;
    }
    @keyframes spin { to { transform: rotate(360deg); } }
    .info-card {
      background: #0f172a;
      border: 1px solid var(--border);
      border-radius: 0.5rem;
      padding: 1rem;
      font-size: 0.875rem;
      margin-bottom: 1.25rem;
    }
    .info-row { display: flex; justify-content: space-between; margin-bottom: 0.5rem; }
    .info-row:last-child { margin-bottom: 0; }
    .info-label { color: var(--text-muted); }
    .info-val { font-weight: 600; color: #fff; font-family: monospace; }
    .btn {
      width: 100%;
      background: var(--primary);
      color: #fff;
      border: none;
      border-radius: 0.5rem;
      padding: 0.75rem 1rem;
      font-size: 0.95rem;
      font-weight: 600;
      cursor: pointer;
      transition: background 0.2s;
    }
    .btn:hover { background: var(--primary-hover); }
    .btn-danger { background: #dc2626; margin-top: 0.5rem; }
    .btn-danger:hover { background: #b91c1c; }
    .token-input {
      width: 100%;
      background: #0f172a;
      border: 1px solid var(--border);
      color: #fff;
      padding: 0.5rem 0.75rem;
      border-radius: 0.375rem;
      margin-bottom: 0.75rem;
      font-size: 0.875rem;
    }
    .footer { text-align: center; margin-top: 1.5rem; font-size: 0.75rem; color: var(--text-muted); }
    .footer a { color: #38bdf8; text-decoration: none; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>WhatsApp Gateway</h1>
      <p>Clean Architecture • Ultra-Lightweight • Port 3030</p>
      <div id="statusBadge" class="status-badge status-disconnected">
        <span class="badge-dot"></span>
        <span id="statusText">MEMERIKSA KONEKSI...</span>
      </div>
    </div>

    <!-- QR Container -->
    <div id="qrSection">
      <div class="qr-box" id="qrContainer">
        <div style="text-align:center; color:#64748b; font-size:0.85rem;">
          <div class="spinner" style="margin:0 auto 12px;"></div>
          <span>Menghubungkan ke WhatsApp...</span>
        </div>
      </div>
      <p style="text-align: center; font-size: 0.85rem; color: var(--text-muted); line-height: 1.4;">
        Buka WhatsApp di HP &gt; <b>Perangkat Tertaut</b> &gt; <b>Tautkan Perangkat</b>, lalu arahkan kamera ke QR Code di atas.
      </p>
    </div>

    <!-- Connected Info Container -->
    <div id="infoSection" style="display: none;">
      <div class="info-card">
        <div class="info-row">
          <span class="info-label">Status:</span>
          <span class="info-val" style="color: #34d399;">TERHUBUNG (CONNECTED)</span>
        </div>
        <div class="info-row">
          <span class="info-label">Akun WhatsApp:</span>
          <span class="info-val" id="infoJid">-</span>
        </div>
        <div class="info-row">
          <span class="info-label">Nama Perangkat:</span>
          <span class="info-val" id="infoName">-</span>
        </div>
      </div>

      <div class="info-card" style="margin-top: 1rem;">
        <p style="font-weight: 600; margin-bottom: 0.5rem; font-size: 0.85rem;">Test Kirim Notifikasi:</p>
        <div style="margin-bottom: 0.5rem;">
          <input type="password" id="tokenInput" class="token-input" placeholder="Masukkan Bearer Token Anda">
        </div>
        <input type="text" id="testTo" class="token-input" placeholder="Nomor HP (e.g. 08123456789) - Kosongkan jika pakai default">
        <input type="text" id="testMsg" class="token-input" placeholder="Pesan..." value="Halo! Ini pesan tes dari WA Gateway RPi.">
        <button class="btn" onclick="sendTestMessage()">Kirim Pesan Uji Coba</button>
        <div id="testResult" style="margin-top: 0.5rem; font-size: 0.8rem;"></div>
      </div>

      <button class="btn btn-danger" onclick="logoutDevice()">Putuskan Perangkat (Logout)</button>
    </div>

    <div class="footer">
      <a href="/docs" target="_blank">Lihat Dokumentasi OpenAPI / Swagger &rarr;</a>
    </div>
  </div>

  <script>
    const tokenInput = document.getElementById('tokenInput');
    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.has('token')) {
      tokenInput.value = urlParams.get('token');
      localStorage.setItem('wa_token', urlParams.get('token'));
    } else if (localStorage.getItem('wa_token')) {
      tokenInput.value = localStorage.getItem('wa_token');
    }

    tokenInput.addEventListener('change', () => {
      localStorage.setItem('wa_token', tokenInput.value);
    });

    function getHeaders() {
      const h = { 'Content-Type': 'application/json' };
      const token = tokenInput ? tokenInput.value.trim() : '';
      if (token) {
        h['Authorization'] = 'Bearer ' + token;
      }
      return h;
    }

    async function checkStatus() {
      try {
        const res = await fetch('/api/v1/session/status');
        const json = await res.json();
        if (json.success && json.data) {
          const d = json.data;
          updateUI(d.connected, d.status, d.jid, d.push_name);
        }
      } catch (e) {
        console.error('Status fetch error:', e);
      }
    }

    async function fetchQR() {
      try {
        const res = await fetch('/api/v1/session/qr');
        const json = await res.json();
        if (json.success && json.data) {
          if (json.data.connected) {
            checkStatus();
            return;
          }
          if (json.data.qr_code_data) {
            document.getElementById('qrContainer').innerHTML = '<img src="' + json.data.qr_code_data + '" alt="Scan WhatsApp QR">';
          }
        }
      } catch (e) {
        console.error('QR fetch error:', e);
      }
    }

    function updateUI(connected, status, jid, name) {
      const badge = document.getElementById('statusBadge');
      const text = document.getElementById('statusText');
      const qrSec = document.getElementById('qrSection');
      const infoSec = document.getElementById('infoSection');

      badge.className = 'status-badge';

      if (connected) {
        badge.classList.add('status-connected');
        text.innerText = 'TERHUBUNG';
        qrSec.style.display = 'none';
        infoSec.style.display = 'block';
        document.getElementById('infoJid').innerText = jid || '-';
        document.getElementById('infoName').innerText = name || '-';
      } else {
        if (status === 'PAIRING') {
          badge.classList.add('status-pairing');
          text.innerText = 'PAIRING (SCAN QR)';
        } else {
          badge.classList.add('status-disconnected');
          text.innerText = 'TERPUTUS';
        }
        qrSec.style.display = 'block';
        infoSec.style.display = 'none';
        fetchQR();
      }
    }

    async function sendTestMessage() {
      const to = document.getElementById('testTo').value.trim();
      const msg = document.getElementById('testMsg').value.trim();
      const resDiv = document.getElementById('testResult');
      resDiv.innerHTML = '<span style="color: #facc15;">Mengirim pesan...</span>';

      try {
        const res = await fetch('/api/v1/messages/send', {
          method: 'POST',
          headers: getHeaders(),
          body: JSON.stringify({ to: to, message: msg })
        });
        const data = await res.json();
        if (data.success) {
          resDiv.innerHTML = '<span style="color: #34d399;">✓ Pesan terkirim! ID: ' + data.data.message_id + '</span>';
        } else {
          resDiv.innerHTML = '<span style="color: #f87171;">✗ Gagal: ' + (data.error ? data.error.message : 'Error') + '</span>';
        }
      } catch (e) {
        resDiv.innerHTML = '<span style="color: #f87171;">✗ Error: ' + e.message + '</span>';
      }
    }

    async function logoutDevice() {
      if (!confirm('Apakah Anda yakin ingin memutuskan tautan perangkat ini?')) return;
      try {
        const res = await fetch('/api/v1/session/logout', {
          method: 'POST',
          headers: getHeaders()
        });
        await res.json();
        checkStatus();
      } catch (e) {
        alert('Gagal logout: ' + e.message);
      }
    }

    // Auto-refresh status loop
    checkStatus();
    setInterval(() => {
      checkStatus();
    }, 3000);
  </script>
</body>
</html>`

func ServePairingUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(PairingHTML))
}
