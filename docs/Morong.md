Trong sơ đồ tổng quát, bạn nên thêm một nhóm chức năng dành riêng cho Admin (Người quản lý các ông IT HD):

**Use Case chính:** Audit Incident Operations

**Các thành phần con:**
* `<<include>>` **View Action Timeline:** Xem lại toàn bộ lịch sử biến động của một sự cố (Ai nhận, lúc nào, ghi chú gì).
* `<<include>>` **Trace Verification Proof:** Xem lại bằng chứng quét Baseline mà hệ thống đã thực hiện trước khi IT đóng Case.
* `<<extend>>` **Export Audit Report:** Xuất báo cáo tuân thủ để phục vụ hậu kiểm.

## 2. Luồng Audit khép kín (The Audit Trail)
Phần Audit này sẽ chạy song song và ghi lại mọi "dấu vết" trong luồng quản lý sự cố mà chúng ta đã bàn:

### A. Log mọi thay đổi trạng thái (State Change Logs)
Mỗi khi Case chuyển từ `Open` $\rightarrow$ `In Progress` $\rightarrow$ `Resolved`, hệ thống phải ghi lại:
* **User ID:** Ai là người thực hiện?
* **Timestamp:** Thời gian chính xác đến từng mili giây.
* **Ip Address:** Ông IT đó thao tác từ máy nào?

### B. Bằng chứng kỹ thuật (Technical Evidence)
Đây là phần quan trọng nhất để chống "gian lận":
* Nếu IT chọn đóng Case là "OS Reinstalled": Hệ thống Audit phải ghi lại trạng thái của Agent sau đó (Ví dụ: Agent mới được Enroll, Hardware ID cũ nhưng OS Install Date mới).
* Nếu IT chọn "Resolved": Hệ thống Audit đính kèm bản quét Baseline Scan Result (đã ký số hoặc băm Hash) để chứng minh máy thực sự đã sạch tại thời điểm đó.

### C. Nhật ký điều tra (Investigation Journal)
Toàn bộ các Resolution Note và các file đính kèm (ảnh chụp màn hình lỗi, log tay...) mà IT nhập vào đều được lưu trữ không cho phép chỉnh sửa (Immutable logs).

## 3. Cấu trúc bảng Audit (Gợi ý cho Backend Go/SQL)
Để phục vụ tính năng này, bạn nên có một bảng `Incident_Audits` với cấu trúc đại loại như sau:

| Field | Description |
| :--- | :--- |
| `incident_id` | Liên kết với sự cố gốc. |
| `actor_id` | ID của nhân viên IT thực hiện thao tác. |
| `action_type` | `ASSIGN`, `INVESTIGATE`, `TRIGGER_SCAN`, `CLOSE`. |
| `pre_status` | Trạng thái trước khi thao tác (ví dụ: `Open`). |
| `post_status` | Trạng thái sau khi thao tác (ví dụ: `In Progress`). |
| `evidence_data` | JSON chứa kết quả quét Baseline hoặc link tới Log điều tra. |
| `audit_hash` | Mã băm để đảm bảo bản ghi Audit không bị sửa đổi trong DB. |

## 4. Gợi ý Note cho sơ đồ Use Case của bạn
Bạn nên thêm một cái Note màu vàng nổi bật cạnh Actor Admin:

> **Audit Policy:**
> "Mọi thao tác của IT HD trên hệ thống đều được ghi lại vĩnh viễn (Immutable Audit Logs)."
> "Bằng chứng quét Baseline là điều kiện bắt buộc để xác thực tính đúng đắn của việc đóng sự cố."

## Đánh giá tổng quát
Việc thêm Audit giúp hệ thống của bạn chuyển từ một công cụ "hỗ trợ làm việc" sang một công cụ "quản trị rủi ro". Đúng chất dân làm bảo mật: 
> *"Never trust, always verify, and ALWAYS LOG."* (Không bao giờ tin, luôn luôn xác thực và LUÔN LUÔN GHI LOG).