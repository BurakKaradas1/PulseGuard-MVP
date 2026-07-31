# PulseGuard - Docker Compose Kullanımı

Bu kurulum `pulseguard-collector` (backend) ve `pulseguard-dashboard`
(frontend) servislerini container'da ayağa kaldırır. **Agent dahil
değildir** - agent, Windows host'unda native olarak çalıştırılmalıdır.

## Neden agent Docker'da değil?

Docker Desktop, Windows'ta container'ları WSL2 içindeki bir Linux sanal
makinesinde çalıştırır. Container içinden okunan CPU/RAM/Disk
metrikleri, gerçek Windows host'un değil, o VM'in kaynaklarıdır - bu
Docker'ın izolasyon modelinin bir sonucu, bir bug değil.

Datadog Agent, New Relic Infrastructure, Netdata, Wazuh gibi tüm
endpoint-monitoring ürünleri de aynı sebepten native installer
(.msi/.exe, systemd/launchd servisi) olarak dağıtılır. Bu proje de
aynı yaklaşımı izliyor: agent host'a native kurulur, backend/frontend
container'da çalışır.

## Kurulum

1. `.env.example` dosyasını `.env` olarak kopyala ve `PULSEGUARD_SECRET`
   değerini agent'ta (`pulseguard-agent/.env`) kullandığın değerle
   **aynı** yap:

   ```powershell
   Copy-Item .env.example .env
   notepad .env
   ```

2. Collector ve dashboard'u ayağa kaldır:

   ```powershell
   docker compose up -d --build
   ```

3. Agent'ı Windows'ta native çalıştır (Docker dışında):

   ```powershell
   cd pulseguard-agent
   go run main.go
   ```

   `config.yaml` içindeki `collector_url` değeri zaten
   `http://localhost:8080` olduğu için, collector container'ının
   published port'una (8080) sorunsuz ulaşır.

4. Dashboard'a tarayıcıdan eriş:

   ```
   http://localhost:5173
   ```

## Notlar

- **VITE_API_BASE_URL sadece build sırasında** JS bundle'ına gömülür
  (Vite'ın çalışma şekli budur). Bu değeri değiştirirsen
  `docker compose up -d --build` ile yeniden build etmen gerekir.
- Collector'ın SQLite veritabanı `collector-data` adlı bir Docker
  volume'ünde saklanır, `docker compose down` ile silinmez. Tamamen
  sıfırlamak istersen: `docker compose down -v`.
- Agent'ı arka planda ve terminal açmadan çalıştırmak istersen, NSSM
  veya Windows Task Scheduler ile bir servis haline getirebilirsin -
  istersen bu adımda da rehberlik edebilirim.
