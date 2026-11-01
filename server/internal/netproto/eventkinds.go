package netproto

// Bảng loại sự kiện — Protocol v2 (NETCODE_SPEC §4.1).
//
// Vì sao cần bảng: ở msgpack v1 mỗi sự kiện chở `kind` dạng CHUỖI. "TienKhiPicked" là 14
// byte, cộng tên khoá, cộng khung map — đo bằng `cmd/loadtest` thì phần `ev` chiếm **27%
// gói lớn nhất** với 43 byte mỗi sự kiện. Thay chuỗi bằng một byte chỉ số là chỗ tiết kiệm
// lớn nhất mà không mất gì.
//
// ⚠ **THỨ TỰ LÀ GIAO THỨC.** Client tra theo chỉ số (`client/src/net/eventkinds.ts`).
// Thêm loại mới thì thêm vào CUỐI; chèn vào giữa là mọi sự kiện sau đó dịch đi một bậc và
// client vẽ nhầm hiệu ứng mà không có lỗi nào được ném ra.
var EventKinds = []string{
	"AltarBurst",
	"AltarWarn",
	"Blinded",
	"Breakthrough",
	"BuffEnd",
	"BuffStart",
	"ChestIncoming",
	"Death",
	"Emote",
	"GateWarn",
	"Hit",
	"ItemExpired",
	"ItemPicked",
	"KillFeed",
	"LevelUp",
	"MonsterDie",
	"MonsterHurt",
	"ObstacleBroken",
	"ObstacleRegrow",
	"ObstacleTooHard",
	"RealmDrop",
	"RingFlip",
	"Score",
	"SkillCast",
	"SkillCharge",
	"SkillEffect",
	"StallDown",
	"StallUp",
	"SwordBreak",
	"TeleportIn",
	"TeleportOut",
	"TienKhiPicked",
	"BotLeave",    // bot rời phòng do admin rút (§9.5): client vẽ khói
	"StatusStart", // v6.36 §6.5: trạng thái bắt đầu {A bị, B gây, N mã, F giây}
	"StatusEnd",   // v6.36 §6.5: trạng thái hết
}

// eventIndex tra ngược, dựng một lần lúc khởi động.
var eventIndex = func() map[string]uint8 {
	m := make(map[string]uint8, len(EventKinds))
	for i, k := range EventKinds {
		m[k] = uint8(i)
	}
	return m
}()

// EventKindUnknown là chỉ số trả về khi gặp loại chưa có trong bảng.
//
// 255 chứ không phải 0: 0 là `AltarBurst`, một sự kiện CÓ THẬT. Nếu một loại mới lọt ra
// dây mà quên thêm vào bảng, client sẽ nổ pháo sáng ở Đài Tiên Khí giữa đồng không — một
// lỗi trông như lỗi hiệu ứng chứ không như lỗi giao thức, và sẽ mất nhiều giờ để tìm.
const EventKindUnknown = 255

// EventKindIndex trả chỉ số của một loại sự kiện.
func EventKindIndex(kind string) uint8 {
	if i, ok := eventIndex[kind]; ok {
		return i
	}
	return EventKindUnknown
}

// EventKindName tra ngược từ chỉ số, dùng trong test và công cụ đo.
func EventKindName(i uint8) string {
	if int(i) < len(EventKinds) {
		return EventKinds[i]
	}
	return ""
}

// MonsterKinds: thứ tự là GIAO THỨC, như `ItemKinds`. Khớp `config.monsters.kinds`.
//
// Gửi 1 byte thay cho chuỗi: phần `m` chiếm 27% gói lớn nhất ở bản v1, và mỗi quái mới vào
// tầm nhìn chở theo tên loại của nó.
var MonsterKinds = []string{"small", "medium", "large", "boss"}
