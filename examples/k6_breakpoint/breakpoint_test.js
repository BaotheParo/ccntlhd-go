import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
    // Kich ban Breakpoint: Tang tai lien tuc va dot ngot de tim diem bao hoa
    stages: [
        { duration: "30s", target: 20 },  // Giai doan 1: Tang nhanh len 20 nguoi dung
        { duration: "30s", target: 50 },  // Giai doan 2: Day len 50 nguoi dung
        { duration: "30s", target: 100 }, // Giai doan 3: "Doi bom" 100 nguoi dung dong thoi
    ],
    thresholds: {
        // He thong se bao loi danh gia (Fail) neu do tre p(95) vuot qua muc 800ms
        "http_req_duration": ["p(95)<800"],
    },
};

export default function () {
    // Goi API lay danh sach su kien
    // Luu y: Dung "http://host.docker.internal:8080" de tro ve may host neu chay tu Docker
    const url = "http://host.docker.internal:8080/api/v1/events";
    
    const res = http.get(url, {
        headers: { "Accept": "application/json" },
    });

    check(res, {
        "Trang thai 200": (r) => r.status === 200,
        "Do tre < 200ms": (r) => r.timings.duration < 200,
    });

    // Nghi rat ngan (0.1 giay) de tao ap luc cuc lon
    sleep(0.1);
}
