# Hướng dẫn setup & làm việc với Claude Code (cho người mới)

## 1. Tóm tắt lựa chọn (đã chọn sẵn cho bạn)

| Thứ | Chọn | Vì sao |
|---|---|---|
| Server | **Go** | Nhanh, dễ đọc, build ra 1 file chạy, Claude viết Go rất ổn định |
| Client | **TypeScript + PixiJS** chạy trên trình duyệt | Game .io không cần cài đặt, chơi được cả PC lẫn điện thoại, xem thay đổi ngay khi lưu file |
| Database | SQLite | Chỉ là 1 file, không phải cài gì |
| Vẽ bản đồ | Tiled | Miễn phí, kéo thả |
| Skill / plugin | **Không cài thêm gì** | Xem mục 5 |

Vì sao không dùng Rust + Bevy như bản trước: Bevy đổi API liên tục qua mỗi phiên bản nên AI hay viết code cũ không chạy được, compile lâu, đưa lên điện thoại khó. Phần cần hiệu năng của game này nằm ở server (va chạm, 40 nhân vật × 45 kiếm), và Go xử lý thoải mái. Phía trình duyệt chỉ vẽ vài nghìn sprite, PixiJS làm việc này rất nhẹ.

## 2. Cài đặt (làm 1 lần)

1. **Git**: tải tại git-scm.com (Windows cài bản Git for Windows, để mặc định).
2. **Go**: tải bản ổn định mới nhất tại go.dev/dl.
3. **Node.js**: tải bản LTS tại nodejs.org.
4. **VS Code** (để xem code, không bắt buộc): code.visualstudio.com.
5. **Tiled** (vẽ bản đồ, dùng từ M4): mapeditor.org.
6. **Claude Code** (cần gói Claude trả phí: Pro / Max / Team / Enterprise):
   - macOS / Linux: `curl -fsSL https://claude.ai/install.sh | bash`
   - Windows (PowerShell): `irm https://claude.ai/install.ps1 | iex`
   - Mở terminal mới, gõ `claude --version` để kiểm tra; có lỗi thì chạy `claude doctor`.
   - Gõ `claude` lần đầu → trình duyệt mở ra để đăng nhập tài khoản Claude.
   - Nếu không quen terminal: cài extension **Claude Code** (của Anthropic) trong VS Code, hoặc dùng tab Code trong app Claude Desktop — cùng một công cụ, giao diện dễ hơn.
   - Tài liệu chính thức: https://code.claude.com/docs/en/setup

Kiểm tra: mở terminal gõ lần lượt `git --version`, `go version`, `node --version`. Ra số phiên bản là được.

## 3. Tạo dự án

```bash
mkdir kiem-xoay && cd kiem-xoay
mkdir docs
# copy 3 file GAME_SPEC.md, ART_PROMPTS.md, SETUP_GUIDE.md vào thư mục docs/
# copy CLAUDE.md vào thư mục gốc kiem-xoay/
git init
claude
```

## 4. Cách làm việc mỗi phiên

1. Mở terminal trong thư mục `kiem-xoay`, gõ `claude`.
2. Bật **Plan mode** (nhấn `Shift+Tab` tới khi thấy chữ plan) → dán prompt của milestone → Claude đưa kế hoạch → đọc lướt, đồng ý.
3. Để Claude làm. Khi nó hỏi quyền chạy lệnh (go, npm, git) → cho phép.
4. Claude xong → tự chạy game thử theo hướng dẫn nó đưa. Lỗi gì → copy nguyên thông báo lỗi hoặc mô tả bằng lời ("kiếm không đổi chiều khi va nhau") rồi dán cho Claude.
5. Chạy ổn → bảo Claude commit. Làm hỏng → bảo Claude "quay lại commit trước" (Git giữ mọi phiên bản).
6. Xong milestone → gõ `/clear` để bắt đầu phiên sạch cho milestone sau (Claude vẫn đọc CLAUDE.md và docs/PROGRESS.md nên không mất bối cảnh).

Nguyên tắc vàng: **mỗi lần chỉ giao 1 việc vừa phải**. Đừng bảo "làm hết game".

## 5. Về skill / plugin

Không cần cài skill nào từ mạng cho dự án này:
- `CLAUDE.md` + thư mục `docs/` đã cho Claude đủ bối cảnh.
- Skill trên mạng chất lượng thất thường, nhiều cái nhồi thêm hướng dẫn chung chung làm Claude lệch khỏi spec của bạn.
- Skill có thể chứa script chạy trên máy bạn → chỉ dùng của nguồn bạn tin tưởng.

