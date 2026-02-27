# Hướng dẫn Kiểm thử và Tích hợp Redis (Chống Overselling)

Tài liệu này giải thích cơ chế hoạt động của Redis trong hệ thống Ticketing để ngăn chặn over-selling khi có tải cao (Flash Sale) và cách để bạn có thể test trực tiếp.

## Cơ chế hoạt động (2-Phase Commit)

Hệ thống hoạt động dựa trên cơ chế khóa vé ở Redis trước khi tạo đơn ở Database.
1.  **Phase 1 (Redis Lua Script)**: Khi có request `PlaceOrder`, hệ thống sẽ gọi một đoạn Lua Script trên Redis để kiểm tra tồn kho (`GET`) và tự động trừ đi số vé (`DECRBY`) trong cùng một thao tác duy nhất (Atomic operation). Do Redis bản chất là single-threaded và thực thi script atomic, điều này đảm bảo tuyệt đối không có 2 user cùng mua được 1 vé cuối cùng.
2.  **Phase 2 (PostgreSQL)**: Ngay khi lấy được vé bên Redis, hệ thống lập tức mở một SQL Transaction để lưu `Order`, lưu `OrderItems`, và chạy lệnh `UPDATE remaining_quantity`.
3.  **Rollback Cấp Tốc (Fallback)**: Nếu trong quá trình chạy DB gặp sự cố (mất kết nối, lỗi dữ liệu, ...), DB Transaction bị GORM huỷ (rollback), hệ thống sẽ gọi hàm bù (Compensate) bằng cách gọi lệnh `INCRBY` lên cái Key ở Redis để nhả vé cho người khác mua.

## Cache Warm-up (Khởi tạo dữ liệu vé lên Redis)

Redis ban đầu trống không. Khi tạo một sự kiện bán vé mới ở Database, bạn **BẮT BUỘC** phải ghi số lượng vé (`initial_quantity`) lên Redis. Quá trình này gọi là Cache Warm-up.

Trong ứng dụng hiện tại, tôi đã chuẩn bị sẵn hàm `SetStock` trong `RedisTicketRepository`.

**Cách Warm-up (Ví dụ cho API Tạo vé / Boot server):**
```go
// Giả sử có một TicketTypeID mới tạo
ticketID := uuid.MustParse("d5691060-f4ca-4fbc-84fa-f11ecf3cbef1")
initialStock := 100

// Gọi qua interface cache
err := cacheRepo.SetStock(ctx, ticketID, initialStock)
// Lúc này ở Redis sẽ có key: ticket:d5691060...:stock = 100
```

## Cách Test Tính Năng Redis Overselling Tại Máy Của Bạn!

Bạn có thể chạy thử trực tiếp trên Terminal của mình.

### 1. Khởi động hệ thống (bao gồm Redis)
Chắc chắn Docker của bạn đang chạy:
```bash
docker-compose up -d --build
```

### 2. Thiết lập dữ liệu trên Database (Sử dụng Adminer từ trình duyệt)
-   Truy cập: [http://localhost:8081](http://localhost:8081)
-   Đăng nhập hệ thống (System: `PostgreSQL`, Server: `postgres`, DB: `ticket_db`, user/password: `user`/`password`).
-   Tạo một user mới trong bảng `users`.
-   Tạo một bản ghi trong bảng `ticket_types` với DB ID ví dụ là `4eeaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee`. Giả sử nhập `remaining_quantity = 5`.

### 3. Đẩy dữ liệu vé này lên Redis (Mô phỏng Warm-up trực tiếp)
Vì ta chưa viết Job tự động đồng bộ (bài toán thực tế sẽ cần làm), bạn có thể gõ trực tiếp số lượng vào Redis thông qua CLI của Docker:

```bash
# Mở redis-cli trong container
docker exec -it ccntlhd-go-redis-1 redis-cli

# Thiết lập số lượng vé trong Redis bằng đúng ID vé ở bước 2
SET ticket:4eeaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee:stock 5

# Thoát redis
exit
```

### 4. Gửi Request Bắn Dữ Liệu (Flood Test)
Nếu không rành code Golang Test, bạn có thể dùng **Postman Client**.
*   **Method**: `POST`
*   **URL**: `http://localhost:8080/api/v1/orders` (Có cần Header Bearer JWT Auth theo thiết kế hệ thống của bạn).
*   **Body (JSON)**:
    ```json
    {
        "items": [
            {
                "ticket_type_id": "4eeaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
                "quantity": 1
            }
        ]
    }
    ```

**Thực hành Test**:
-   Bạn hãy nhấn gạch liên tục nút **Send** cực nhanh trên Postman (hoặc dùng Runner của Postman chạy 20 lần cùng 1 lúc).
-   Bạn sẽ thấy đúng **5** requests đầu được ghi nhận (Trả về HTTP 2xx Create Order thành công).
-   Đến request thứ 6 trở đi, nó sẽ báo ngay chữ "hết vé" của Redis! Khẳng định Redis Lua Script đã hoạt động chặn kịp thời, **Database hoàn toàn không bị gọi tạo Order dư thừa!**

### 5. Kiểm tra kết quả
-   Mở lại UI Adminer. Click xem bảng `orders` của DB. Dù bắn bao nhiêu request, bạn sẽ chỉ thấy sinh ra đúng **5 mã Order**.
-   Mở Terminal. Vào lại Redis, chạy thử:
    ```bash
    docker exec -it ccntlhd-go-redis-1 redis-cli GET ticket:4eeaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee:stock
    ```
    -> Kết quả in ra sẽ là `0`.

---
Nếu mọi công đoạn trên ok, tức là hệ thống 2-phase commit với Lua Script đã vận hành hoàn hảo! 🚀
