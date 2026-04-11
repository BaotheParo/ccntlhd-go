package main

import (
	"errors"
	"fmt"
)

// Khai báo các lỗi chuẩn (Sentinel errors) để dùng chung cho toàn hệ thống
var (
	ErrNotFound     = errors.New("không tìm thấy bản ghi")
	ErrUnauthorized = errors.New("không có quyền truy cập")
)

// ===== TẦNG REPOSITORY: Trả lỗi thô từ Cơ sở dữ liệu =====
func fetchUserFromDB(id int) (string, error) {
	if id <= 0 {
		// Trả về lỗi định dạng thông thường
		return "", fmt.Errorf("fetchUserFromDB: id không hợp lệ: %d", id)
	}
	if id > 100 {
		// Trả về lỗi chuẩn (Sentinel Error) được bọc trong ngữ cảnh của hàm
		return "", fmt.Errorf("fetchUserFromDB: %w", ErrNotFound)
	}
	return fmt.Sprintf("NguoiDung-%d", id), nil
}

// ===== TẦNG SERVICE: Bọc lỗi (Wrap) với ngữ cảnh nghiệp vụ =====
func getUserProfile(id int) (string, error) {
	user, err := fetchUserFromDB(id)
	if err != nil {
		// Sử dụng %w để bọc lỗi gốc (Wrapping), giúp giữ nguyên đặc tính của lỗi bên dưới
		// Điều này cực kỳ quan trọng để tầng Handler có thể "truy vết" ngược lại
		return "", fmt.Errorf("getUserProfile (id=%d): %w", id, err)
	}
	return fmt.Sprintf("Hồ sơ của %s", user), nil
}

// ===== TẦNG HANDLER: Chuyển lỗi thành mã phản hồi HTTP phù hợp =====
func handleGetUser(id int) {
	profile, err := getUserProfile(id)
	if err != nil {
		// Sử dụng errors.Is() để kiểm tra loại lỗi gốc bất kể việc nó đã bị bọc qua bao nhiêu tầng (Wrap)
		if errors.Is(err, ErrNotFound) {
			fmt.Printf("[HTTP 404] %v\n", err)
			return
		}

		// Nếu là lỗi hệ thống không xác định, ẩn chi tiết kỹ thuật đi và chỉ báo lỗi 500 cho người dùng
		fmt.Printf("[HTTP 500] Lỗi hệ thống nội bộ (Chi tiết ẩn: %v)\n", err)
		return
	}
	fmt.Printf("[HTTP 200] %s\n", profile)
}

func main() {
	fmt.Println("--- DEMO ERROR HANDLING (SENTINEL, WRAPPING, IS) ---")
	fmt.Println("")

	fmt.Println("--- KỊCH BẢN 1: TÌM THẤY (Thành công) ---")
	handleGetUser(50)

	fmt.Println("\n--- KỊCH BẢN 2: KHÔNG TÌM THẤY (Lỗi nghiệp vụ - 404) ---")
	fmt.Println("Hành vi: Dùng errors.Is() để phát hiện ErrNotFound dù đã bị Wrap qua 2 tầng.")
	handleGetUser(150)

	fmt.Println("\n--- KỊCH BẢN 3: LỖI THAM SỐ (Lỗi hệ thống - 500) ---")
	fmt.Println("Hành vi: Không khớp với ErrNotFound nên hệ thống trả về lỗi 500 mặc định.")
	handleGetUser(-1)

	fmt.Println("\n--- KẾT THÚC DEMO ---")
}
