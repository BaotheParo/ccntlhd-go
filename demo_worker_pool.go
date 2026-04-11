package main

import (
	"fmt"
	"sync"
	"time"
)

type Task struct {
	ID      int
	Payload string
}

// Hàm Worker đại diện cho một công nhân xử lý tác vụ
func worker(id int, tasks <-chan Task, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		fmt.Printf("[Công nhân %d] Bắt đầu xử lý Tác vụ #%d\n", id, task.ID)
		time.Sleep(300 * time.Millisecond) // Giả lập thời gian xử lý (IO/DB)
		fmt.Printf("[Công nhân %d] Hoàn thành Tác vụ #%d\n", id, task.ID)
	}
}

func main() {
	const numWorkers = 3 // Số lượng công nhân chạy đồng thời
	const numTasks = 10  // Tổng số lượng tác vụ cần xử lý

	fmt.Println("--- BẮT ĐẦU DEMO WORKER POOL ---")
	fmt.Printf("Cấu hình: %d Công nhân vs %d Tác vụ\n\n", numWorkers, numTasks)

	taskQueue := make(chan Task, numTasks)
	var wg sync.WaitGroup

	startTime := time.Now()

	// 1. Khởi động Nhóm công nhân (Worker Pool)
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, taskQueue, &wg)
	}

	// 2. Phân bổ công việc vào Kênh (Task Queue)
	for i := 1; i <= numTasks; i++ {
		taskQueue <- Task{ID: i, Payload: fmt.Sprintf("Dữ-liệu-vé-%d", i)}
	}
	close(taskQueue) // Đóng kênh để báo hiệu không còn tác vụ mới

	// 3. Chờ tất cả công nhân hoàn thành
	wg.Wait()

	duration := time.Since(startTime)
	fmt.Printf("\n=== KẾT QUẢ: Hoàn thành %d tác vụ trong thời gian: %v ===\n", numTasks, duration)
	fmt.Println("(Thời gian dự kiến nếu chạy tuần tự: ~3 giây)")
	fmt.Println("--- KẾT THÚC DEMO ---")
}
