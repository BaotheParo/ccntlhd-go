import http from "k6/http";
import { check, sleep } from "k6";

// Su dung host.docker.internal neu chay k6 tu Docker container
const BASE_URL = __ENV.BASE_URL || "http://host.docker.internal:8080";

export const options = {
    stages: [
        { duration: "10s", target: 10 },
        { duration: "20s", target: 50 },
        { duration: "10s", target: 0 },
    ],
};

// Ham setup() chi thuc thi duy nhat 1 lan khi bat dau kiem thu
export function setup() {
    const payload = JSON.stringify({
        email: "admin@example.com",
        password: "admin"
    });
    
    // Dang nhap de lay ma thong bao (JWT Token)
    const res = http.post(`${BASE_URL}/api/v1/auth/login`, payload, {
        headers: { "Content-Type": "application/json" }
    });

    check(res, { "Dang nhap khoi tao thanh cong": (r) => r.status === 200 });
    
    const token = res.json("data.token") || res.json("token");
    // Tra ve Token de phan phoi cho toan bo nguoi dung ao dung chung
    return { token: token };
}

// Ham default() tiep nhan du lieu tu ham setup()
export default function (data) {
    const params = {
        headers: { 
            "Content-Type": "application/json",
            "Authorization": `Bearer ${data.token}`
        },
    };

    // Goi API Quan tri yeu cau co xac thuc
    const res = http.get(`${BASE_URL}/api/v1/admin/users`, params);
    
    check(res, { "Lay danh sach nguoi dung Quan tri thanh cong": (r) => r.status === 200 });

    sleep(1);
}
