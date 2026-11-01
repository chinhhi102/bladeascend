// Command protofixture sinh bộ gói mẫu để client giải mã và đối chiếu — NETCODE_SPEC §4.9.
//
//	cd server && go run ./cmd/protofixture -o /tmp/wire.json
//	cd client && npm run check:wire -- /tmp/wire.json
//
// Cùng một ý với `cmd/collisionfixture` (GAME_SPEC §8.2a): hai phía cài cùng một thuật toán
// thì phải có một bài ép chúng chứng minh là giống nhau, bằng số, chạy được trong CI.
//
// Hai loại gói, và cần CẢ HAI:
//   - **Golden**: giá trị chọn tay, nhắm vào biên (0, giá trị lớn nhất, chuỗi rỗng, chuỗi
//     nhiều byte UTF-8, bitmask rỗng, bitmask đầy). Đây là những chỗ sai hay nấp.
//   - **Fuzz 10 000 gói**: giá trị ngẫu nhiên có seed cố định. Golden chỉ bắt được lỗi ta
//     NGHĨ RA; fuzz bắt được lỗi ta không nghĩ tới.
package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"

	"bladeascend/server/internal/wire"
)

// Case là một gói mẫu: giá trị từng trường + chuỗi byte server mã hoá ra.
type Case struct {
	Msg    string         `json:"msg"`
	Fields map[string]any `json:"fields"`
	Bytes  string         `json:"bytes"` // base64
}

// Fixture là toàn bộ bộ mẫu.
type Fixture struct {
	SchemaHash    string `json:"schema_hash"`
	SchemaVersion int    `json:"schema_version"`
	Cases         []Case `json:"cases"`
}

func main() {
	out := flag.String("o", "wire_fixture.json", "file kết quả")
	fuzz := flag.Int("fuzz", 10000, "số gói ngẫu nhiên (spec §4.9 yêu cầu 10 000)")
	seed := flag.Int64("seed", 7, "seed cố định — bài kiểm phải lặp lại được")
	flag.Parse()

	f := Fixture{SchemaHash: wire.SchemaHash, SchemaVersion: wire.SchemaVersion}
	f.Cases = append(f.Cases, goldenCases()...)
	f.Cases = append(f.Cases, fuzzCases(rand.New(rand.NewSource(*seed)), *fuzz)...)

	b, err := json.Marshal(f)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*out, b, 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("schema %s v%d · %d gói mẫu → %s (%.0f KB)\n",
		f.SchemaHash, f.SchemaVersion, len(f.Cases), *out, float64(len(b))/1024)
}

