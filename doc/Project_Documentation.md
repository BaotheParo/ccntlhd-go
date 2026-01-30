### 1. Module này làm gì?
Đây là phần xử lý đặt vé trong hệ thống bán vé chịu tải cao (high-concurrency). Mục tiêu chính: **không để bán quá số vé** (overselling), dù có 1000 người bấm đặt cùng lúc.

Dùng **Clean Architecture** nên code sạch, dễ mở rộng sau này.  
Cách chống race condition: dùng **Pessimistic Locking** (`FOR UPDATE`) của PostgreSQL – khóa dòng vé đang xử lý để thằng khác không đụng vào được.

Các thứ chính:
- `TicketType`: quản lý loại vé, số lượng còn lại, giá.
- `Order`: đơn hàng của user.
- `OrderItem`: snapshot giá vé lúc mua (để sau này tính tiền không bị thay đổi).

Repository thì có hàm `GetTicketTypeForUpdate` – cái này siêu quan trọng, nó khóa dòng vé trước khi trừ số lượng.

Service thì có `PlaceOrder` – chạy trong transaction, kiểm tra kho → tính tiền → tạo đơn → commit. Nếu fail thì rollback hết.

### 2. Chạy thử cho nhanh
Dùng Docker Compose là ngon nhất, toàn bộ stack (app + Postgres + Redis) chạy một lệnh.

```bash
docker-compose up --build -d
```

Sau khi chạy:
- App: localhost:8080
- Postgres: localhost:5433 (không đụng cổng 5432 local của mày)
- Redis: localhost:6380

Lý do đổi cổng: vì hầu như máy ai cũng có Postgres/Redis chạy local rồi, đụng cổng là chết luôn.

### 3. Test concurrency (kiểm tra bán quá vé)
Có file test tích hợp siêu hay: `tests/integration/concurrency_test.go`

Nó giả lập nhiều user đặt vé cùng lúc, xem có bị overselling không.

Chạy test:
```bash
docker-compose -f docker-compose.yml -f docker-compose.test.yml run --rm test-runner
```

Test này sẽ:
- Kết nối thẳng vào Postgres đang chạy (từ docker-compose chính).
- Tạo dữ liệu test (event, ticket type, user).
- Bắn 50–100 request đồng thời đặt vé.
- Check tổng vé bán ra có vượt quá kho không.

Nếu pass → yên tâm, locking hoạt động tốt.

### 4. Những lỗi từng gặp (để anh em khỏi mất công search Google)
| Lỗi gì?                              | Triệu chứng / thông báo                           | Tại sao?                                      | Fix thế nào?                                                                 |
|--------------------------------------|---------------------------------------------------|-----------------------------------------------|-----------------------------------------------------------------------------|
| Docker build treo ở go mod tidy      | Build mãi không xong                              | Cache module lỗi hoặc mạng chập chờn          | Chạy `go clean -modcache` local trước, hoặc để Docker tự build lại.        |
| Bind port 5432/6379 fail             | "address already in use"                          | Máy local đang chạy Postgres/Redis            | Đổi cổng host trong docker-compose: 5433 → 5432, 6380 → 6379.              |
| Foreign key constraint violation     | "ticket_types_event_id_fkey"                      | Test tạo vé mà chưa có event cha              | Thêm INSERT event + user bằng SQL trước khi test.                           |
| null value in column "unit_price"    | Không insert được vì cột null                     | Struct tag sai, GORM map sai cột              | Thêm `gorm:"column:price"` vào field `UnitPrice` trong entity.              |
| created_at bị null                   | Transaction không tự fill timestamp               | GORM không luôn auto-fill trong transaction   | Set thủ công `CreatedAt: time.Now()` và `UpdatedAt: time.Now()` trong service. |
| unit_price not-null violation        | Cột cũ tồn tại từ migration lỗi trước             | Migration cũ tạo cột thừa                     | Trong test setup: `db.Migrator().DropTable(&OrderItem{})` rồi migrate lại.  |
| Biên dịch lỗi shadowed variable      | userID bị che bởi biến vòng lặp                   | Khai báo lại userID kiểu int trong for loop   | Đổi tên biến vòng lặp thành `workerID`, giữ `userID` là uuid.UUID.         |

### 5. Code chính nằm ở đâu?
- Entity: `internal/core/entity/event.go`, `order.go`
- Repo: `internal/adapter/repository/order_repository.go` (có hàm lock FOR UPDATE)
- Service: `internal/core/service/order_service.go` (logic transaction)
- Test concurrency: `tests/integration/concurrency_test.go`

Xong phần này là module đặt vé đã khá chắc chắn rồi. Nếu anh em nào chạy test fail hoặc gặp lỗi lạ, cứ paste log vào group, fix chung cho nhanh.

Chạy thử đi, có gì báo tao nhé! 🍻