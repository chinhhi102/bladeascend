package room

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"bladeascend/server/internal/config"
)

// NameFilter validates player names: 2–16 characters, Vietnamese allowed, profanity masked
// (spec §2). Từ tục tách theo ngôn ngữ `config/bad_words/<lang>.txt` (spec v6.34 §9.4).
type NameFilter struct {
	min, max int
	bad      map[string][]string
}

// LoadNameFilter đọc mọi `<lang>.txt` trong dir (một từ mỗi dòng, `#` là chú thích).
// Thiếu thư mục thì chỉ kiểm độ dài.
func LoadNameFilter(cfg *config.Config, dir string) *NameFilter {
	f := &NameFilter{min: cfg.Player.NameMin, max: cfg.Player.NameMax, bad: map[string][]string{}}
	for _, l := range config.BotNameLangs {
		if words := readWordList(filepath.Join(dir, l+".txt")); len(words) > 0 {
			f.bad[l] = words
		}
	}
	return f
}

func readWordList(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	var out []string
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		w := strings.ToLower(strings.TrimSpace(sc.Text()))
		if w != "" && !strings.HasPrefix(w, "#") {
			out = append(out, w)
		}
	}
	return out
}

// wordsFor: danh sách áp cho một người chơi = vi + en + ngôn ngữ của chính họ.
//
// KHÔNG gộp cả 6 danh sách: từ tục tiếng này là tên thường tiếng kia ("Babi" là tên
// Indonesia, "Begoña" là tên Tây Ban Nha) — gộp hết thì che nhầm tên người thật.
func (f *NameFilter) wordsFor(lang string) []string {
	out := append([]string{}, f.bad["vi"]...)
	out = append(out, f.bad["en"]...)
	if l := config.NormLang(lang); l != "" && l != "vi" && l != "en" {
		out = append(out, f.bad[l]...)
	}
	return out
}

// Clean returns a safe display name, or "" when the name is unusable.
func (f *NameFilter) Clean(name, lang string) string {
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, name)
	name = strings.Join(strings.Fields(name), " ")
	if n := utf8.RuneCountInString(name); n < f.min || n > f.max {
		return ""
	}
	return maskWords(name, f.wordsFor(lang))
}

// HasBadWord: tên có chứa từ nào trong danh sách của ngôn ngữ lang không (CI tên bot).
func (f *NameFilter) HasBadWord(name, lang string) (string, bool) {
	lower := strings.ToLower(name)
	words := f.bad[lang]
	if lang == "intl" {
		words = f.bad["en"]
	}
	for _, w := range words {
		if strings.Contains(lower, w) {
			return w, true
		}
	}
	return "", false
}

// maskWords thay mỗi từ tục (không phân biệt hoa thường) bằng chừng ấy dấu `*`.
//
// Làm trên MẢNG RUNE: chữ hoa và chữ thường có thể khác số byte trong UTF-8, nên cắt theo
// vị trí byte của bản viết thường là cắt lệch tên gốc.
func maskWords(name string, words []string) string {
	src := []rune(name)
	lower := []rune(strings.ToLower(name))
	if len(lower) != len(src) {
		lower = src
	}
	for _, w := range words {
		wr := []rune(w)
		for i := 0; i+len(wr) <= len(lower); i++ {
			if string(lower[i:i+len(wr)]) != w {
				continue
			}
			for k := i; k < i+len(wr); k++ {
				src[k], lower[k] = '*', '*'
			}
		}
	}
	return string(src)
}