// goldenCases: giá trị chọn tay, nhắm thẳng vào biên.
func goldenCases() []Case {
	var cs []Case
	add := func(name string, fields map[string]any, bytes []byte) {
		cs = append(cs, Case{Msg: name, Fields: fields, Bytes: base64.StdEncoding.EncodeToString(bytes)})
	}

	// Spawn: toàn 0, rồi toàn giá trị lớn nhất, rồi tên nhiều byte UTF-8.
	for _, sp := range []wire.Spawn{
		{},
		{ID: 0xFFFFFFFF, DescVer: 0xFFFF, Name: "", Char: 255, Title: 255, X: 65535, Y: 65535,
			Facing: 255, Level: 255, Scale: 255, Swords: 65535, RingPhase: 65535, RingDir: -128,
			Mood: 255, Buffs: 65535, Flags: 65535, Tilt: -128, Status: 65535, Crack: 255},
		// Tên tiếng Việt có dấu: mỗi chữ tốn 2–3 byte UTF-8. Đây là chỗ bộ đếm độ dài hay sai
		// vì người ta đếm KÝ TỰ thay vì BYTE — và lỗi đó chỉ lộ ra với tên tiếng Việt.
		{ID: 42, Name: "Đệ Tử Hay Khóc", Char: 3, Level: 9, Swords: 400},
		{ID: 7, Name: "A", RingDir: 1, Scale: 50},
	} {
		add("Spawn", map[string]any{
			"id": sp.ID, "descVer": sp.DescVer, "name": sp.Name, "char": sp.Char, "title": sp.Title,
			"x": sp.X, "y": sp.Y, "facing": sp.Facing, "level": sp.Level, "scale": sp.Scale,
			"swords": sp.Swords, "ringPhase": sp.RingPhase, "ringDir": sp.RingDir,
			"mood": sp.Mood, "buffs": sp.Buffs, "flags": sp.Flags,
			"tilt": sp.Tilt, "status": sp.Status, "crack": sp.Crack,
		}, sp.Encode(nil))
	}

	// Update: bitmask RỖNG (không trường nào đi theo) và bitmask ĐẦY.
	// Bitmask rỗng là ca quan trọng nhất của cả giao thức — entity đứng yên thì gói chỉ còn
	// 6 byte, và đó là chỗ tiết kiệm băng thông lớn nhất. Đọc sai nó là hỏng mọi thứ.
	empty := wire.Update{ID: 1, Changed: 0}
	add("Update", updateFields(empty), empty.Encode(nil))

	// 0x3FFF: ĐỦ 14 bit. Để 0x7FF như cũ thì ba bit mới (tilt/status/crack) không bao giờ
	// được ca "bitmask đầy" chạm tới — mà ca đó sinh ra chính là để chạm hết.
	full := wire.Update{ID: 2, Changed: 0x3FFF, X: 1, Y: 2, Facing: 3, Level: 4, Scale: 5,
		Swords: 6, RingPhase: 7, RingDir: -8, Mood: 9, Buffs: 10, Flags: 11,
		Tilt: -12, Status: 13, Crack: 14}
	add("Update", updateFields(full), full.Encode(nil))

	// Chỉ một bit — ca hay gặp nhất khi chạy thật (chỉ vị trí đổi).
	onlyXY := wire.Update{ID: 3, Changed: 0b11, X: 1234, Y: 5678}
	add("Update", updateFields(onlyXY), onlyXY.Encode(nil))

	for _, d := range []wire.Desc{{}, {ID: 9, DescVer: 3, Name: "Lão Quái Ham Ngủ85", Char: 7, Title: 2}} {
		add("Desc", map[string]any{"id": d.ID, "descVer": d.DescVer, "name": d.Name, "char": d.Char, "title": d.Title}, d.Encode(nil))
	}
	for _, d := range []wire.Despawn{{}, {ID: 123, Reason: 3}} {
		add("Despawn", map[string]any{"id": d.ID, "reason": d.Reason}, d.Encode(nil))
	}
	// SelfState: số ÂM là chỗ i32 hay bị đọc thành u32.
	for _, s := range []wire.SelfState{{}, {AckSeq: 99, X: -123456, Y: 123456, Vx: -1, Vy: 1}} {
		add("SelfState", map[string]any{"ackSeq": s.AckSeq, "x": s.X, "y": s.Y, "vx": s.Vx, "vy": s.Vy}, s.Encode(nil))
	}
	for _, i := range []wire.Input{{}, {Seq: 4294967295, Angle: 128, Buttons: 7}} {
		add("Input", map[string]any{"seq": i.Seq, "angle": i.Angle, "buttons": i.Buttons}, i.Encode(nil))
	}

	// Event: bitmask rỗng (ca hay gặp nhất — sự kiện chỉ có toạ độ), bitmask đầy, và chuỗi
	// tiếng Việt có dấu trong trường `s`.
	for _, ev := range []wire.Event{
		{Kind: 10, Has: 0, X: 100, Y: 200},
		{Kind: 255, Has: 0x1F, X: 65535, Y: 65535, A: 0xFFFFFFFF, B: 1, N: -2147483648,
			S: "Đệ Tử Hay Khóc", F: -1.5},
		{Kind: 0, Has: 8, S: ""},
	} {
		add("Event", maskFields(
			map[string]any{"kind": ev.Kind, "has": ev.Has, "x": ev.X, "y": ev.Y},
			ev.Has, []optField{{"a", ev.A}, {"b", ev.B}, {"n", ev.N}, {"s", ev.S}, {"f", f32(ev.F)}},
		), ev.Encode(nil))
	}
	for _, ia := range []wire.ItemAdd{
		{}, {ID: 7, X: 1, Y: 2, Kind: 12, Has: 1, Value: 500},
	} {
		add("ItemAdd", maskFields(
			map[string]any{"id": ia.ID, "x": ia.X, "y": ia.Y, "kind": ia.Kind, "has": ia.Has},
			ia.Has, []optField{{"value", f32(ia.Value)}},
		), ia.Encode(nil))
	}
	for _, d := range []wire.ItemDel{{}, {ID: 4294967295}} {
		add("ItemDel", map[string]any{"id": d.ID}, d.Encode(nil))
	}
	for _, m := range []wire.Monster{
		{ID: 1, X: 2, Y: 3, Has: 0},            // quái đứng yên: 9 byte, ca hay gặp nhất
		{ID: 2, X: 4, Y: 5, Has: 1, Hp: 7},     // chỉ máu đổi
		{ID: 3, X: 6, Y: 7, Has: 0xF, Hp: 100, Kind: 3, MaxHp: 100, R: 60}, // lần đầu thấy
	} {
		add("Monster", maskFields(
			map[string]any{"id": m.ID, "x": m.X, "y": m.Y, "has": m.Has},
			m.Has, []optField{{"hp", m.Hp}, {"kind", m.Kind}, {"maxHp", m.MaxHp}, {"r", m.R}},
		), m.Encode(nil))
	}
	for _, d := range []wire.MonsterDel{{}, {ID: 99}} {
		add("MonsterDel", map[string]any{"id": d.ID}, d.Encode(nil))
	}
	for _, h := range []wire.SnapHead{
		{}, {Tick: 4294967295, Ack: 4294967295, TimeMs: 4294967295, SrvUs: 65535, Flags: 1},
	} {
		add("SnapHead", map[string]any{"tick": h.Tick, "ack": h.Ack, "timeMs": h.TimeMs,
			"srvUs": h.SrvUs, "flags": h.Flags}, h.Encode(nil))
	}
	for _, e := range []wire.SelfExtra{
		{}, {Has: 0x1F, Exp: 694960, Charges: 3, Title: 1, TienHave: 15000, TienNeed: 15000},
	} {
		add("SelfExtra", maskFields(
			map[string]any{"has": e.Has},
			e.Has, []optField{{"exp", f32(e.Exp)}, {"charges", e.Charges}, {"title", e.Title},
				{"tienHave", f32(e.TienHave)}, {"tienNeed", f32(e.TienNeed)}},
		), e.Encode(nil))
	}
	// Bảng xếp hạng phòng (spec §11.5 v6.19): hạng 0 (chưa có), số lớn nhất, tên Thái nhiều
	// byte UTF-8, và ca Ẩn Thân (khu 255).
	for _, hd := range []wire.LbHead{{}, {Rank: 65535, Total: 65535, MyScore: 4294967295, Count: 10}} {
		add("LbHead", lbHeadFields(hd), hd.Encode(nil))
	}
	for _, row := range []wire.LbRow{
		{},
		{ID: 0xFFFFFFFF, Name: "Mít Tơ Bean", Character: 255, TitleID: 1, Rank: 1, ZoneID: 4,
			Realm: 13, Tier: 10, Score: 4294967295, Stealth: 0},
		{ID: 5, Name: "สมชาย", Character: 2, TitleID: 3, Rank: 3, ZoneID: 255, Realm: 2, Tier: 9,
			Score: 1070, Stealth: 1},
	} {
		add("LbRow", lbRowFields(row), row.Encode(nil))
	}
	return cs
}

