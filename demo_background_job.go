package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Hàm khởi tạo tác vụ chạy ngầm định kỳ (Background Job)
func startPeriodicJob(ctx context.Context, interval time.Duration, job func()) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop() // Luôn giải phóng tài nguyên bộ đếm nhịp

		log.Printf("BẮT ĐẦU: Tác vụ chạy ngầm khởi động, chu kỳ: %v\n", interval)
		for {
			select {
			case t := <-ticker.C:
				log.Printf("NHỊP ĐẾM: Phát hiện tín hiệu lúc %v - Đang thực hiện tác vụ...\n", t.Format("15:04:05"))
				job()

			case <-ctx.Done():
				log.Println("DỪNG: Nhận tín hiệu hủy bỏ (Context Cancel). Tiến trình kết thúc an toàn.")
				return
			}
		}
	}()
}

func main() {
	// Khởi tạo Ngữ cảnh (Context) với khả năng hủy bỏ
	ctx, cancel := context.WithCancel(context.Background())
	jobCount := 0

	fmt.Println("--- DEMO BACKGROUND JOB & CLEAN SHUTDOWN ---")
	
	// Đăng ký một tác vụ chạy ngầm mỗi 5 giây
	startPeriodicJob(ctx, 5*time.Second, func() {
		jobCount++
		fmt.Printf("    -> [Tác vụ #%d] PHÂN TÍCH DỮ LIỆU TICKET THÀNH CÔNG\n", jobCount)
	})

	fmt.Println("Hệ thống giả lập đang hoạt động. Sẽ tự động tắt sau 22 giây...")
	time.Sleep(22 * time.Second)

	fmt.Println("\n[Hệ thống] Gửi tín hiệu CANCEL tới tất cả tác vụ chạy ngầm...")
	cancel() 

	// Chờ một giây để luồng ngầm kịp in nhật ký kết thúc
	time.Sleep(1 * time.Second) 
	fmt.Println("Hệ thống đã dừng hoàn toàn. Trạng thái: Sạch sẽ (Clean).")
}
