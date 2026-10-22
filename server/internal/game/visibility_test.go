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