func lbHeadFields(h wire.LbHead) map[string]any {
	return map[string]any{"rank": h.Rank, "total": h.Total, "myScore": h.MyScore, "count": h.Count}
}

func lbRowFields(v wire.LbRow) map[string]any {
	return map[string]any{"id": v.ID, "name": v.Name, "character": v.Character, "titleId": v.TitleID,
		"rank": v.Rank, "zoneId": v.ZoneID, "realm": v.Realm, "tier": v.Tier, "score": v.Score,
		"stealth": v.Stealth}
}

func updateFields(u wire.Update) map[string]any {
	m := map[string]any{"id": u.ID, "changed": u.Changed}
	// CHỈ đưa vào những trường bit đang bật — đúng như client sẽ giải mã. Đưa cả trường không
	// gửi vào thì bài đối chiếu so với một giá trị không bao giờ đi trên dây.
	type bit struct {
		name string
		i    int
		val  any
	}
	for _, b := range []bit{
		{"x", 0, u.X}, {"y", 1, u.Y}, {"facing", 2, u.Facing}, {"level", 3, u.Level},
		{"scale", 4, u.Scale}, {"swords", 5, u.Swords}, {"ringPhase", 6, u.RingPhase},
		{"ringDir", 7, u.RingDir}, {"mood", 8, u.Mood}, {"buffs", 9, u.Buffs}, {"flags", 10, u.Flags},
		{"tilt", 11, u.Tilt}, {"status", 12, u.Status}, {"crack", 13, u.Crack},
	} {
		if u.Changed&(1<<b.i) != 0 {
			m[b.name] = b.val
		}
	}
	return m
}

