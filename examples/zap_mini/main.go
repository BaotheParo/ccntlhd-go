package main

import (
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// logRequest giả lập phần mềm trung gian (middleware) ghi nhật ký cho các yêu cầu HTTP
func logRequest(log *zap.Logger, method, path string, status int, latency time.Duration) {
	fields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status", status),
		zap.Duration("latency", latency),
	}

	// Phân loại cấp độ nhật ký dựa trên mã trạng thái HTTP
	switch {
	case status >= 500:
		log.Error("Yêu cầu HTTP thất bại (Lỗi máy chủ)", fields...)
	case status >= 400:
		log.Warn("Yêu cầu HTTP cảnh báo (Lỗi máy khách)", fields...)
	default:
		log.Info("Yêu cầu HTTP thành công", fields...)
	}
}

func main() {
	// Khởi tạo Zap Logger ở chế độ Môi trường thực tế (Production - định dạng JSON)
	logger, _ := zap.NewProduction()
	defer logger.Sync() // Đảm bảo giải phóng bộ đệm trước khi thoát

	// Giả lập một số yêu cầu mạng
	requests := []struct {
		method string
		path   string
		status int
	}{
		{"GET", "/api/v1/events", 200},
		{"POST", "/api/v1/orders", 202},
		{"GET", "/api/v1/orders/abc", 404},
		{"POST", "/api/v1/auth/login", 401},
		{"POST", "/api/v1/orders", 500},
	}

	for _, r := range requests {
		lat := time.Duration(rand.Intn(50)+1) * time.Millisecond
		logRequest(logger, r.method, r.path, r.status, lat)
	}
}
