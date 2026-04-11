import http from "k6/http";
import { check, sleep, group } from "k6";

export const options = {
    vus: 10,
    duration: "20s",
    thresholds: {
        // Dat nguong danh gia rieng biet cho tung loai yeu cau thong qua the (Tag)
        "http_req_duration{type:get}": ["p(95)<300"], 
        "http_req_duration{type:post}": ["p(95)<500"], 
    },
};

export default function () {
    // Gom nhom (Group) de tach biet du lieu thong ke cua yeu cau GET
    group("Yeu cau GET", function () {
        const r = http.get("https://httpbin.org/get", { tags: { type: "get" } });
        check(r, { "GET thanh cong (Ma 200)": (res) => res.status === 200 });
    });

    sleep(0.5);

    // Gom nhom (Group) de tach biet du lieu thong ke cua yeu cau POST
    group("Yeu cau POST", function () {
        const payload = JSON.stringify({ key: "value", timestamp: Date.now() });
        const params = {
            headers: { "Content-Type": "application/json" },
            tags: { type: "post" },
        };
        const r = http.post("https://httpbin.org/post", payload, params);
        check(r, { "POST thanh cong (Ma 200)": (res) => res.status === 200 });
    });

    sleep(0.5);
}
