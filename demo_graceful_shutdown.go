package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Công nhân xử lý tác vụ
func worker(id int, tasks <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		fmt.Printf("[Công nhân %d] Đang xử lý tác vụ %d...\n", id, task)
		time.Sleep(500 * time.Millisecond) // Giả lập công việc
	}
	fmt.Printf("[Công nhân %d] >>> Đã dừng lại một cách sạch sẽ.\n", id)
}

func main() {
	fmt.Println("--- DEMO GRACEFUL SHUTDOWN (DỪNG HỆ THỐNG AN TOÀN) ---")
	fmt.Println("Ghi chú: Hệ thống sẽ tự động gửi tín hiệu dừng sau 3 giây để demo.")

	taskQueue := make(chan int, 20)
	var wg sync.WaitGroup

	// 1. Khởi động 3 công nhân
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go worker(i, taskQueue, &wg)
	}

	// 2. Luồng sản xuất liên tục sinh tác vụ
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		id := 1
		for {
			select {
			case <-ctx.Done():
				fmt.Println("[Hệ thống] Đã dừng phân bổ tác vụ mới.")
				return
			case taskQueue <- id:
				fmt.Printf("[Hệ thống] Đã đẩy tác vụ %d vào hàng đợi\n", id)
				id++
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	// 3. Lắng nghe tín hiệu ngắt (Ctrl+C / SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// MÔ PHỎNG: Tự động gửi tín hiệu dừng sau 3.5 giây để demo
	go func() {
		time.Sleep(3500 * time.Millisecond)
		fmt.Println("\n[Demo] Tự động gửi tín hiệu SIGINT (Ctrl+C)...")
		proc, _ := os.FindProcess(os.Getpid())
		proc.Signal(os.Interrupt)
	}()

	<-quit // Chờ tín hiệu dừng

	fmt.Println("\n=== NHẬN TÍN HIỆU DỪNG - BẮT ĐẦU QUY TRÌNH TẮT AN TOÀN ===")
	
	cancel()         // 1. Dừng luồng sản xuất
	close(taskQueue) // 2. Khóa hàng đợi (không cho nhận thêm việc mới)
	
	fmt.Println("[Hệ thống] Đang chờ các công nhân xử lý nốt công việc tồn đọng...")
	wg.Wait()        // 3. Chờ tất cả công nhân hoàn thành nốt công việc đang làm dở
	
	fmt.Println("=== TẤT CẢ LUỒNG ĐÃ DỪNG. MÁY CHỦ TẮT AN TOÀN! ===")
}
