# Ticketing System (Flash Sale)

Hệ thống Backend xử lý **Ticketing/Flash Sale** chịu tải cao, được xây dựng bằng **Go (Golang)** tuân thủ kiến trúc **Clean Architecture**.

## 🚀 Tech Stack

-   **Language:** Go 1.22+
-   **Framework:** Fiber v2
-   **Config:** Viper
-   **Logger:** Zap
-   **Infra:** PostgreSQL 16, Redis 7, Docker

## �️ Prerequisites

Bạn cần cài đặt công cụ tùy theo mục đích:

| Role | Yêu cầu |
| :--- | :--- |
| **Runner** (Chỉ chạy app) | [Docker Desktop](https://www.docker.com/) |
| **Developer** (Code & Debug) | Docker + [Go 1.22+](https://go.dev/) |

## 🏃 Quick Start (Khuyên dùng)

Dành cho cả Developer và Runner. Chỉ cần 1 lệnh để dựng toàn bộ môi trường (App + DB + Redis).

```bash
make up
# Hoặc: docker-compose up -d --build
```

Sau khi chạy xong:
-   **Health Check**: [http://localhost:8080/health](http://localhost:8080/health)
-   **Logs**: `make logs`
-   **Stop**: `make down`

## 👨‍� Development Workflow

### 1. Project Structure
```text
├── cmd/server/main.go       # Entry point
├── config/                  # Config chuẩn (YAML)
├── internal/                # Logic code (Clean Arch)
├── pkg/                     # Libraries (Logger, Config)
├── Dockerfile               # Multi-stage build
└── docker-compose.yml       # Dev Environment
```

### 2. Dependency Management
Dự án sử dụng Go Modules.
-   Khi chạy bằng Docker, quá trình build sẽ **tự động** chạy `go mod tidy` bên trong container để tải thư viện (kể cả khi bạn chưa tải về máy host).
-   Nếu code local, hãy chạy: `go mod tidy`.

### 3. Cấu hình (Configuration)
File gốc: `config/config.yaml`.
Khi chạy Docker, cấu hình được override bằng biến môi trường (Environment Variables) trong `docker-compose.yml`:
-   `SERVER_PORT` -> `server.port`
-   `DATABASE_HOST` -> `database.host`
-   `REDIS_ADDR` -> `redis.addr`

## ❓ Troubleshooting

### Lỗi: `bind: address already in use`
*   **Nguyên nhân**: Port 8080, 5432 hoặc 6379 đang bị chiếm dụng bởi ứng dụng khác.
*   **Khắc phục**: Tắt ứng dụng đó hoặc đổi port mapping trong `docker-compose.yml` (Ví dụ: `"8081:8080"`).

### Lỗi: `dial tcp: connect: connection refused` (DB/Redis)
*   **Nguyên nhân**: App khởi động nhanh hơn Database.
*   **Khắc phục**: Container App sẽ tự restart (Do policy `restart: always`). Hãy chờ vài giây và kiểm tra lại logs bằng lệnh `make logs`.

### Lỗi chạy `go run` local không được
*   **Nguyên nhân**: Chưa cài Go hoặc chưa có DB local.
*   **Khắc phục**: Hãy dùng Docker (`make up`) để đảm bảo môi trường đồng nhất và không cần cài đặt phức tạp.

## 🔌 API Endpoints

| Method | Path | Mô tả |
| :--- | :--- | :--- |
| `GET` | `/health` | Kiểm tra hệ thống sống hay chết |
