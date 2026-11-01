# Chính sách bảo mật

> **English:** Please report security vulnerabilities **privately** using GitHub's
> *Security → Report a vulnerability* (private vulnerability reporting) on this repository.
> **Do not open a public Issue, Pull Request or discussion** for security problems.

## Báo lỗ hổng bảo mật

Nếu bạn tìm thấy lỗ hổng (ví dụ: vượt quyền trang quản trị, gian lận được nhờ server tin
client, làm sập server từ xa, lộ dữ liệu người chơi), xin báo **riêng tư**:

1. Vào trang repo trên GitHub → tab **Security** → **Report a vulnerability**
   (tính năng *private vulnerability reporting*).
2. Mô tả: lỗi là gì, cách tái hiện từng bước, ảnh hưởng tới đâu, bản build/commit bạn thử.

Đây là kênh duy nhất — dự án **không** dùng email cho báo cáo bảo mật. Báo cáo chỉ bạn và chủ
dự án đọc được (tính năng đã bật từ 02/11/2026).

**Xin đừng** báo lỗ hổng qua Issue, Pull Request, bình luận hay bất cứ chỗ công khai nào — như
vậy là chỉ cho người xấu biết trước khi kịp sửa.

## Sau khi báo

- Chủ dự án sẽ xác nhận đã nhận báo cáo và trao đổi với bạn ngay trong báo cáo riêng đó.
- Lỗi được sửa trước, rồi mới công bố. Bạn được ghi công nếu muốn.
- Đây là dự án cá nhân, không có chương trình trả thưởng (bug bounty).

## Phạm vi

- Mã nguồn trong repo này (server Go, client, công cụ).
- Bản đang chạy tại tên miền chính thức của game.

Xin **không** thử tấn công từ chối dịch vụ, không đọc/sửa dữ liệu của người chơi khác, và
không làm ảnh hưởng tới người đang chơi khi kiểm tra.
