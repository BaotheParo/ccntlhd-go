package main

import (
	"fmt"
	"time"
)

func main() {
	// Khởi tạo một Kênh đệm (Buffered Channel) với sức chứa là 2
	messageChannel := make(chan string, 2)

	// Khởi chạy Goroutine xử lý độc lập
	fmt.Println("--- BẮT ĐẦU DEMO GOROUTINE & CHANNELS ---")
	go func() {
		fmt.Println("[Goroutine] Đang xử lý dữ liệu nặng trong nền...")
		time.Sleep(2 * time.Second) // Giả lập tác vụ tốn thời gian

		// Gửi dữ liệu vào Kênh
		messageChannel <- "KẾT QUẢ 01: Đã xử lý xong"
		messageChannel <- "KẾT QUẢ 02: Đã xử lý xong"
		fmt.Println("[Goroutine] Đã gửi toàn bộ dữ liệu vào Kênh.")
	}()

	fmt.Println("[Luồng chính] Đang bị chặn (Block) để chờ kết quả từ Goroutine...")

	// Nhận dữ liệu từ Kênh (Luồng chính sẽ đợi cho đến khi có dữ liệu)
	msg1 := <-messageChannel
	msg2 := <-messageChannel

	fmt.Printf("[Luồng chính] NHẬN ĐƯỢC: %s\n", msg1)
	fmt.Printf("[Luồng chính] NHẬN ĐƯỢC: %s\n", msg2)
	fmt.Println("--- KẾT THÚC DEMO ---")
}
