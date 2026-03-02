# Hướng dẫn Kiểm thử và Tích hợp Redis (Chống Overselling)

Tài liệu này giải thích cơ chế hoạt động của Redis trong hệ thống Ticketing để ngăn chặn over-selling khi có tải cao (Flash Sale) và cách để bạn có thể test trực tiếp.

## Cơ chế hoạt động (Async Worker Pool & 2-Phase Commit)

Hệ thống hoạt động dựa trên cơ chế khóa vé tức thì ở Redis và xử lý tạo đơn hàng Bất Đồng Bộ (Asynchronous) để chịu tải 10,000+ RPS mà không làm sập Database.

1.  **Phase 1 (Redis Lua Script - Chặn Overselling)**: Khi request `PlaceOrder` được gửi tới, hệ thống gọi Lua Script trên Redis kiểm tra tồn kho (`GET`) và trừ vé (`DECRBY`) tự động (Atomic). Script bảo đảm tuyệt đối không có chuyện 2 user cùng mua 1 vé cuối.
2.  **Phase 2 (Queue & Fast Response)**: Khi lấy vé từ Redis thành công, hệ thống **không** gọi Database ngay. Thay vào đó, nó tạo một `OrderPayload` và đẩy vào một hàng đợi trong RAM (Buffered Channel). Sau đó, API trả về ngay cho Client HTTP Status `202 Accepted` (Đơn hàng PENDING đang được xử lý).
3.  **Phase 3 (Background Workers)**: Một nhóm Workers (hiện cấu hình là 50 Goroutines) chạy ngầm liên tục nhặt các `OrderPayload` từ Queue để mở SQL Transaction lưu vào PostgreSQL. Điều này giúp Database chỉ phải chịu một số lượng connections cố định, không bị kiệt sức (Connection Exhaustion) lúc Flash Sale.
4.  **Rollback Cấp Tốc (Fallback)**: Nếu Queue bị đầy (hệ thống quá tải) hoặc lúc Worker insert DB bị lỗi (rớt nhịp DB), hệ thống sẽ lập tức chạy lệnh bù (`INCRBY`) lên Key ở Redis để hoàn trả vé cho người khác lập tức mua được.

## Cache Warm-up (Khởi tạo dữ liệu vé lên Redis)

Redis ban đầu trống không. Khi tạo một sự kiện bán vé mới, bạn **BẮT BUỘC** phải ghi số lượng vé (`initial_quantity`) lên Redis.

Trong ứng dụng hiện tại, tôi đã chuẩn bị sẵn hàm `SetStock` trong `RedisTicketRepository`.

**Ví dụ khởi tạo:**
```go
// Giả sử có một TicketTypeID mới tạo
ticketID := uuid.MustParse("d5691060-f4ca-4fbc-84fa-f11ecf3cbef1")
initialStock := 100

// Gọi qua interface cache
err := cacheRepo.SetStock(ctx, ticketID, initialStock)
// Lúc này Redis: SET ticket:d56910...:stock 100
```

## Cách Test Tính Năng Flash Sale Tại Máy Của Bạn!

Bạn có thể chạy thử trực tiếp trên Terminal của mình.

### 1. Khởi động hệ thống
Chắc chắn Docker của bạn đang chạy:
```bash
docker-compose up -d --build
```

### 2. Thiết lập dữ liệu trên Database (Sử dụng Adminer từ trình duyệt)
-   Truy cập: [http://localhost:8081](http://localhost:8081)
-   Đăng nhập hệ thống (System: `PostgreSQL`, Server: `postgres`, DB: `ticket_db`, user: `user`, pass: `password`).
-   Tạo một user mới trong bảng `users`.
-   Tạo một bản ghi trong bảng `ticket_types` với DB ID ví dụ là `4eeaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee`. Giả sử nhập `remaining_quantity = 5`.

### 3. Đẩy dữ liệu vé này lên Redis (Mô phỏng Warm-up)
Mở Terminal, dùng `redis-cli` trong Docker:

```bash
docker exec -it ccntlhd-go-redis-1 redis-cli
SET ticket:4eeaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee:stock 5
exit
```

### 4. Gửi Request Bắn Dữ Liệu (Flood Test)
Mở **Postman Client**.
*   **Method**: `POST`
*   **URL**: `http://localhost:8080/api/v1/orders` (Kèm Header Bearer JWT Auth).
*   **Body (JSON)**:
    ```json
    {
        "items": [
            {
                "ticket_type_id": "4eeaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
                "quantity": 1,
                "unit_price": "100"
            }
        ]
    }
    ```

**Thực hành Test**:
-   Nhấn nút **Send** cực nhanh trên Postman (hoặc dùng Runner chạy 20 lần cùng lúc).
-   API sẽ phản hồi cực kỳ nhanh mã **HTTP 202 Accepted** cho 5 request đầu tiên lọt qua Redis.
-   Đến request thứ 6 trở đi, nó sẽ lập tức báo lỗi (HTTP 400 - "hết vé") từ Redis trả về!
-   Trong lúc này, Terminal chạy server sẽ log ra "Worker X bắt đầu chạy" báo hiệu PostgreSQL đang thong thả nhặt 5 request thành công hồi nãy để insert Database ngầm.

### 5. Kiểm tra kết quả
-   Mở lại UI Adminer. Chờ vài giây rồi xem bảng `orders` của DB. Dù bắn bao nhiêu request, bạn sẽ chỉ thấy sinh ra đúng **5 mã Order**.
-   Mở Terminal. Vào lại Redis, chạy thử:
    ```bash
    docker exec -it ccntlhd-go-redis-1 redis-cli GET ticket:4eeaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee:stock
    ```
    -> Kết quả in ra sẽ là `0`.

---

