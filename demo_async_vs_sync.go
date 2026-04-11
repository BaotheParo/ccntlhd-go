package main

import (
	"fmt"
	"time"
)

// Tác vụ giả lập mất 2 giây để hoàn thành (ví dụ: gửi mail, nén ảnh)
func heavyTask() {
	time.Sleep(2 * time.Second)
	fmt.Println("[Tác vụ nền] Đã xử lý xong dữ liệu nặng.")
}

func main() {
	fmt.Println("--- SO SÁNH PHẢN HỒI: ĐỒNG BỘ VS BẤT ĐỒNG BỘ ---")
	fmt.Println("")

	fmt.Println("--- 1. CHẠY ĐỒNG BỘ (Synchronous) ---")
	startSync := time.Now()
	fmt.Println("[Luồng chính] Đang thực thi tác vụ nặng...")
	heavyTask() // Chặn luồng chính tại đây
	fmt.Printf(">>> Thời gian phản hồi đồng bộ: %v\n\n", time.Since(startSync))

	fmt.Println("--- 2. CHẠY BẤT ĐỒNG BỘ (Asynchronous) ---")
	startAsync := time.Now()
	fmt.Println("[Luồng chính] Đẩy tác vụ chạy ngầm, phản hồi ngay lập tức...")
	go heavyTask() // Đẩy tác vụ chạy ngầm, không chặn luồng chính
	fmt.Printf(">>> Thời gian phản hồi bất đồng bộ: %v\n", time.Since(startAsync))
	fmt.Println("[Luồng chính] Hệ thống vẫn đang phản hồi cho người dùng, tác vụ chạy ngầm đang tiếp tục...")
	
	// Chờ một chút để tác vụ ngầm in kết quả trước khi chương trình thoát
	time.Sleep(3 * time.Second) 
	fmt.Println("\n--- KẾT THÚC DEMO ---")
}
