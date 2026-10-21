package room

import (
	"bufio"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"kiemxoay/server/internal/config"
)

// NameFilter validates player names: 2–16 characters, Vietnamese allowed, profanity masked (spec §2).
type NameFilter struct {
	min, max int
	bad      []string
}

// LoadNameFilter reads one banned word per line; missing file means no word filter.
func LoadNameFilter(cfg *config.Config, path string) *NameFilter {
	f := &NameFilter{min: cfg.Player.NameMin, max: cfg.Player.NameMax}
	file, err := os.Open(path)
	if err != nil {
		return f
	}
	defer file.Close()
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		w := strings.ToLower(strings.TrimSpace(sc.Text()))
		if w != "" && !strings.HasPrefix(w, "#") {
			f.bad = append(f.bad, w)
		}
	}
	return f
}

// Clean returns a safe display name, or "" when the name is unusable.
func (f *NameFilter) Clean(name string) string {
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
	lower := strings.ToLower(name)
	for _, w := range f.bad {
		for {
			i := strings.Index(lower, w)
			if i < 0 {
				break
			}
			stars := strings.Repeat("*", utf8.RuneCountInString(w))
			name = name[:i] + stars + name[i+len(w):]
			lower = lower[:i] + stars + lower[i+len(w):]
		}
	}
	return name
}
