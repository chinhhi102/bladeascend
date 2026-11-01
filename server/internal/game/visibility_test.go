package game

import "testing"

// Người đứng ngoài mép màn hình nhưng vòng kiếm với tới trong màn hình thì vẫn phải được
// gửi xuống — nếu không, kiếm vô hình chém trúng mình, nhìn như lag.
func TestPlayerVisibleWhenOnlyTheirRingReachesTheView(t *testing.T) {
	g := quietGame(t, nil)
	viewer := g.AddPlayer(SpawnOptions{Name: "V", Level: 0, Swords: 1})
	big := g.AddPlayer(SpawnOptions{Name: "Big", Level: 80, Swords: 400})
	place(g, viewer, Vec{X: 6000, Y: 6000}, 0)
	hw, _ := g.ViewHalfExtents(viewer)
	reach := g.Geom(big).Reach
	if reach < 200 {
		t.Fatalf("thiết lập sai: tầm kiếm %v quá ngắn để kiểm tra", reach)
	}
	// Thân nằm ngoài mép màn hình, nhưng mũi kiếm thò vào trong.
	place(g, big, Vec{X: 6000 + hw + reach*0.5, Y: 6000}, 0)
	if !g.CanSee(viewer, big) {
		t.Fatal("người ngoài mép nhưng vòng kiếm với tới trong màn hình phải được gửi xuống")
	}
	// Ra đủ xa để cả vòng kiếm cũng không chạm màn hình thì mới được cắt.
	place(g, big, Vec{X: 6000 + hw + reach*2, Y: 6000}, 0)
	if g.CanSee(viewer, big) {
		t.Fatal("ra ngoài hẳn tầm kiếm rồi vẫn còn gửi: phí băng thông")
	}
}

// Perceivable là CanSee bỏ phép thử khung nhìn (NETCODE_SPEC §4.4): đứng xa vẫn "cảm nhận
// được" nếu không bị che, còn trốn trong bụi thì xa bao nhiêu cũng không.
func TestPerceivableIgnoresViewButKeepsBushHiding(t *testing.T) {
	g := quietGame(t, nil)
	viewer := g.AddPlayer(SpawnOptions{Name: "V", Level: 0, Swords: 1})
	other := g.AddPlayer(SpawnOptions{Name: "O", Level: 0, Swords: 1})
	place(g, viewer, Vec{X: 6000, Y: 6000}, 0)
	hw, _ := g.ViewHalfExtents(viewer)
	place(g, other, Vec{X: 6000 + hw + 600, Y: 6000}, 0)
	if g.CanSee(viewer, other) {
		t.Fatal("thiết lập sai: người thử phải nằm NGOÀI khung nhìn")
	}
	if !g.Perceivable(viewer, other) {
		t.Fatal("ngoài khung nhìn nhưng không bị che thì Perceivable phải đúng")
	}
	other.Bush = 0
	other.RevealedUntil = 0
	if g.Perceivable(viewer, other) {
		t.Fatal("trốn trong bụi ở xa mà Perceivable vẫn đúng: server sẽ lộ vị trí")
	}
}