// fuzzCases sinh gói ngẫu nhiên. Seed cố định nên hỏng là lặp lại được.
func fuzzCases(r *rand.Rand, n int) []Case {
	names := []string{"", "A", "Đệ Tử", "Lão Quái Ham Ngủ85", "ユーザー", "a.b_c", "Bé Tí", "~Kikix"}
	var cs []Case
	for i := 0; i < n; i++ {
		switch i % 14 { // 14 = 13 ca có tên + default (Input) — để 13 thì default không bao giờ tới
		case 0:
			sp := wire.Spawn{
				ID: r.Uint32(), DescVer: uint16(r.Intn(65536)), Name: names[r.Intn(len(names))],
				Char: uint8(r.Intn(256)), Title: uint8(r.Intn(256)),
				X: uint16(r.Intn(65536)), Y: uint16(r.Intn(65536)), Facing: uint8(r.Intn(256)),
				Level: uint8(r.Intn(256)), Scale: uint8(r.Intn(256)), Swords: uint16(r.Intn(65536)),
				RingPhase: uint16(r.Intn(65536)), RingDir: int8(r.Intn(256) - 128),
				Mood: uint8(r.Intn(256)), Buffs: uint16(r.Intn(65536)), Flags: uint16(r.Intn(65536)),
				Tilt: int8(r.Intn(256) - 128), Status: uint16(r.Intn(65536)), Crack: uint8(r.Intn(256)),
			}
			cs = append(cs, Case{Msg: "Spawn", Fields: map[string]any{
				"id": sp.ID, "descVer": sp.DescVer, "name": sp.Name, "char": sp.Char, "title": sp.Title,
				"x": sp.X, "y": sp.Y, "facing": sp.Facing, "level": sp.Level, "scale": sp.Scale,
				"swords": sp.Swords, "ringPhase": sp.RingPhase, "ringDir": sp.RingDir,
				"mood": sp.Mood, "buffs": sp.Buffs, "flags": sp.Flags,
				"tilt": sp.Tilt, "status": sp.Status, "crack": sp.Crack,
			}, Bytes: b64(sp.Encode(nil))})
		case 1:
			u := wire.Update{
				// 1<<14: bitmask nay có 14 bit (thêm tilt, status, crack). Để 1<<11 như cũ thì
				// ba bit mới KHÔNG BAO GIỜ được fuzz chạm tới — và một bài fuzz không chạm
				// tới nhánh mới thì xanh một cách vô nghĩa.
				ID: r.Uint32(), Changed: uint16(r.Intn(1 << 14)),
				X: uint16(r.Intn(65536)), Y: uint16(r.Intn(65536)), Facing: uint8(r.Intn(256)),
				Level: uint8(r.Intn(256)), Scale: uint8(r.Intn(256)), Swords: uint16(r.Intn(65536)),
				RingPhase: uint16(r.Intn(65536)), RingDir: int8(r.Intn(256) - 128),
				Mood: uint8(r.Intn(256)), Buffs: uint16(r.Intn(65536)), Flags: uint16(r.Intn(65536)),
				Tilt: int8(r.Intn(256) - 128), Status: uint16(r.Intn(65536)), Crack: uint8(r.Intn(256)),
			}
			cs = append(cs, Case{Msg: "Update", Fields: updateFields(u), Bytes: b64(u.Encode(nil))})
		case 2:
			d := wire.Desc{ID: r.Uint32(), DescVer: uint16(r.Intn(65536)),
				Name: names[r.Intn(len(names))], Char: uint8(r.Intn(256)), Title: uint8(r.Intn(256))}
			cs = append(cs, Case{Msg: "Desc", Fields: map[string]any{
				"id": d.ID, "descVer": d.DescVer, "name": d.Name, "char": d.Char, "title": d.Title,
			}, Bytes: b64(d.Encode(nil))})
		case 3:
			d := wire.Despawn{ID: r.Uint32(), Reason: uint8(r.Intn(4))}
			cs = append(cs, Case{Msg: "Despawn", Fields: map[string]any{"id": d.ID, "reason": d.Reason}, Bytes: b64(d.Encode(nil))})
		case 4:
			s := wire.SelfState{AckSeq: r.Uint32(),
				X: int32(r.Int63n(1<<32) - (1 << 31)), Y: int32(r.Int63n(1<<32) - (1 << 31)),
				Vx: int32(r.Int63n(1<<32) - (1 << 31)), Vy: int32(r.Int63n(1<<32) - (1 << 31))}
			cs = append(cs, Case{Msg: "SelfState", Fields: map[string]any{
				"ackSeq": s.AckSeq, "x": s.X, "y": s.Y, "vx": s.Vx, "vy": s.Vy}, Bytes: b64(s.Encode(nil))})
		case 5:
			// Event là message NẶNG NHẤT của gói (37% ở bản v1) và cũng là message duy nhất
			// có cả bitmask LẪN chuỗi — hai thứ hay sai nhất, nên fuzz nó kỹ.
			ev := wire.Event{
				Kind: uint8(r.Intn(256)), Has: uint8(r.Intn(32)),
				X: uint16(r.Intn(65536)), Y: uint16(r.Intn(65536)),
				A: r.Uint32(), B: r.Uint32(), N: int32(r.Int63n(1<<32) - (1 << 31)),
				S: names[r.Intn(len(names))], F: float32(r.NormFloat64()),
			}
			cs = append(cs, Case{Msg: "Event", Fields: maskFields(
				map[string]any{"kind": ev.Kind, "has": ev.Has, "x": ev.X, "y": ev.Y},
				ev.Has, []optField{{"a", ev.A}, {"b", ev.B}, {"n", ev.N}, {"s", ev.S}, {"f", f32(ev.F)}},
			), Bytes: b64(ev.Encode(nil))})
		case 6:
			ia := wire.ItemAdd{ID: r.Uint32(), X: uint16(r.Intn(65536)), Y: uint16(r.Intn(65536)),
				Kind: uint8(r.Intn(256)), Has: uint8(r.Intn(2)), Value: float32(r.NormFloat64() * 1000)}
			cs = append(cs, Case{Msg: "ItemAdd", Fields: maskFields(
				map[string]any{"id": ia.ID, "x": ia.X, "y": ia.Y, "kind": ia.Kind, "has": ia.Has},
				ia.Has, []optField{{"value", f32(ia.Value)}},
			), Bytes: b64(ia.Encode(nil))})
		case 7:
			d := wire.ItemDel{ID: r.Uint32()}
			cs = append(cs, Case{Msg: "ItemDel", Fields: map[string]any{"id": d.ID}, Bytes: b64(d.Encode(nil))})
		case 8:
			m := wire.Monster{ID: r.Uint32(), X: uint16(r.Intn(65536)), Y: uint16(r.Intn(65536)),
				Has: uint8(r.Intn(16)), Hp: uint16(r.Intn(65536)), Kind: uint8(r.Intn(256)),
				MaxHp: uint16(r.Intn(65536)), R: uint16(r.Intn(65536))}
			cs = append(cs, Case{Msg: "Monster", Fields: maskFields(
				map[string]any{"id": m.ID, "x": m.X, "y": m.Y, "has": m.Has},
				m.Has, []optField{{"hp", m.Hp}, {"kind", m.Kind}, {"maxHp", m.MaxHp}, {"r", m.R}},
			), Bytes: b64(m.Encode(nil))})
		case 9:
			h := wire.SnapHead{Tick: r.Uint32(), Ack: r.Uint32(), TimeMs: r.Uint32(),
				SrvUs: uint16(r.Intn(65536)), Flags: uint8(r.Intn(256))}
			cs = append(cs, Case{Msg: "SnapHead", Fields: map[string]any{
				"tick": h.Tick, "ack": h.Ack, "timeMs": h.TimeMs, "srvUs": h.SrvUs, "flags": h.Flags,
			}, Bytes: b64(h.Encode(nil))})
		case 10:
			e := wire.SelfExtra{Has: uint8(r.Intn(32)), Exp: float32(r.NormFloat64() * 1e6),
				Charges: uint8(r.Intn(256)), Title: uint8(r.Intn(256)),
				TienHave: float32(r.NormFloat64() * 1e4), TienNeed: float32(r.NormFloat64() * 1e4)}
			cs = append(cs, Case{Msg: "SelfExtra", Fields: maskFields(
				map[string]any{"has": e.Has},
				e.Has, []optField{{"exp", f32(e.Exp)}, {"charges", e.Charges}, {"title", e.Title},
					{"tienHave", f32(e.TienHave)}, {"tienNeed", f32(e.TienNeed)}},
			), Bytes: b64(e.Encode(nil))})
		case 11:
			hd := wire.LbHead{Rank: uint16(r.Intn(65536)), Total: uint16(r.Intn(65536)),
				MyScore: r.Uint32(), Count: uint8(r.Intn(256))}
			cs = append(cs, Case{Msg: "LbHead", Fields: lbHeadFields(hd), Bytes: b64(hd.Encode(nil))})
		case 12:
			row := wire.LbRow{ID: r.Uint32(), Name: names[r.Intn(len(names))],
				Character: uint8(r.Intn(256)), TitleID: uint8(r.Intn(256)), Rank: uint8(r.Intn(256)),
				ZoneID: uint8(r.Intn(256)), Realm: uint8(r.Intn(256)), Tier: uint8(r.Intn(256)),
				Score: r.Uint32(), Stealth: uint8(r.Intn(2))}
			cs = append(cs, Case{Msg: "LbRow", Fields: lbRowFields(row), Bytes: b64(row.Encode(nil))})
		default:
			in := wire.Input{Seq: r.Uint32(), Angle: uint8(r.Intn(256)), Buttons: uint8(r.Intn(8))}
			cs = append(cs, Case{Msg: "Input", Fields: map[string]any{
				"seq": in.Seq, "angle": in.Angle, "buttons": in.Buttons}, Bytes: b64(in.Encode(nil))})
		}
	}
	return cs
}

