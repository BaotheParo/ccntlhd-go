import http from 'k6/http';
import { check, sleep } from 'k6';

// 1. Cau hinh kich ban Load Test - Flash Sale
export const options = {
    stages: [
        { duration: '5s', target: 100 },  // Ramp-up: Tang len 100 VUs trong 5s
        { duration: '10s', target: 200 }, // Hold: Giu 200 VUs lien tuc ban request trong 10s
        { duration: '5s', target: 0 },    // Ramp-down: Ha ve 0 VUs
    ],
    // Nguong Fail: Yeu cau 95% request phai co response < 500ms
    thresholds: {
        http_req_duration: ['p(95)<500'],
    },
};

// 2. Dinh nghia URL va Token (Uu tien dung bien moi truong tu -e)
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080/api/v1';
const BEARER_TOKEN = __ENV.TOKEN || 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNjViNzZlYTAtMjZiYi00MDI0LWE0Y2UtODkxM2E0YTk2YmQ1Iiwicm9sZSI6InVzZXIiLCJleHAiOjE3NzU2MjU4MDR9.zXhzskSPrzUyyLhdxNeDfJptwLruklAQDvDbvXECHRY'; 
const TICKET_TYPE_ID = __ENV.TICKET_ID || 'f88e7256-549a-46a3-b03d-4c802da8e2fb';

export default function () {
    const url = `${BASE_URL}/orders`;
    
    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${BEARER_TOKEN}`,
        },
    };

    const payload = JSON.stringify({
        items: [
            {
                ticket_type_id: TICKET_TYPE_ID,
                quantity: 1,
                unit_price: "200000"
            }
        ]
    });

    // 3. Gui Request POST
    const res = http.post(url, payload, params);

    // 4. Kiem tra tinh dung đan (Checks)
    check(res, {
        'He thong phan hoi 202 (Accepted) hoac 400 (Het ve)': (r) => {
            return r.status === 202 || r.status === 400;
        },
        'Thoi gian xu ly request < 1 giay': (r) => r.timings.duration < 1000,
        'Khong co loi he thong 5xx': (r) => r.status < 500,
    });

    // Nghi ngau nhien giua cac request
    sleep(Math.random() * 0.1 + 0.05);
}
