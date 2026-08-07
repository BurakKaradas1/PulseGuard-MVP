# 🛡️ PulseGuard

PulseGuard is a lightweight, secure, and real-time Cloud Workload Protection and Endpoint Monitoring platform. It consists of a Go-based agent-collector architecture and a modern React dashboard for fleet management.

## ✨ Features

* **Real-Time Telemetry:** Monitors CPU, RAM, and Disk usage across the entire fleet with low latency.
* **Secure Communication:** All agent-to-collector traffic is verified using **HMAC-SHA256** signatures to prevent unauthorized data spoofing.
* **Offline Resilience:** Agents utilize a custom Write-Ahead Log (.wal) queueing system. If the C2 (Command & Control) server is unreachable, data is safely stored locally and pushed dynamically upon reconnection.
* **Dynamic Thresholds & Alarms:** Administrators can set global or host-specific limits (e.g., Max CPU 90%). The system provides instant visual alerts and tracks critical events.
* **Modern Dashboard:** Built with React, TypeScript, and Recharts, offering live area charts and seamless fleet status tracking.

## 🏗️ Architecture

* **Backend / Collector:** Go, SQLite (WAL mode enabled for high concurrency)
* **Agent:** Go, gopsutil (Hardware metrics), HMAC Security
* **Frontend:** React, TypeScript, Vite, Recharts

## 🚀 Getting Started

### 1. Collector (Server) Setup
Navigate to the collector directory and set up your secure environment.
```bash
cd pulseguard-collector
# Create a .env file and add your secret key
# PULSEGUARD_SECRET=your-super-secret-key
go run main.go

### 2. Agent (Setup) Setup
cd pulseguard-agent
# Create a .env file and add your secret key
# PULSEGUARD_SECRET=your-super-secret-key
go run main.go

> **Not:** `.env` dosyaları artık uygulama başlarken otomatik olarak okunuyor
> (bağımlılıksız minik bir loader eklendi). Agent'ı ve collector'ı **aynı**
> `PULSEGUARD_SECRET` değeriyle çalıştırdığından emin ol, aksi halde agent
> event göndermeyi sessizce iptal eder ve dashboard'da metrikler hep %0
> görünür.

### 3. Dashboard (Frontend) Setup
cd pulseguard-dashboard
# Create a .env file and add your API base URL
# VITE_API_BASE_URL="http://localhost:8080"
npm install
npm run dev