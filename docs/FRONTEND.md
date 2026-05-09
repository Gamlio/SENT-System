Git Hash: [Latest Update - SOC Dashboard Realtime]

# 🌐 Hướng dẫn Kiến trúc Frontend (SENT-SYSTEM)

Tài liệu này mô tả kiến trúc tầng Frontend của nền tảng Sentinex SOC, được xây dựng dựa trên React (Vite), Tailwind CSS và WebSocket cho khả năng xử lý Real-time.

## 1. Phong cách thiết kế (UI/UX)
- **Giao diện SOC/Cyberpunk:** Sử dụng tông nền Deep Navy (`#050B14`, `#0A101D`). 
- **Color Scheme:**
  - `Emerald (#10b981)`: Hệ thống an toàn, Zero-Trust.
  - `Red (#ef4444)`: Cảnh báo nguy hiểm (Critical, High Risk).
  - `Indigo (#6366f1)`: Thành phần tương tác, hệ thống.
  - `Orange/Purple`: Phân loại Alert/Incident.
- **Hiệu ứng:** Sử dụng Glow (box-shadow neon) và Pulse Animation (chớp tắt) để điều hướng sự chú ý của Admin tới các sự cố nóng.

## 2. Kiến trúc Real-time (WebSocket Integration)

Hệ thống nhận log từ hàng nghìn máy trạm (assets), vì vậy việc liên tục F5 (Polling) là không thể. Frontend sử dụng một luồng WebSocket (`WebSocketContext`) duy nhất (Singleton) và phân tán sự kiện tới các components con qua Custom Hook.

### A. Global Hub & Socket Context
- `WebSocketContext` bao bọc toàn bộ App. Nó tự động duy trì kết nối (Heartbeat/Ping) với Go Backend và cung cấp hàm `subscribe()`.
- Khi nhận được Message, nó sẽ dùng cơ chế Pub/Sub nội bộ để tìm và trigger tất cả các hàm callback đã đăng ký với Event tương ứng.

### B. Đăng ký nhận sự kiện (useSocketSubscription)
Thay vì phải quản lý logic mount/unmount thủ công, Frontend cung cấp hook `useSocketSubscription`.

```javascript
// Lắng nghe nhiều Event và thực hiện gọi lại API (Kèm Debounce 500ms)
useSocketSubscription(['NEW_INCIDENT', 'ASSET_STATUS_CHANGED'], () => {
    fetchDashboardData();
}, 500);

// Lắng nghe Event đếm số lượng Real-time (Không Debounce)
useSocketSubscription('NEW_ALERT_TICK', (payload) => {
    setLiveTotalAlerts(prev => prev + 1);
});
```

### C. Xử lý "Bão Log" (Log Flood) với Debounce
Khi có hàng chục thiết bị cùng đẩy Log (Ví dụ: Một cuộc tấn công USB hàng loạt), WebSocket sẽ nhận vài chục packets mỗi giây. 
- Việc chạy trực tiếp hàm callback (nhất là gọi API bằng Axios) sẽ làm treo trình duyệt. 
- Hook `useSocketSubscription` hỗ trợ tham số thứ 3 là `debounceMs`. 
- Sử dụng `debounceMs` đảm bảo Frontend chỉ re-render 1 lần duy nhất sau khi luồng dữ liệu dồn dập đã kết thúc.

## 3. Quản lý Biểu đồ (Recharts)

Các biểu đồ phức hợp (Radar, Area, Donut) được tối ưu hóa như sau:
1. **ResponsiveContainer:** Luôn bọc ngoài để tương thích mọi kích thước màn hình.
2. **useMemo Data:** Dữ liệu cấp cho biểu đồ bắt buộc phải được bọc trong `useMemo` để ngăn việc Component con bị re-render không cần thiết khi Component cha đổi State.
3. **Live Streaming:** Với biểu đồ như Area (Trend), sử dụng state mảng trực tiếp. Khi có event Socket, `push` vào phần tử mảng ở Index cuối cùng để biểu đồ có hiệu ứng nhích từ từ, mô phỏng Live Telemetry như màn hình chứng khoán.

## 4. Cấu trúc Component cốt lõi
- `Dashboard.jsx`: Khung xương Layout (Tier 1 KPI, Tier 2 Charts, Tier 3 Tables). Quản lý việc lấy dữ liệu (Data Fetching).
- `DashboardCharts.jsx`: Lọc và Map dữ liệu (từ API + Socket) chuyển hóa thành Format của Recharts.
- `KpiCard`: Các thẻ KPI được tái sử dụng, có hỗ trợ prop `isAlert` để kích hoạt giao diện cảnh báo (Pulse Red).

## 5. Quy tắc Lập trình (Best Practices)
1. **Null-check Data:** Dữ liệu Backend trả về đôi khi có thể rỗng hoặc chưa mount kịp. Luôn dùng Optional Chaining (`stats?.summary?.total_alerts || 0`).
2. **Iconography:** Chỉ sử dụng bộ `lucide-react`.
3. **Thanh cuộn:** Sử dụng class `custom-scrollbar` để thanh cuộn nhìn tinh tế (thin, gray track) cho các vùng nội dung tràn.