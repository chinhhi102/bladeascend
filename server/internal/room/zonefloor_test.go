package room

import (
	"encoding/json"
	"testing"

	"bladeascend/server/internal/config"
	"bladeascend/server/internal/netproto"
	"bladeascend/server/internal/world"
)

// Sàn bot theo khu — spec v6.35 §9.5.

func TestZoneFloorsJSON(t *testing.T) {
	var f config.ZoneFloors
	if err := json.Unmarshal([]byte(`10`), &f); err != nil || f.For("thien_dinh") != 10 || f.For("pham_tran") != 10 {
		t.Fatalf("dạng cũ (một số) phải áp cho mọi khu: %v %v", f, err)
	}
	if err := json.Unmarshal([]byte(`{"thien_dinh": 30, "default": 20}`), &f); err != nil {
		t.Fatal(err)
	}
	if f.For("thien_dinh") != 30 || f.For("truc_lam") != 20 {
		t.Fatalf("sàn theo khu sai: %v", f)
	}
}

// Config production: Thiên Đình 30, khu rìa 20; staging 15 / 10.
func TestZoneFloorsEnvDefaults(t *testing.T) {
	for env, want := range map[string][2]int{config.EnvProduction: {30, 20}, config.EnvStaging: {15, 10}} {
		cfg := config.MustLoadDefault()
		e := cfg.ApplyEnv(env)
		if e.ZoneMinChars.For("thien_dinh") != want[0] || e.ZoneMinChars.For("linh_son_coc") != want[1] {
			t.Errorf("%s: sàn %v, cần Thiên Đình %d · khu rìa %d", env, e.ZoneMinChars, want[0], want[1])
		}
	}
}

// Phòng vắng production với 5 bản đồ thật: bot lấp Thiên Đình đủ 30, mỗi khu rìa đủ 20,
// và không rút bot của Thiên Đình cho khu rìa (≈ 110 nhân vật).
func TestZoneFloorsFillPerZone(t *testing.T) {
	if testing.Short() {
		t.Skip("nạp 5 bản đồ thật")
	}
	cfg := config.MustLoadDefault()
	cfg.ApplyEnv(config.EnvProduction)
	cfg.Bots.ModeDefault = BotOn
	maps, err := world.LoadDir(config.FindRepoFile("maps"), world.LoadOptions{Tier: cfg.Obstacles.ObstacleTier,
		FootprintW: cfg.Obstacles.FootprintW, FootprintH: cfg.Obstacles.FootprintH})
	if err != nil {
		t.Fatal(err)
	}
	m := NewManager(cfg, maps, &memRecorder{})
	m.Start = func(*Room) {}
	m.Join(NewClient(&recConn{}, false), netproto.Join{Name: "Teo", Char: "a_hon"})
	r := m.AllRooms()[0]
	dt := 1 / float64(cfg.Net.TickHz)
	for i := 0; i < int(90/dt); i++ {
		r.Step(dt)
	}
	for _, z := range r.Zones() {
		want := cfg.Room.ZoneMinChars.For(z.G.Map.ID)
		if c := z.Count(); c < want || c > want+1 {
			t.Errorf("khu %s (%s): %d nhân vật, cần %d", z.Name, z.G.Map.ID, c, want)
		}
	}
}

// Admin đặt sàn riêng từng khu; số âm / quá lớn bị từ chối.
func TestZoneFloorsSettingsValidate(t *testing.T) {
	ok := RoomSettings{MaxHumans: 50, ZoneMinChars: config.ZoneFloors{"thien_dinh": 30, "default": 20}, BotAutoOffAt: 40}
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	bad := ok
	bad.ZoneMinChars = config.ZoneFloors{"thien_dinh": 99, "default": 20}
	if bad.Validate() == nil {
		t.Fatal("sàn 99 phải bị từ chối")
	}
}
