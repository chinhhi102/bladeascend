# Hướng dẫn vẽ bản đồ bằng Tiled

Game đọc file `.tmj` (Tiled JSON) trong thư mục `maps/`. Có sẵn 3 bản đồ do code tạo:
`lang_tre.tmj` (Làng Tre), `cho_dem.tmj` (Chợ Đêm), `nui_da.tmj` (Núi Đá).
Muốn tạo lại từ đầu: `cd server && go run ./cmd/mapgen` (sẽ **ghi đè** 3 file trên).

## 1. Mở / tạo bản đồ
1. Cài Tiled (mapeditor.org) → **File → Open** một file trong `maps/` (cách dễ nhất: sửa từ bản có sẵn).
2. Tạo mới: **File → New Map** → Orthogonal, Tile size **100 × 100**, Map size **120 × 120** tiles (= 12 000 × 12 000 đơn vị).
3. **Map → Map Properties** → thêm 2 thuộc tính kiểu `string`:
   - `name`: tên hiển thị, ví dụ `Làng Tre`
   - `biome`: `grass` | `concrete` | `dirt` (chọn ảnh nền)
4. Lưu bằng **File → Save As…** chọn định dạng **JSON map files (*.tmj)** vào thư mục `maps/`. Tên file = id bản đồ (chữ thường, không dấu, ví dụ `bai_bien.tmj`).
5. Khởi động lại server → bản đồ mới được chọn ngẫu nhiên cho phòng mới.

## 2. Các layer (Object Layer) — tên phải đúng
Tạo 4 **Object Layer** (Layer → New → Object Layer), đặt tên đúng chữ thường:

| Layer | Dùng để | Hình |
|---|---|---|
| `obstacles` | Vật cản chặn người | Hình chữ nhật hoặc hình tròn (Ellipse, giữ Shift để tròn đều) |
| `zones` | Bụi rậm, bùn, gió (không chặn) | Chữ nhật hoặc tròn |
| `gates` | Cổng dịch chuyển | Hình tròn bất kỳ (lấy tâm) |
| `decor` | Đồ trang trí (chỉ để nhìn) | Điểm (Point) hoặc hình bất kỳ |

Không xoay object (rotation) — game bỏ qua góc xoay.

## 3. Đặt loại cho object
Chọn object → khung **Properties** → ô **Class** (Tiled 1.9+) hoặc **Type** (bản cũ) → gõ một trong các giá trị:

**Layer `obstacles`:**

| Class | Vật cản | HP (trong `config/game.json`) |
|---|---|---|
| `fence` | Hàng rào tre (mỏng, dài) | 30 |
| `stall` | Sạp hàng (vỡ có 30% rơi đồ) | 60 |
| `cart` | Xe đẩy (vỡ có 30% rơi đồ) | 60 |
| `brick` | Tường gạch | 150 |
| `boulder` | Đá tảng (nên vẽ hình tròn to) | 400 |
| `stone_wall` | Tường đá cổ — **không phá được** | ∞ |

Loại nào không có trong bảng HP của config thì coi là không phá được.

**Layer `zones`:**

| Class | Hiệu ứng | Thuộc tính thêm |
|---|---|---|
| `bush` | Bụi rậm: người ngoài không thấy người trong | – |
| `mud` | Bùn: chạy chậm ×0.6 | – |
| `wind` | Đường gió: theo gió ×1.5, ngược ×0.8 | `dir` (float, độ): 0 = sang phải, 90 = xuống dưới, 180 = sang trái, 270 = lên trên |

**Layer `gates`:** Class `gate`, thêm thuộc tính kiểu `string`:
- `network`: tên mạng cổng. Các cổng cùng tên nối với nhau. 2 cổng = cặp A↔B; 3–4 cổng = vào 1 cổng ra ngẫu nhiên cổng khác.
- `color`: `purple` | `blue` | `orange` — mỗi mạng một màu để người chơi đoán.

**Layer `decor`:** Class = tên ảnh trong `assets/decor/` không có `.png`, ví dụ `red_plastic_stool`, `banana_tree`, `lazy_sleeping_dog`, `stone_lantern`…

## 4. Mẹo thiết kế
- **Hẻm hẹp**: đặt 2 bức `stone_wall` song song, khe hở 150 / 250 / 350. Người to (Đại Thừa có thân rộng 300) không chui lọt khe 150 → người nhỏ trốn vào được.
- Để trống khu giữa bản đồ và quanh cổng (~300 đơn vị) cho người chơi xuất hiện.
- Khoảng 150–300 vật cản là vừa (server chạy tốt). Quái và vật phẩm tự rải, không cần vẽ.
- Không cần vẽ tường biên — mép bản đồ tự chặn.
- Sửa xong → lưu → khởi động lại server. Lỗi đọc file sẽ hiện trong cửa sổ server.