Về sau, nếu có việc lặp đi lặp lại (ví dụ "thêm 1 nhân vật mới vào game từ thư mục ảnh"), cứ bảo Claude: *"Tạo cho tôi 1 skill trong dự án để làm việc X"* — nó tự viết skill riêng đúng với code của bạn.

## 6. Prompt cho từng milestone (copy & dán)

**M0**
```
Đọc CLAUDE.md và docs/GAME_SPEC.md. Làm Milestone M0: tạo cấu trúc repo đúng mục 13.2,
server Go có WebSocket, client Vite + PixiJS. Người chơi là chấm tròn di chuyển bằng WASD,
vị trí do server tính và gửi về. Tạo config/game.json theo mục 16 và docs/PROGRESS.md.
Hướng dẫn tôi từng lệnh để chạy thử.
```

**M1**
```
Làm Milestone M1 theo docs/GAME_SPEC.md: vòng kiếm (mục 4), quái (mục 14), item kiếm,
EXP và 9 cảnh giới (mục 3), scale 1→5 và camera (mục 4.1). Dùng hình placeholder.
Viết unit test cho EXP, scale, giới hạn kiếm. Thêm phím tắt debug để cộng EXP nhanh khi test.
```

**M2**
```
Làm Milestone M2: chém thân và rớt tầng (5.1), va kiếm đổi chiều và bể kiếm theo chênh lệch
cảnh giới (5.2), chết → màn hình Tổng kết → Chơi tiếp (mục 7). Test bằng cách mở 2 tab.
Viết unit test cho mọi trường hợp trong bảng 5.2.
```

**M3**
```
Làm Milestone M3: phòng, matchmaking, bot fill (mục 9), bot AI cơ bản không cần địa hình (mục 10),
trang chủ nhập tên và chọn nhân vật, leaderboard SQLite + API (mục 11), cmd/sim để giả lập.
```

**M4**
```
Làm Milestone M4: đọc bản đồ Tiled (mục 8). Tạo sẵn 3 file .tmj mẫu bằng code nếu tôi chưa vẽ,
và viết hướng dẫn ngắn cách tôi vẽ map trong Tiled (tên layer, thuộc tính object cần đặt).
Vật cản phá được + mọc lại, bụi rậm, bùn, gió, hẻm, cổng. Cập nhật bot AI dùng địa hình.
```

**M5**
```
Làm Milestone M5: toàn bộ vật phẩm buff và Rương Báu (mục 6), nội suy, dự đoán chuyển động,
AOI, không gửi người đang ẩn thân/trong bụi (mục 13.3). Thêm chế độ giả lập ping để test.
```

**M6**
```
Làm Milestone M6: tôi đã đặt art vào assets/ theo docs/ART_PROMPTS.md. Viết script ghép frame
thành atlas, ghép animation theo mood (mục 12), VFX, SFX, UI hoàn chỉnh. Asset nào thiếu thì
giữ placeholder và liệt kê cho tôi.
```

**M7**
```
Làm Milestone M7: tối ưu cho điện thoại (joystick ảo, giao diện dọc/ngang, giới hạn FPS),
rồi hướng dẫn tôi từng bước thuê VPS và deploy có HTTPS (dùng Caddy) để bạn bè vào bằng link.
```

## 7. Chạy thử trên điện thoại khi đang phát triển

Máy tính và điện thoại cùng mạng Wi-Fi → bảo Claude: *"Cho tôi cách mở game trên điện thoại qua mạng LAN"*. Nó sẽ chỉnh Vite cho phép truy cập và đưa địa chỉ dạng `http://192.168.x.x:5173`.

## 8. Khi bị kẹt
- Dán nguyên lỗi cho Claude, kèm "đã làm gì trước khi lỗi".
- Claude sửa 2–3 lần không được → bảo nó *"dừng lại, giải thích nguyên nhân trước khi sửa tiếp"*.
- Muốn đổi luật chơi → sửa `docs/GAME_SPEC.md` (hoặc bảo Claude sửa) **trước**, rồi mới bảo code theo.
- Chỉnh cân bằng (tốc độ, EXP, HP...) → chỉ cần sửa `config/game.json`, không cần đụng code.
