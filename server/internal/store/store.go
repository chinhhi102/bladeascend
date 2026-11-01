// Package store persists finished "lives" in SQLite and serves the global leaderboard (spec §11).
package store

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite" // pure-Go driver, no CGO
)

// Life is one player's run from spawn to death (spec §11.2).
type Life struct {
	GuestID   string  `json:"-"`
	Name      string  `json:"name"`
	Character string  `json:"character"`
	Score     float64 `json:"score"` // Uy Danh, bảng chính
	Hits      int     `json:"hits"`  // số lần chém trúng người, bảng Sát Thần
	PeakExp   float64 `json:"peak_exp"`
	PeakLevel int     `json:"peak_level"`
	Kills     int     `json:"kills"`
	// DaiThuaSecs là số giây từ lúc vào trận tới Đại Thừa, 0 nếu chưa đạt (bảng Thần Tốc).
	DaiThuaSecs float64 `json:"dai_thua_secs"`
	Top1Secs    float64 `json:"top1_secs"`
	SurviveSecs float64 `json:"survive_secs"`
	Map         string  `json:"map"`
	EndedAt     int64   `json:"ended_at"`
	// SkillCasts: số lần dùng kỹ năng trong mạng — để chỉnh hồi chiêu theo trận thật
	// (spec v6.25 §6.3, telemetry ≥ 2 ngày).
	SkillCasts int `json:"skill_casts"`
}

type Store struct {
	db       *sql.DB
	cacheTTL time.Duration
	mu       sync.Mutex
	cache    map[string]cached
	// medians: mốc trung vị Uy Danh theo cảnh giới, cache 5 phút (xem median.go).
	medians medianCache
}

type cached struct {
	at   time.Time
	rows []Life
}

const schema = `
CREATE TABLE IF NOT EXISTS lives (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	guest_id TEXT NOT NULL,
	name TEXT NOT NULL,
	character TEXT NOT NULL,
	peak_exp REAL NOT NULL,
	peak_level INTEGER NOT NULL,
	kills INTEGER NOT NULL,
	survive_secs REAL NOT NULL,
	map TEXT NOT NULL,
	ended_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS lives_ended ON lives(ended_at);
CREATE INDEX IF NOT EXISTS lives_exp ON lives(peak_exp);
CREATE INDEX IF NOT EXISTS lives_kills ON lives(kills);

-- hidden_lives: mạng chơi bị ẨN khỏi bảng xếp hạng (ADMIN_SPEC §6.2).
--
-- Ẩn, KHÔNG xoá. Hai lý do. Một: ẩn sai thì bỏ ẩn là xong, còn xoá sai thì mất hẳn — và
-- người bị gắn cờ oan có quyền được phục hồi. Hai: danh sách này chính là bằng chứng, nên
-- nó phải còn lại để đối chiếu (docs/LEADERBOARD_AUDIT.md).
--
-- Khoá (guest_id, ended_at): lives.id tự tăng nên nó đổi khi nhập lại DB từ sao lưu, còn
-- cặp (ai, kết thúc lúc nào) thì không.
CREATE TABLE IF NOT EXISTS hidden_lives (
	guest_id TEXT NOT NULL,
	ended_at INTEGER NOT NULL,
	reason   TEXT NOT NULL,
	at       INTEGER NOT NULL,
	by_ip    TEXT NOT NULL,
	PRIMARY KEY (guest_id, ended_at)
);`

// migrations add the v3.3 leaderboard columns to databases created before them (spec §11.2).
// Re-running them on an up-to-date database is a no-op.
var migrations = []string{
	`ALTER TABLE lives ADD COLUMN score REAL NOT NULL DEFAULT 0`,
	`ALTER TABLE lives ADD COLUMN hits INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE lives ADD COLUMN dai_thua_secs REAL NOT NULL DEFAULT 0`,
	`ALTER TABLE lives ADD COLUMN top1_secs REAL NOT NULL DEFAULT 0`,
	`CREATE INDEX IF NOT EXISTS lives_score ON lives(score)`,
	`CREATE INDEX IF NOT EXISTS lives_hits ON lives(hits)`,
	`CREATE INDEX IF NOT EXISTS lives_speed ON lives(dai_thua_secs)`,
	`ALTER TABLE lives ADD COLUMN skill_casts INTEGER NOT NULL DEFAULT 0`, // v6.25 §6.3
}

