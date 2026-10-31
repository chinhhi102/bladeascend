package room

import (
	"testing"
)

// Bot MỚI SINH không bao giờ vượt `bots.max_spawn_realm` (Hợp Thể, chủ dự án chốt 31/10).
//
// Không có trần thì `spawn_rank` chép thẳng cảnh giới của người hạng N: phòng có một người
// lên Chân Tiên là mọi bot mới cũng sinh thẳng ở Chân Tiên, và người chơi vừa vào đã gặp cả
// phòng Tiên giới mà không ai tu lên được như thế.
func TestBotSpawnRealmCapped(t *testing.T) {
	m, _, cfg := testManager(t)
	cs := joinN(m, 1)
	r := m.AllRooms()[0]
	r.Step(1.0 / 30) // người chơi chỉ thật sự vào bàn cờ sau một nhịp Step
	// `spawn_rank` mặc định là 11, mà phòng test chỉ có 1 người → `Ranking()` không đủ dài
	// nên `randomBotLevel` rơi về Luyện Khí và bài test xanh mà chẳng kiểm gì. Đặt hạng 1
	// để nó thật sự chép cảnh giới của người đứng đầu.
	r.cfg.Bots.SpawnRank = 1
	prog := r.zones[0].G.Prog
	maxR := cfg.Bots.MaxSpawnRealm
	if maxR != 6 {
		t.Fatalf("max_spawn_realm = %d, bài test giả định 6 (Hợp Thể)", maxR)
	}

	forceTopPlayer(t, r, cs[0].PlayerID, prog.TienLevel(prog.TienCount()-1)) // Đạo Tổ đứng hạng 1
	for i := 0; i < 200; i++ {
		level, _ := r.randomBotLevel()
		if got := prog.Realm(level); got > maxR {
			t.Fatalf("bot sinh ở cảnh giới %d (%s), trần là %d (%s)",
				got, prog.RealmName(got), maxR, prog.RealmName(maxR))
		}
	}
}

// Trần âm = không giới hạn (giữ đường thoát cho công cụ đo).
func TestBotSpawnRealmUncappedWhenNegative(t *testing.T) {
	m, _, _ := testManager(t)
	cs := joinN(m, 1)
	r := m.AllRooms()[0]
	r.Step(1.0 / 30)
	r.cfg.Bots.SpawnRank = 1
	r.cfg.Bots.MaxSpawnRealm = -1
	prog := r.zones[0].G.Prog
	forceTopPlayer(t, r, cs[0].PlayerID, prog.TienLevel(0)) // Chân Tiên đứng hạng 1
	saw := false
	for i := 0; i < 200; i++ {
		if level, _ := r.randomBotLevel(); prog.Realm(level) > 6 {
			saw = true
			break
		}
	}
	if !saw {
		t.Error("trần -1 mà bot vẫn bị chặn dưới Hợp Thể")
	}
}

// forceTopPlayer đẩy một người lên cảnh giới `level` VÀ lên hạng 1.
//
// Phải ghim cả Uy Danh, không chỉ cảnh giới: `Ranking()` xếp theo Uy Danh trước, mà phòng
// test đã có sẵn một đám bot. Chỉ nâng cảnh giới thì hạng 1 vẫn là một con bot Luyện Khí,
// `randomBotLevel` chép đúng con bot đó, và bài test xanh mà chẳng kiểm gì.
func forceTopPlayer(t *testing.T, r *Room, id uint32, level int) {
	t.Helper()
	p := r.Zone(0).G.Players[id]
	if p == nil {
		t.Fatalf("không tìm thấy người chơi %d trong khu 0", id)
	}
	r.Zone(0).G.ForceState(p, level, 20, p.Pos)
	p.Score = 1e9
}
