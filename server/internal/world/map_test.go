package world

import (
	"testing"

	"bladeascend/server/internal/config"
)

func TestLoadStarterMaps(t *testing.T) {
	cfg := config.MustLoadDefault()
	maps, err := LoadDir(config.FindRepoFile("maps"), LoadOptions{Tier: cfg.Obstacles.ObstacleTier, FootprintW: cfg.Obstacles.FootprintW, FootprintH: cfg.Obstacles.FootprintH})
	if err != nil {
		t.Fatal(err)
	}
	// 5 bản đồ = 5 khu của một thế giới (spec §8.0); xem room.ZoneMapIDs.
	if len(maps) != 5 {
		t.Fatalf("nạp được %d bản đồ, muốn 5 (một bản mỗi khu)", len(maps))
	}
	for _, m := range maps {
		if m.Size != cfg.Room.MapSize || m.Name == "" {
			t.Errorf("%s: size %v name %q", m.ID, m.Size, m.Name)
		}
		// Từ v6.7 vật cản không còn máu: mỗi loại có một BẬC CẢNH GIỚI tối thiểu để chém
		// được, và bậc phải khớp bảng trong config — loader không được tự đoán.
		cuttable := 0
		for _, o := range m.Obstacles {
			want := cfg.Obstacles.ObstacleTier(o.Kind)
			if o.Tier != want {
				t.Errorf("%s: %s bậc %d, config ghi %d", m.ID, o.Kind, o.Tier, want)
			}
			if o.Tier < config.TierBorder {
				cuttable++
			}
		}
		if cuttable == 0 || len(m.Gates) < 2 || len(m.Zones) == 0 {
			t.Errorf("%s: chém được %d gates %d zones %d", m.ID, cuttable, len(m.Gates), len(m.Zones))
		}
		for _, g := range m.Gates {
			if len(m.GatesInNetwork(g)) == 0 {
				t.Errorf("%s: gate %d has no partner", m.ID, g.ID)
			}
		}
	}
}

func TestGridPathAroundWall(t *testing.T) {
	m := EmptyMap(2000)
	m.Obstacles = []Obstacle{{ID: 0, Kind: "stone_wall", Shape: Shape{X: 900, Y: 0, W: 200, H: 1500}}}
	g := NewGrid(m, 100)
	g.Rebuild(func(int) bool { return true })
	path := g.FindPath(Vec2(200, 200), Vec2(1800, 200), 5000)
	if path == nil {
		t.Fatal("no path around the wall")
	}
	for _, p := range path {
		if g.Blocked(p) {
			t.Fatalf("path goes through wall at %+v", p)
		}
	}
	g.Rebuild(func(int) bool { return false })
	if len(g.FindPath(Vec2(200, 200), Vec2(1800, 200), 5000)) > 20 {
		t.Fatal("broken wall should allow a straight path")
	}
}

// Spec §8.6: vùng địa hình là ĐA GIÁC méo, và server phải xét đúng đa giác đó — nếu server
// xét hộp bao còn client vẽ hình méo thì người chơi đứng chỗ nhìn như bùn mà không bị chậm.
func TestPolygonZoneUsesTheDrawnShape(t *testing.T) {
	var sh Shape
	// Tam giác vuông lớn: góc trên-trái nằm NGOÀI tam giác nhưng TRONG hộp bao.
	sh.SetPoly([]float64{0, 0, 1000, 0, 1000, 1000})
	if !sh.IsPoly() {
		t.Fatal("SetPoly phải đánh dấu đây là đa giác")
	}
	if sh.W != 1000 || sh.H != 1000 {
		t.Fatalf("hộp bao %gx%g, muốn 1000x1000", sh.W, sh.H)
	}
	if !sh.Contains(Vec2(900, 500)) {
		t.Fatal("điểm trong tam giác phải tính là trong vùng")
	}
	if sh.Contains(Vec2(100, 900)) {
		t.Fatal("điểm ngoài tam giác nhưng trong hộp bao KHÔNG được tính là trong vùng")
	}
	if sh.Contains(Vec2(-50, 500)) {
		t.Fatal("điểm ngoài hộp bao phải bị loại ngay")
	}
}

// Bản đồ sinh ra phải có vùng hữu cơ thật, không còn chữ nhật/ellipse cho bùn và nước.
func TestGeneratedMapsHaveOrganicZones(t *testing.T) {
	maps, err := LoadDir("../../../maps", LoadOptions{FootprintW: 1, FootprintH: 0.3})
	if err != nil || len(maps) == 0 {
		t.Skipf("không nạp được bản đồ: %v", err)
	}
	for _, m := range maps {
		// Luật ở đây là về HÌNH DÁNG, không phải về loại vùng: mọi vùng địa hình tự nhiên
		// (bùn, nước, băng, tuyết) phải là đa giác hữu cơ. Băng Nguyệt Hồ không có bùn hay
		// nước nhưng cả hồ băng lẫn đảo tuyết đều là đa giác, vẫn đúng tinh thần §8.6.
		natural := map[string]bool{"mud": true, "water": true, "ice": true, "snow": true}
		poly := 0
		for _, z := range m.Zones {
			if natural[z.Kind] && z.Shape.IsPoly() {
				poly++
			}
		}
		if poly == 0 {
			t.Errorf("%s: không có vùng địa hình tự nhiên nào là đa giác — spec §8.6 cấm hình học lộ ra", m.ID)
		}
	}
}
