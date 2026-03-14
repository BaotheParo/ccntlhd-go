import http from 'k6/http';
import { check, sleep } from 'k6';

// 1. K6 Options - Cấu hình mô phỏng Flash Sale VUs (Virtual Users)
export const options = {
    stages: [
        { duration: '5s', target: 100 },  // Ramp-up: Tăng tốc lên 100 VUs trong 5s đầu
        { duration: '10s', target: 100 }, // Hold: Giữ nguyên mức 100 VUs liên tục bắn request trong 10s (BÃO TRAFFIC)
        { duration: '5s', target: 0 },    // Ramp-down: Hạ từ từ số lượng VUs về l trong 5s
    ],
    // Ngưỡng Fail: Yêu cầu 95% request phải có response < 500ms
    thresholds: {
        http_req_duration: ['p(95)<500'],
    },
};

// 2. Định nghĩa hằng số URL và Auth (Thay thế Token thật của bạn ở đây)
// Lưu ý: Đang chạy k6 bằng Docker, nên cần dùng host.docker.internal để gọi vào localhost của máy Host.
const BASE_URL = 'http://host.docker.internal:8080/api/v1';

// [TODO: REPLACE THIS TOKEN] Lấy Token được cấp từ /api/v1/users/login
const BEARER_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzM1NDI1MzQsInN1YiI6IjU1MGU4NDAwLWUyOWItNDFkNC1hNzE2LTQ0NjY1NTQ0MDAwMCIsInVzZXJfaWQiOiI1NTBlODQwMC1lMjliLTQxZDQtYTcxNi00NDY2NTU0NDAwMDAifQ.1V1KjXWjxkFsexzwZMdv-DfVHgo-w_pch_dHUb2B88A'; 

// [TODO: REPLACE THIS TICKET ID] Lấy ID của vé bạn muốn flash sale 
const TICKET_TYPE_ID = '770e8400-e29b-41d4-a716-446655440002';

export default function () {
    const url = `${BASE_URL}/orders`;
    
    // Headers cấu hình API (Kèm JWT Token)
    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${BEARER_TOKEN}`,
        },
    };

    // Body Đặt mua 1 vé (với giá unit_price là 100.00 cho sẵn hệ thống bỏ qua kiểm tra DB)
    const payload = JSON.stringify({
        items: [
            {
                ticket_type_id: TICKET_TYPE_ID,
                quantity: 1,
                unit_price: "100.00"
            }
        ]
    });

    // 3. Gửi Request POST tới App
    const res = http.post(url, payload, params);

    // 4. Verification Checkers
    // Trong Flash sale mua vé, Server trả 202 Accepted là Thành Công, hoặc 400 Bad Request kèm "insufficient ticket stock" là hết vé (Vẫn đúng luồng logic chặn Oversell).
    check(res, {
        'Hệ thống phản hồi 202 (Processing) hoặc 400 (Hết vé/Sold out)': (r) => {
            return r.status === 202 || r.status === 400;
        },
        'Thời gian CPU xử lý 1 request nhỏ hơn 1 giây': (r) => r.timings.duration < 1000,
    });

    // Nghỉ chân ngẫu nhiên 50ms - 100ms giữa mỗi vòng lặp user để giả lập hành vi gõ bàn phím
    sleep(Math.random() * 0.1 + 0.05);
}