// maskFields gom trường của một message có bitmask: CHỈ những trường bit đang bật.
//
// Đưa cả trường bit tắt vào là so với một giá trị KHÔNG BAO GIỜ đi trên dây — bài đối chiếu
// sẽ đỏ dù hai phía hoàn toàn khớp nhau.
func maskFields(base map[string]any, mask uint8, opt []optField) map[string]any {
	for i, f := range opt {
		if mask&(1<<i) != 0 {
			base[f.name] = f.val
		}
	}
	return base
}

type optField struct {
	name string
	val  any
}

// f32 ghi một float32 ra JSON dưới dạng **giá trị float64 chính xác của nó**.
//
// Không có bước này thì `encoding/json` in float32 bằng cách ngắn nhất đủ để đọc lại ra
// CHÍNH float32 đó — `640940.94`. JavaScript đọc chuỗi ấy thành float64 `640940.94`, còn bộ
// giải mã của nó đọc 4 byte trên dây ra `640940.9375`. Hai con số khác nhau, mà hai phía
// hoàn toàn khớp: khác biệt nằm ở CHỮ SỐ trong JSON, không ở byte trên dây.
//
// Đổi bài đối chiếu sang "so xấp xỉ" là sai hướng — nó sẽ bỏ qua cả những lệch thật.
func f32(v float32) float64 { return float64(v) }

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }
