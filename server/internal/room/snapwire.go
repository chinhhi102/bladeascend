package room

import (
	"math"

	"bladeascend/server/internal/game"
	"bladeascend/server/internal/netproto"
	"bladeascend/server/internal/wire"
)

// Mã hoá snapshot nhị phân — Protocol v2 (NETCODE_SPEC §4.1, §4.3, §4.6).
//
// Tệp này KHÔNG quyết định gửi gì cho ai. Việc đó vẫn do `snapshotFor` làm, với toàn bộ
// luật AOI, ngân sách byte, nhịp 10/20 Hz và vòng đời Despawn đã có test. Ở đây chỉ nhận
// một `netproto.Snapshot` đã dựng xong rồi biến nó thành byte.
//
// Tách bạch như vậy là chủ ý: đổi CÁCH MÃ HOÁ và đổi LUẬT GỬI cùng lúc là cách chắc chắn
// nhất để một con số sai và không biết nó sai vì phần nào.
//
// ── Khung gói ───────────────────────────────────────────────────────────────────────────
//
//	[u8 TypeSnapHead][SnapHead]  [u8 loại][payload]  [u8 loại][payload]  …
//
// Mỗi message tự biết độ dài của mình (generator trả về số byte đã đọc), nên không cần
// trường độ dài. Hết buffer là hết gói.

// keyframeEverySecs — §4.3: cứ ngần này giây gửi một gói TOÀN BỘ trạng thái.
//
// Vì sao cần dù đã có delta: delta chỉ đúng nếu client nhận được mọi gói trước đó. WebSocket
// trên TCP thì không mất gói, NHƯNG `SendLossy` **cố ý vứt** snapshot cũ khi hàng đợi đầy
// (mạng yếu) — và đó chính là lúc client cần đúng nhất. Không có keyframe thì một lần vứt
// gói là người chơi đó kẹt với trạng thái sai cho tới khi entity đổi hết mọi trường.
const keyframeEverySecs = 2.0

// wirePlayer là trạng thái một người SAU KHI lượng tử hoá.
//
// So sánh ở dạng đã lượng tử, không so ở dạng float gốc: nếu so float thì vị trí nhúc nhích
// 0,001 đơn vị cũng tính là "đã đổi" và delta không tiết kiệm được gì — trong khi hai giá
// trị đó mã hoá ra cùng một con số u16.
type wirePlayer struct {
	X, Y      uint16
	Facing    uint8
	Level     uint8
	Scale     uint8
	Swords    uint16
	RingPhase uint16
	RingDir   int8
	Mood      uint8
	Buffs     uint16
	Flags     uint16
	Tilt      int8
	Status    uint16
	Crack     uint8
}

// quantPlayer lượng tử hoá theo §4.6.
func quantPlayer(p netproto.PlayerState) wirePlayer {
	return wirePlayer{
		X: netproto.QuantPos(float64(p.X)), Y: netproto.QuantPos(float64(p.Y)),
		Facing: quantAngle(float64(p.Facing)),
		Level:  clampU8(p.Level),
		// Scale bước 0,02: cỡ nhân vật chạy từ 1 tới ~5, nên 0,02 là 1/250 của khoảng đó —
		// nhỏ hơn một điểm ảnh ở mọi mức phóng to.
		Scale:     clampU8(int(math.Round(float64(p.Scale) / 0.02))),
		Swords:    clampU16(p.Count),
		RingPhase: quantAngle16(float64(p.Phase)),
		RingDir:   p.Dir,
		Mood:      p.Mood,
		Buffs:     uint16(p.Buffs),
		Flags:     uint16(p.Flags),
		// Tilt |x| ≤ 35° = 0,61 rad; bước 1/128 rad cho sai số tối đa 0,2° — mắt không thấy.
		Tilt:   int8(clampF(float64(p.Tilt)*128, -127, 127)),
		Status: p.St,
		Crack:  uint8(clampF(float64(p.Crack)*255, 0, 255)),
	}
}

// Bit của `changed` trong message Update — THỨ TỰ phải khớp dòng bitmask trong schema.
const (
	chX = 1 << iota
	chY
	chFacing
	chLevel
	chScale
	chSwords
	chRingPhase
	chRingDir
	chMood
	chBuffs
	chFlags
	chTilt
	chStatus
	chCrack
)

// diffPlayer trả về bitmask những trường đã đổi.
func diffPlayer(old, now wirePlayer) uint16 {
	var m uint16
	set := func(cond bool, bit uint16) {
		if cond {
			m |= bit
		}
	}
	set(old.X != now.X, chX)
	set(old.Y != now.Y, chY)
	set(old.Facing != now.Facing, chFacing)
	set(old.Level != now.Level, chLevel)
	set(old.Scale != now.Scale, chScale)
	set(old.Swords != now.Swords, chSwords)
	set(old.RingPhase != now.RingPhase, chRingPhase)
	set(old.RingDir != now.RingDir, chRingDir)
	set(old.Mood != now.Mood, chMood)
	set(old.Buffs != now.Buffs, chBuffs)
	set(old.Flags != now.Flags, chFlags)
	set(old.Tilt != now.Tilt, chTilt)
	set(old.Status != now.Status, chStatus)
	set(old.Crack != now.Crack, chCrack)
	return m
}

// EncodeSnapshot biến một snapshot thành gói nhị phân cho MỘT client.
//
// `c` giữ trạng thái đã gửi lần trước, nên hàm này có tác dụng phụ — gọi đúng một lần cho
// mỗi client mỗi tick.
func (c *Client) EncodeSnapshot(s *netproto.Snapshot, chars []string) []byte {
	keyframe := c.forceKeyframe || c.lastWire == nil || s.Time-c.keyframeAt >= keyframeEverySecs
	if keyframe {
		c.keyframeAt, c.forceKeyframe = s.Time, false
		// Quên hết: gói này chở lại toàn bộ, nên mọi so sánh delta phải bắt đầu từ con số 0.
		c.lastWire = make(map[uint32]wirePlayer, len(s.Players))
	}

	buf := make([]byte, 0, 1024)
	var flags uint8
	if keyframe {
		flags |= 1
	}
	buf = append(buf, wire.TypeSnapHead)
	buf = wire.SnapHead{
		Tick: s.Tick, Ack: s.Ack,
		TimeMs: uint32(s.Time * 1000),
		SrvUs:  clampU16(int(s.SrvMs * 1000)),
		Flags:  flags,
	}.Encode(buf)

	for _, p := range s.Players {
		now := quantPlayer(p)
		old, known := c.lastWire[p.ID]
		if !known {
			// Spawn chở ĐỦ mọi thứ, kể cả descriptor — client dựng được ngay, không hỏi lại.
			buf = append(buf, wire.TypeSpawn)
			buf = wire.Spawn{
				ID: p.ID, DescVer: 1, Name: p.Name, Char: charIndex(chars, p.Char),
				Title: clampU8(p.Title),
				X:     now.X, Y: now.Y, Facing: now.Facing, Level: now.Level, Scale: now.Scale,
				Swords: now.Swords, RingPhase: now.RingPhase, RingDir: now.RingDir,
				Mood:   now.Mood, Buffs: now.Buffs, Flags: now.Flags,
				Tilt:   now.Tilt, Status: now.Status, Crack: now.Crack,
			}.Encode(buf)
			c.lastWire[p.ID] = now
			continue
		}
		// Tên đến muộn (server gửi lại mỗi 2 giây cho client lỡ gói đầu) thì đi bằng Desc,
		// KHÔNG bằng Spawn: Spawn nghĩa là "dựng lại từ đầu" và client sẽ vứt view đang có
		// — đúng cái làm nhân vật nhấp nháy mà bước 3 vừa sửa xong.
		if p.Name != "" {
			buf = append(buf, wire.TypeDesc)
			buf = wire.Desc{ID: p.ID, DescVer: 1, Name: p.Name,
				Char: charIndex(chars, p.Char), Title: clampU8(p.Title)}.Encode(buf)
		}
		changed := diffPlayer(old, now)
		if changed == 0 {
			continue // đứng yên, không đổi gì: không tốn một byte nào
		}
		buf = append(buf, wire.TypeUpdate)
		buf = wire.Update{
			ID: p.ID, Changed: changed,
			X: now.X, Y: now.Y, Facing: now.Facing, Level: now.Level, Scale: now.Scale,
			Swords: now.Swords, RingPhase: now.RingPhase, RingDir: now.RingDir,
			Mood:   now.Mood, Buffs: now.Buffs, Flags: now.Flags,
			Tilt:   now.Tilt, Status: now.Status, Crack: now.Crack,
		}.Encode(buf)
		c.lastWire[p.ID] = now
	}

	// Phần CHỈ chủ nhân cần. Gửi khi đổi hoặc ở keyframe.
	buf = c.encodeSelf(buf, s, keyframe)

	for _, m := range s.Monsters {
		var has uint8
		if m.HP > 0 {
			has |= 1
		}
		if m.Kind != "" {
			has |= 2 | 4 | 8 // kind, max_hp, r đi cùng nhau: cả ba chỉ gửi lần đầu
		}
		buf = append(buf, wire.TypeMonster)
		buf = wire.Monster{
			ID: m.ID, X: m.X, Y: m.Y, Has: has,
			Hp:    clampU16(m.HP),
			Kind:  monsterKindIndex(m.Kind),
			MaxHp: clampU16(m.Max),
			R:     clampU16(int(m.R)),
		}.Encode(buf)
	}
	for _, it := range s.ItemsAdded {
		var has uint8
		if it.Value != 0 {
			has |= 1
		}
		buf = append(buf, wire.TypeItemAdd)
		buf = wire.ItemAdd{ID: it.ID, X: it.X, Y: it.Y, Kind: it.Kind, Has: has, Value: it.Value}.Encode(buf)
	}
	for _, id := range s.ItemsRemoved {
		buf = append(buf, wire.TypeItemDel)
		buf = wire.ItemDel{ID: id}.Encode(buf)
	}
	for _, id := range s.Despawn {
		buf = append(buf, wire.TypeDespawn)
		buf = wire.Despawn{ID: id, Reason: 0}.Encode(buf)
		// Quên trạng thái delta của entity đã rời: nếu nó quay lại, phải Spawn lại từ đầu
		// chứ không gửi Update so với một bản client đã vứt đi.
		delete(c.lastWire, id)
	}
	for _, ev := range s.Events {
		buf = append(buf, wire.TypeEvent)
		buf = encodeEvent(buf, ev)
	}
	return buf
}

// encodeSelf gửi phần chỉ chủ nhân cần (§4.6).
func (c *Client) encodeSelf(buf []byte, s *netproto.Snapshot, keyframe bool) []byte {
	var me *netproto.PlayerState
	for i := range s.Players {
		if s.Players[i].ID == c.PlayerID {
			me = &s.Players[i]
			break
		}
	}
	if me == nil {
		return buf
	}
	// SelfState đi MỌI tick: client dùng nó để hiệu chỉnh dự đoán (GAME_SPEC §8.2a), và
	// vị trí ở đây giữ 1/256 đơn vị chứ không lượng tử về u16 như người khác — sai số lượng
	// tử trên vị trí của chính mình biến thành giật.
	buf = append(buf, wire.TypeSelfState)
	buf = wire.SelfState{
		AckSeq: s.Ack,
		X:      fixed256(float64(me.X)), Y: fixed256(float64(me.Y)),
		Vx: fixed256(float64(me.Vx)), Vy: fixed256(float64(me.Vy)),
	}.Encode(buf)

	ex := wire.SelfExtra{
		Exp: me.Exp, Charges: clampU8(me.Charges), Title: clampU8(me.Title),
		TienHave: me.TienHave, TienNeed: me.TienNeed,
	}
	if keyframe || ex != c.lastSelfExtra {
		ex.Has = 1 | 2 | 4 | 8 | 16
		buf = append(buf, wire.TypeSelfExtra)
		buf = ex.Encode(buf)
		c.lastSelfExtra = ex
	}
	return buf
}

// encodeEvent — §4.1. `has` quyết trường nào có mặt; sự kiện thường chỉ dùng 2–3 trường.
func encodeEvent(buf []byte, ev game.Event) []byte {
	var has uint8
	if ev.A != 0 {
		has |= 1
	}
	if ev.B != 0 {
		has |= 2
	}
	if ev.N != 0 {
		has |= 4
	}
	if ev.S != "" {
		has |= 8
	}
	if ev.F != 0 {
		has |= 16
	}
	return wire.Event{
		Kind: netproto.EventKindIndex(ev.Kind), Has: has,
		X: netproto.QuantPos(float64(ev.X)), Y: netproto.QuantPos(float64(ev.Y)),
		A: ev.A, B: ev.B, N: int32(ev.N), S: ev.S, F: ev.F,
	}.Encode(buf)
}

// ── Lượng tử hoá (§4.6) ─────────────────────────────────────────────────────────────────

// quantAngle: góc 0..2π về 0..255.
func quantAngle(rad float64) uint8 {
	t := math.Mod(rad/(2*math.Pi), 1)
	if t < 0 {
		t++
	}
	return uint8(t * 256)
}

// quantAngle16: góc 0..2π về 0..65535 — dùng cho pha vòng kiếm.
//
// Cần 16 bit chứ không 8 như hướng nhìn: pha vòng kiếm quay rất nhanh và client NỘI SUY
// giữa hai snapshot. Sai số 1/256 vòng ở 20 Hz làm vòng kiếm giật thấy rõ; 1/65536 thì không.
func quantAngle16(rad float64) uint16 {
	t := math.Mod(rad/(2*math.Pi), 1)
	if t < 0 {
		t++
	}
	return uint16(t * 65536)
}

// fixed256: số thực về số nguyên 1/256 đơn vị.
func fixed256(v float64) int32 {
	return int32(clampF(v*256, math.MinInt32, math.MaxInt32))
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampU8(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func clampU16(v int) uint16 {
	if v < 0 {
		return 0
	}
	if v > 65535 {
		return 65535
	}
	return uint16(v)
}

// charIndex tra chỉ số nhân vật trong `config.characters` — gửi 1 byte thay cho chuỗi.
func charIndex(chars []string, name string) uint8 {
	for i, c := range chars {
		if c == name {
			return uint8(i)
		}
	}
	return 255
}

// monsterKindIndex tra chỉ số loại quái.
func monsterKindIndex(kind string) uint8 {
	for i, k := range netproto.MonsterKinds {
		if k == kind {
			return uint8(i)
		}
	}
	return 255
}
