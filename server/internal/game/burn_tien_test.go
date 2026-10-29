package game

import "testing"

// Thiêu Đốt Tinh Huyết tiêu EXP của tầng hiện tại (§4.1c), mà người Tiên giới không còn
// hấp thụ Linh khí nữa (§3.4) — họ lên bằng sát mạng. Bài test này khoá lại hậu quả của
// hai luật đó cộng vào nhau: **lên Tiên giới là mất luôn khả năng chạy nhanh**.
//
// Test không phán đây là đúng hay sai; nó chỉ bắt buộc ai đổi luật phải thấy chỗ này.
// Câu hỏi "Tiên giới nên đốt Tiên khí chứ?" đang chờ chủ dự án ở Q24.
func TestBurnStopsInTienGioi(t *testing.T) {
	g := rulesGame(t)
	p := g.AddPlayer(SpawnOptions{Name: "A", Char: g.Cfg.Characters[0]})

	// Người thường có EXP tầng để đốt.
	g.GainExp(p, g.Prog.ExpNeeded(p.Level)*0.9, false)
	p.Moving, p.Burning = true, true
	if !g.Burning(p) {
		t.Fatal("người thường đang chạy và còn EXP tầng thì phải đốt được")
	}

	// Lên thẳng cảnh giới Tiên giới cao nhất.
	top := g.Prog.TienLevel(len(g.Cfg.Cultivation.TienRealms) - 1)
	g.ForceState(p, top, 10, p.Pos)
	p.Moving, p.Burning = true, true
	if !g.BanksTienKhi(p) {
		t.Fatalf("level %d phải là Tiên giới", p.Level)
	}
	if g.Burning(p) {
		t.Error("đang kỳ vọng Tiên giới KHÔNG đốt được (không còn EXP tầng) — nếu luật đã đổi thì sửa test và Q24")
	}

	// Và Tiên khí dù đầy cũng không giúp gì: đốt chỉ nhìn vào EXP.
	g.GainTienKhi(p, 500)
	if g.Burning(p) {
		t.Error("Tiên khí không phải nhiên liệu của Thiêu Đốt — xem Q24")
	}
}
