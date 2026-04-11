package main

import "fmt"

// Hàm tạo dữ liệu (Source/Generator), đẩy vào Kênh và tự động đóng Kênh khi hoàn tất
func generate(nums ...int) <-chan int {
	out := make(chan int, len(nums))
	for _, n := range nums {
		out <- n
	}
	close(out)
	return out
}

// Hàm nhận dữ liệu (Processor), xử lý bất đồng bộ và gửi kết quả vào Kênh mới
func square(in <-chan int) <-chan int {
	out := make(chan int, 5)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()
	return out
}

func main() {
	fmt.Println("--- BẮT ĐẦU DEMO CHANNEL PIPELINE ---")
	fmt.Println("Mục tiêu: Xử lý chuỗi dữ liệu (1, 2, 3, 4, 5) qua các công đoạn...")

	// Khởi tạo luồng xử lý (Pipeline)
	// Công đoạn 1: Tạo dữ liệu
	c := generate(1, 2, 3, 4, 5)
	// Công đoạn 2: Tính bình phương (Chạy bất đồng bộ)
	out := square(c)

	// Công đoạn 3: Tiêu thụ dữ liệu (Consumer) và in kết quả
	fmt.Println("--- Kết quả chuỗi xử lý ---")
	for result := range out {
		fmt.Printf("[Pipeline] Kết quả bình phương: %d\n", result)
	}
	fmt.Println("--- KẾT THÚC DEMO ---")
}
