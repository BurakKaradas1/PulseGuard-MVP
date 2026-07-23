# PulseGuard 

PulseGuard is a lightweight and modular cybersecurity monitoring tool designed to securely, losslessly, and asynchronously stream system statuses of devices (agents) on a network to a central server (Collector). Built using Go and React.

---

## Key Features (Engineering Details)

*   **HMAC SHA-256 Data Integrity:** Traffic between the agent and the C2 server is protected with end-to-end cryptographic signatures, making it resilient against Man-in-the-Middle (MitM) and Replay Attack scenarios.
*   **Continuous WAL (Write-Ahead Logging) & Safe Truncation:** Even if an agent's network connection drops, data is written to a local WAL file on disk. When connectivity returns, data is forwarded to the C2. The newly implemented "Safe Truncation" architecture eliminates read/write Race Condition risks, ensuring 100% data security (Zero Data Loss).
*   **Idempotency Protection:** Even if an agent accidentally sends the same log batch twice, the C2 server recognizes packets via a "Batch ID" and prevents duplicate database entries.
*   **Dynamic Command-Line Tool (pulsectl):** A professional CLI tool built with the Cobra framework to manage the fleet via the terminal, supporting dynamic target server (`--server`) routing.
*   **React-Based Dashboard:** A modern interface for real-time fleet management, featuring `.env`-based dynamic routing and one-click **PDF & CSV report export** capabilities.

---

## Architectural Layers

1.  **PulseGuard Agent:** A Go-based agent deployed on client machines to collect system metrics (CPU, RAM, Disk) and buffer them on a WAL during offline periods.
2.  **PulseGuard Collector (C2):** A central REST API server that validates cryptographic data coming from agents and persists it using SQLite.
3.  **PulseGuard Dashboard:** A real-time fleet management panel built with React/Vite.
4.  **PulseGuard pulsectl:** A command-line interface designed for system administrators (SysAdmins).

---

## Installation & Setup Guide

Follow these steps sequentially to set up the project in your local environment. (Requirements: `Go 1.21+` and `Node.js 18+`)

### 1. Starting the Central Server (Collector)
The core backbone managing database connections and the REST API.
```bash
cd pulseguard-collector
go mod tidy
go run main.go

### 2. Starting the Interface (Dashboard)
- To run the React-based control panel:
cd pulseguard-dashboard
npm install

- Once installation is complete, if you do not have an .env file in your environment, rename .env.example to .env.
npm run dev

### 3. Starting the Agent
- To simulate an agent collecting device metrics and transmitting them securely to the C2:
cd pulseguard-agent
go mod tidy
go run main.go

### 4. Using the CLI Tool (pulsectl)
- To check system status from the command line:
cd pulseguard-pulsectl
go mod tidy

# To fetch fleet status connecting to the default localhost C2 server:
go run main.go status

# To connect to a C2 server at a different IP address:
go run main.go status --server [http://192.168.1.150:9000](http://192.168.1.150:9000)