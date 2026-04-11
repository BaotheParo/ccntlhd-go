package main

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"
)

// === ĐIỂM KIỂM TRA SỨC KHỎE (HEALTH CHECK ENDPOINT) ===
func startHealthServer() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok","service":"ticketing-api"}`)
	})
	go func() {
		if err := http.ListenAndServe(":8081", nil); err != nil {
			fmt.Printf("[Lỗi] Không thể khởi động máy chủ Health Check: %v\n", err)
		}
	}()
	fmt.Println("[Hệ thống] Điểm kiểm tra sức khỏe mở tại: http://localhost:8081/health")
}

// === CƠ CHẾ THỬ LẠI VỚI ĐỘ TRỄ TĂNG DẦN CẤP SỐ NHÂN (EXPONENTIAL BACKOFF) ===
func retryWithBackoff(fn func() error, maxRetries int, baseDelay time.Duration) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		lastErr = fn()
		if lastErr == nil {
			return nil // Thành công rực rỡ!
		}
		
		if i == maxRetries-1 {
			break
		}

		// Tính toán độ trễ: base * 2^i (ví dụ: 500ms, 1s, 2s, 4s...)
		delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(i)))
		fmt.Printf("[Thử lại] Lần %d THẤT BẠI: %v. Đang chờ %v để thử lại...\n", i+1, lastErr, delay)
		time.Sleep(delay)
	}
	return fmt.Errorf("hệ thống bỏ cuộc sau %d lần thử thất bại liên tiếp: %w", maxRetries, lastErr)
}

func main() {
	fmt.Println("--- DEMO RESILIENCY (HEALTH CHECK & EXPONENTIAL BACKOFF) ---")
	fmt.Println("")

	// Khởi động Health Check Server trong nền
	startHealthServer()

	// Giả lập kết nối CSDL: Cố tình thất bại 3 lần đầu, thành công ở lần thứ 4
	attempt := 0
	err := retryWithBackoff(func() error {
		attempt++
		if attempt < 4 {
			return errors.New("kết nối CSDL bị từ chối (Connection Refused)")
		}
		fmt.Printf("[Kết nối] >>> CSDL ĐÃ KẾT NỐI THÀNH CÔNG ở lần thử thứ %d!\n", attempt)
		return nil
	}, 5, 500*time.Millisecond) // Tối đa 5 lần thử, độ trễ cơ sở 500ms

	if err != nil {
		fmt.Println(">>> LỖI NGHIÊM TRỌNG:", err)
	} else {
		fmt.Println(">>> Hệ thống đã sẵn sàng phục vụ khách hàng!")
	}

	// Giữ chương trình chạy thêm 1 giây để quan sát
	time.Sleep(1 * time.Second)
	fmt.Println("\n--- KẾT THÚC DEMO ---")
}
