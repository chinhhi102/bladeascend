package room

import (
	"math"
	"testing"
	"unicode"
	"unicode/utf8"

	"bladeascend/server/internal/config"
	"bladeascend/server/internal/netproto"
)

// Tên bot đa quốc gia (spec v6.34 §9.4).

func joinLang(m *Manager, name, lang string) {
	c := NewClient(&recConn{}, false)
	c.Lang = lang
	m.Join(c, netproto.Join{Name: name, Char: "a_hon"})
}

func langCounts(r *Room, n int) map[string]float64 {
	r.Step(1.0 / 30) // người vào phòng qua hộp thư, chạy một tick cho chắc đã vào khu
	if r.Humans() == 0 {
		panic("chưa có người thật trong phòng")
	}
	out := map[string]float64{}
	for i := 0; i < n; i++ {
		out[r.pickBotLang()]++
	}
	for k := range out {
		out[k] /= float64(n)
	}
	return out
}

func near(t *testing.T, what string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.02 {
		t.Errorf("%s: %.3f, cần ≈ %.2f", what, got, want)
	}
}

// Phòng toàn người Brazil, toàn server 7 ngày toàn người Thái → 60% pt, 25% th, 15% intl.
func TestBotLangMix(t *testing.T) {
	m, _, _ := testManager(t)
	m.LangShares = func() map[string]float64 { return map[string]float64{"th": 1} }
	joinLang(m, "Gui", "pt")
	joinLang(m, "Tata", "pt")
	got := langCounts(m.AllRooms()[0], 40000)
	near(t, "pt (người trong phòng)", got["pt"], 0.60)
	near(t, "th (toàn server)", got["th"], 0.25)
	near(t, "intl", got["intl"], 0.15)
}

// Server chưa có số liệu 7 ngày → phần 25% cũng theo người trong phòng: 1 vi + 1 en →
// 42.5% mỗi bên, 15% intl.
func TestBotLangRoomNoStats(t *testing.T) {
	m, _, _ := testManager(t)
	joinLang(m, "Teo", "vi")
	joinLang(m, "Bong", "en")
	m.AllRooms()[0].Step(1.0 / 30)
	got := langCounts(m.AllRooms()[0], 20000)
	near(t, "vi", got["vi"], 0.425)
	near(t, "en", got["en"], 0.425)
	near(t, "intl", got["intl"], 0.15)
	if len(got) != 3 {
		t.Errorf("chỉ được vi + en + intl, ra %v", got)
	}
}

// Phòng chưa có người rõ ngôn ngữ → toàn bộ phần không-intl theo toàn server.
func TestBotLangEmptyRoomUsesServer(t *testing.T) {
	m, _, _ := testManager(t)
	m.LangShares = func() map[string]float64 { return map[string]float64{"es": 0.5, "id": 0.5} }
	joinLang(m, "Khach", "")
	got := langCounts(m.AllRooms()[0], 20000)
	near(t, "es", got["es"], 0.425)
	near(t, "id", got["id"], 0.425)
	near(t, "intl", got["intl"], 0.15)
}

func TestNormLang(t *testing.T) {
	for in, want := range map[string]string{"pt-BR": "pt", "EN": "en", "th": "th", "intl": "", "fr": "", "": ""} {
		if got := config.NormLang(in); got != want {
			t.Errorf("NormLang(%q) = %q, cần %q", in, got, want)
		}
	}
}

// CI dữ liệu (spec §9.4): mỗi tệp ≥ 400 tên (intl ≥ 300), không lọt bộ lọc từ tục của
// ngôn ngữ đó, độ dài hợp lệ, chỉ chữ cái (Latin có dấu, Thái) + số + `_ . ~ ^`, không trùng,
// không chữ "bot".
func TestBotNameFiles(t *testing.T) {
	cfg := config.MustLoadDefault()
	filter := LoadNameFilter(cfg, config.FindRepoFile("config/bad_words"))
	files := config.LoadBotNameDir(config.FindRepoFile("config/bot_names"))
	for _, lang := range config.BotNameLangs {
		names := files[lang]
		min := 400
		if lang == "intl" {
			min = 300
		}
		if len(names) < min {
			t.Errorf("%s.txt: %d tên, cần ≥ %d", lang, len(names), min)
		}
		checkNames(t, filter, cfg, lang, names)
	}
}

func checkNames(t *testing.T, f *NameFilter, cfg *config.Config, lang string, names []string) {
	t.Helper()
	seen := map[string]bool{}
	for _, n := range names {
		if w, bad := f.HasBadWord(n, lang); bad {
			t.Errorf("%s.txt: %q chứa từ cấm %q", lang, n, w)
		}
		if l := utf8.RuneCountInString(n); l < cfg.Player.NameMin || l > cfg.Player.NameMax {
			t.Errorf("%s.txt: %q dài %d ký tự", lang, n, l)
		}
		if r, ok := badRune(n); ok {
			t.Errorf("%s.txt: %q có ký tự %q ngoài bộ cho phép", lang, n, r)
		}
		if seen[n] {
			t.Errorf("%s.txt: %q lặp", lang, n)
		}
		seen[n] = true
	}
}

func badRune(n string) (rune, bool) {
	for _, r := range n {
		ok := unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Mn, r) || r == ' ' ||
			r == '_' || r == '.' || r == '~' || r == '^'
		if !ok {
			return r, true
		}
	}
	return 0, false
}