// Open opens (or creates) the database at path. Use ":memory:" for tests.
// DB trả về kết nối SQLite bên dưới.
//
// Mở ra cho package `admin` dùng CHUNG một file (ADMIN_SPEC §8), thay vì mở file thứ hai:
// hai kết nối tới cùng một file SQLite là nguồn của "database is locked" dưới tải ghi.
func (s *Store) DB() *sql.DB { return s.db }

func Open(path string, cacheTTL time.Duration) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite: one writer; also keeps :memory: on a single connection
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return nil, fmt.Errorf("migrate %q: %w", m, err)
		}
	}
	return &Store{db: db, cacheTTL: cacheTTL, cache: map[string]cached{}}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// Record saves a finished life. Callers must never pass bots (spec §9.3).
func (s *Store) Record(l Life) error {
	_, err := s.db.Exec(`INSERT INTO lives
		(guest_id,name,character,score,hits,peak_exp,peak_level,kills,dai_thua_secs,top1_secs,survive_secs,map,ended_at,skill_casts)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		l.GuestID, l.Name, l.Character, l.Score, l.Hits, l.PeakExp, l.PeakLevel, l.Kills,
		l.DaiThuaSecs, l.Top1Secs, l.SurviveSecs, l.Map, l.EndedAt, l.SkillCasts)
	return err
}

// periodStart returns the unix time where a period begins ("day", "week", "all").
func periodStart(period string, now time.Time) (int64, error) {
	switch period {
	case "day":
		y, m, d := now.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, now.Location()).Unix(), nil
	case "week":
		y, m, d := now.Date()
		wd := (int(now.Weekday()) + 6) % 7 // Monday = 0
		return time.Date(y, m, d-wd, 0, 0, 0, 0, now.Location()).Unix(), nil
	case "all":
		return 0, nil
	default:
		return 0, fmt.Errorf("bad period %q", period)
	}
}

// board describes one global leaderboard (spec §11.2). Thần Tốc sorts ascending and only
// counts lives that actually reached Đại Thừa.
type board struct {
	col   string
	order string
	where string
}

var boards = map[string]board{
	"score": {col: "score", order: "DESC"},
	"hits":  {col: "hits", order: "DESC"},
	"speed": {col: "dai_thua_secs", order: "ASC", where: "AND dai_thua_secs > 0"},
}

// Boards lists the valid board names, for error messages and tests.
func Boards() []string { return []string{"score", "hits", "speed"} }

// Top returns the top 50 lives for a period and board ("score" | "hits" | "speed"),
// cached for cacheTTL.
func (s *Store) Top(period, name string, now time.Time) ([]Life, error) {
	b, ok := boards[name]
	if !ok {
		return nil, fmt.Errorf("bad board %q", name)
	}
	since, err := periodStart(period, now)
	if err != nil {
		return nil, err
	}
	key := period + "/" + name
	s.mu.Lock()
	if c, ok := s.cache[key]; ok && now.Sub(c.at) < s.cacheTTL {
		s.mu.Unlock()
		return c.rows, nil
	}
	s.mu.Unlock()
	rows, err := s.query(b, since)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.cache[key] = cached{at: now, rows: rows}
	s.mu.Unlock()
	return rows, nil
}

func (s *Store) query(b board, since int64) ([]Life, error) {
	// Every piece of SQL below comes from the fixed `boards` table, never from user input.
	//
	// NOT EXISTS, không LEFT JOIN: một mạng chơi bị ẩn hai lần (ẩn, bỏ ẩn, ẩn lại) vẫn chỉ là
	// một dòng trong `lives`, còn LEFT JOIN sẽ nhân nó lên và LIMIT 50 trả về ít hơn 50 người.
	rs, err := s.db.Query(`SELECT name,character,score,hits,peak_exp,peak_level,kills,dai_thua_secs,top1_secs,survive_secs,map,ended_at
		FROM lives WHERE ended_at >= ? `+b.where+`
		AND NOT EXISTS (SELECT 1 FROM hidden_lives h
			WHERE h.guest_id = lives.guest_id AND h.ended_at = lives.ended_at)
		ORDER BY `+b.col+` `+b.order+`, ended_at ASC LIMIT 50`, since)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	out := []Life{}
	for rs.Next() {
		var l Life
		if err := rs.Scan(&l.Name, &l.Character, &l.Score, &l.Hits, &l.PeakExp, &l.PeakLevel, &l.Kills,
			&l.DaiThuaSecs, &l.Top1Secs, &l.SurviveSecs, &l.Map, &l.EndedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rs.Err()
}
