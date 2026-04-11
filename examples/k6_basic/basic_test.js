import http from "k6/http";
import { check, sleep } from "k6";

// Cấu hình kịch bản kiểm thử tải (Performance Testing)
export const options = {
    stages: [
        { duration: "10s", target: 5 },  // Giai đoạn 1: Tăng dần lên 5 người dùng ảo (VU)
        { duration: "20s", target: 10 }, // Giai đoạn 2: Duy trì tải ở mức 10 VU
        { duration: "10s", target: 0 },  // Giai đoạn 3: Giảm tải về 0 (Ramp-down)
    ],
    thresholds: {
        // ĐỊNH NGHĨA KHOẢNG CHẤP NHẬN: 95% yêu cầu phải phản hồi nhanh hơn 500ms
        "http_req_duration": ["p(95)<500"], 
        // TỶ LỆ LỖI: Tổng số yêu cầu thất bại phải dưới 1%
        "http_req_failed": ["rate<0.01"],   
    },
};

// Hàm thực thi luồng người dùng (Main Function)
export default function () {
    // Gửi yêu cầu GET tới dịch vụ giả lập
    const res = http.get("https://httpbin.org/get");
    
    // Kiểm tra và xác nhận tính đúng đắn (Assertions)
    check(res, {
        "Trạng thái HTTP là 200": (r) => r.status === 200,
        "Thời gian phản hồi < 500ms": (r) => r.timings.duration < 500,
    });
    
    // Thời gian suy nghĩ của người dùng giả lập
    sleep(1); 
}
