package world

import (
	"testing"

	"kiemxoay/server/internal/config"
)

func TestLoadStarterMaps(t *testing.T) {
	cfg := config.MustLoadDefault()
	maps, err := LoadDir(config.FindRepoFile("maps"), LoadOptions{HP: cfg.Obstacles.HP, FootprintW: cfg.Obstacles.FootprintW, FootprintH: cfg.Obstacles.FootprintH})
	if err != nil {
		t.Fatal(err)
	}
	if len(maps) != 3 {
		t.Fatalf("loaded %d maps, want 3", len(maps))
	}
	for _, m := range maps {
		if m.Size != cfg.Room.MapSize || m.Name == "" {
			t.Errorf("%s: size %v name %q", m.ID, m.Size, m.Name)
		}
		walls, breakable := 0, 0
		for _, o := range m.Obstacles {
			if o.MaxHP == 0 {
				walls++
			} else {
				breakable++
				if o.MaxHP != cfg.Obstacles.HP[o.Kind] {
					t.Errorf("%s: %s hp %d", m.ID, o.Kind, o.MaxHP)
				}
			}
		}
		if walls == 0 || breakable == 0 || len(m.Gates) < 2 || len(m.Zones) == 0 {
			t.Errorf("%s: walls %d breakable %d gates %d zones %d", m.ID, walls, breakable, len(m.Gates), len(m.Zones))
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
