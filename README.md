# 🛡️ PulseGuard

PulseGuard is a lightweight, secure, and real-time Cloud Workload Protection and Endpoint Monitoring platform. It consists of a Go-based agent-collector architecture and a modern React dashboard for fleet management.

## ✨ Features

* **Real-Time Telemetry:** Monitors CPU, RAM, and Disk usage across the entire fleet with low latency.
* **Secure Communication:** All agent-to-collector traffic is verified using **HMAC-SHA256** signatures to prevent unauthorized data spoofing.
* **Offline Resilience:** Agents utilize a custom Write-Ahead Log (.wal) queueing system. If the C2 (Command & Control) server is unreachable, data is safely stored locally and pushed dynamically upon reconnection.
* **Dynamic Thresholds & Alarms:** Administrators can set global or host-specific limits (e.g., Max CPU 90%). The system provides instant visual alerts and tracks critical events.
* **Slack Integration:** 🚀 Automatically sends real-time critical alerts to your designated Slack channel when host thresholds are exceeded.
* **Modern Dashboard:** Built with React, TypeScript, and Recharts, offering live area charts and seamless fleet status tracking.

## 🏗️ Architecture

* **Backend / Collector:** Go, SQLite (WAL mode enabled for high concurrency)
* **Agent:** Go, gopsutil (Hardware metrics), HMAC Security
* **Frontend:** React, TypeScript, Vite, Recharts
* **Deployment:** Docker & Docker Compose

---

## 🚀 Getting Started

> **Docker ile Hızlı Kurulum:** Collector ve Dashboard'u tek komutla ayağa kaldırmak için ana dizindeki `docker-compose.yml` dosyasını kullanıyoruz. **Agent**, donanım metriklerini doğru okuyabilmesi (Docker sanallaştırma katmanına takılmaması) için Windows host üzerinde native (doğrudan) çalıştırılmalıdır.

### 1. Environment Setup (.env)
Projenin ana dizininde (root) bir `.env` dosyası oluşturun. Docker, Collector ve Dashboard'u ayağa kaldırırken bu dosyayı otomatik olarak okuyacaktır.

```env
# /PulseGuard-MVP/.env
PULSEGUARD_SECRET=your-super-secret-key
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
VITE_API_BASE_URL=http://localhost:8080

### 2. Collector & Dashboard Setup (Docker)
# Servisleri inşa edip arka planda başlatır
docker-compose up -d --build

### 3. Agent Setup (Native)
cd pulseguard-agent
go run .