# Rà soát: chỗ nào khoá theo `id` entity, chỗ nào từng giả định "id theo khu"

Ngày 31/10/2026. Lý do rà: tìm thấy **id entity trùng giữa các khu** — mỗi khu có bộ đếm
`nextID` riêng bắt đầu từ 0, nên khu 0 và khu 3 cùng sinh ra entity số 7.

Một phòng có 5 khu chạy song song, nhưng **client, bảng xếp hạng và trang quản trị nhìn cả
phòng**. Chỗ nào gom dữ liệu của nhiều khu mà khoá theo `id` thì hai entity khác nhau ở hai
khu đè lên nhau.

## Cách sửa: tiền tố khu trong chính id

```go
// server/internal/game/game.go
idBase: uint32(zoneIndex) << 24,

func (g *Game) newID() uint32 {
	g.nextID++
	return g.idBase | g.nextID
}
```

8 bit cao là số khu, 24 bit thấp là bộ đếm → **16 777 216 entity mỗi khu**. Khu đông nhất
trong bài kiểm dùng 7 265, tức còn thừa 2 300 lần.

Vì sao tiền tố chứ không phải một bộ đếm chung cho cả phòng: bộ đếm chung thì 5 goroutine
khu phải khoá lẫn nhau mỗi lần sinh một viên linh khí. Tiền tố cho cùng kết quả mà không
cần đồng bộ gì — mỗi khu vẫn đếm riêng trong khoảng riêng của nó.

Vì sao `zoneIndex` là **tham số bắt buộc của `game.New`** chứ không phải một hàm
`SetIDBase()` gọi sau: bản đầu tôi làm đúng kiểu đó, kèm bình luận "New chưa sinh entity
nào". Bình luận đó **sai** — `seedOrbs()` chạy ngay trong `New`. Kết quả: quái có tiền tố
đúng, còn **linh khí vẫn trùng**. Nửa lỗi, khó thấy hơn cả lỗi gốc. Giờ trình biên dịch
không cho phép quên.

## Bảng rà soát

| Chỗ khoá theo id | Phạm vi | Có hỏng không | Vì sao |
| --- | --- | --- | --- |
| `c.knownPlayers`, `knownMonsters`, `knownItems` | một client | Không | `resetKnown()` xoá sạch khi qua cổng (`room.go:229`) |
| `c.aoi` (AOI có trễ) | một client | Không | `aoi.Forget()` trong cùng `resetKnown()` |
| `Despawn` ids | một client | Không | Sinh ra từ `knownX` nói trên, đã sạch |
| `c.lastWire` (delta Protocol v2) | một client | **CÓ — đã sửa** | xem mục dưới |
| `LeaderEntry.ID` → `rankAt` | **cả phòng** | **CÓ — đã sửa** | xem mục dưới |
| `livePlayer.ID`, `mp.follow` (bản đồ live quản trị) | **cả phòng** | **CÓ — khỏi bệnh nhờ sửa gốc** | Trang quản trị gom người chơi của mọi khu vào một danh sách rồi tìm theo id. Hai người cùng id thì `find()` trả về người sai, và đi theo một người lại thấy chấm sáng ở khu khác |
| `hidden_lives` khoá `(guest_id, ended_at)` | cả server | Không | Không dùng id entity. `guest_id` do client sinh, sống lâu hơn một phiên |
| `admin_audit.target` | cả server | Không | Cũng là `guest_id` |

Hai chỗ **không** hỏng vì cùng một lý do: chúng thuộc về **một client**, và client nào đổi
khu thì bộ nhớ của nó bị xoá — nên không bao giờ có id của hai khu nằm cạnh nhau. Đó là
tình cờ may, không phải thiết kế: không ai viết `resetKnown` để chống trùng id.

## Hai chỗ hỏng thật

### 1. Danh hiệu Top 1 tính theo KHU, không theo phòng

`updateTitles()` xếp hạng trong `g.Ranking()` — tức là trong khu này. Một phòng 5 khu phát
ra **năm vương miện Top 1**, và hai người trong số đó gặp nhau qua cổng thì cả hai đều đeo.

Sửa: `game.RoomRank func(id uint32) int` do phòng cắm vào mỗi tick. `nil` thì rơi về
`g.Ranking()` — giữ cho `cmd/sim` và test một-khu chạy được.

Chỉ đúng được nhờ id duy nhất cả phòng: `rankOf2(id)` tra một bảng gom của 5 khu.

Test: `title_room_test.go` — `TestTop1IsRoomWideNotPerZone`,
`TestTitleFollowsPlayerAcrossZones`.

### 2. Ma của khu cũ — lỗi do chính Protocol v2 tạo ra

`resetKnown()` có từ trước và quên `knownPlayers`. Nhưng `lastWire` và `keyframeAt` là hai
trường tôi **thêm vào mà không nối vào đó**. Hậu quả sau mỗi lần qua cổng:

- Bộ giải mã bên client giữ nguyên entity của khu CŨ và **vẽ chúng ra mỗi snapshot** cho
  tới keyframe kế tiếp — người chơi thấy người và quái của khu vừa rời.
- `lastWire` phình dần: entity khu cũ không bao giờ được dọn.

Sửa lần đầu của tôi là `c.keyframeAt = 0`, và nó **vẫn sai**. `keyframeAt` so với `s.Time`,
là đồng hồ CỦA MỘT KHU; mỗi khu tick riêng. Khu cũ ở giây 500, khu mới ở giây 10 → hiệu số
âm → client mất keyframe suốt 490 giây. Đúng là một cờ `forceKeyframe` riêng, không phụ
thuộc đồng hồ nào.

Test: `snapwire_test.go` — `TestZoneChangeForcesKeyframe` (kiểm cả "có keyframe" và "không
còn Update nào so với trạng thái khu cũ").

## Bài kiểm chống tái phát

`idunique_test.go`:

- `TestEntityIDsUniqueAcrossZonesAfter1000Transfers` — 1 000 lượt qua cổng, gom id của mọi
  entity ở mọi khu sau từng lượt: **37 876 entity, 0 id trùng**, và id người chơi **giữ
  nguyên 7226** suốt cả 1 000 lượt (qua cổng không sinh lại id).
- `TestZoneIDRangesDoNotOverlap` — 5 khoảng id rời nhau, báo luôn khu đông nhất dùng bao
  nhiêu phần khoảng của nó, để nếu sau này có ai nâng số entity mỗi khu thì bài kiểm nói
  trước khi 24 bit hết chỗ.
