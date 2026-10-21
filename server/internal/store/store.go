// Package store persists finished "lives" in SQLite and serves the global leaderboard (spec §11).
package store

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite" // pure-Go driver, no CGO
)

// Life is one player's run from spawn to death.
type Life struct {
	GuestID     string  `json:"-"`
	Name        string  `json:"name"`
	Character   string  `json:"character"`
	PeakExp     float64 `json:"peak_exp"`
	PeakLevel   int     `json:"peak_level"`
	Kills       int     `json:"kills"`
	SurviveSecs float64 `json:"survive_secs"`
	Map         string  `json:"map"`
	EndedAt     int64   `json:"ended_at"`
}

type Store struct {
	db       *sql.DB
	cacheTTL time.Duration
	mu       sync.Mutex
	cache    map[string]cached
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
CREATE INDEX IF NOT EXISTS lives_kills ON lives(kills);`

// Open opens (or creates) the database at path. Use ":memory:" for tests.
func Open(path string, cacheTTL time.Duration) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite: one writer; also keeps :memory: on a single connection
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	return &Store{db: db, cacheTTL: cacheTTL, cache: map[string]cached{}}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// Record saves a finished life. Callers must never pass bots (spec §9.3).
func (s *Store) Record(l Life) error {
	_, err := s.db.Exec(`INSERT INTO lives (guest_id,name,character,peak_exp,peak_level,kills,survive_secs,map,ended_at)
		VALUES (?,?,?,?,?,?,?,?,?)`, l.GuestID, l.Name, l.Character, l.PeakExp, l.PeakLevel, l.Kills, l.SurviveSecs, l.Map, l.EndedAt)
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

// Top returns the top 50 lives for a period and metric ("exp" | "kills"), cached for cacheTTL.
func (s *Store) Top(period, metric string, now time.Time) ([]Life, error) {
	col := map[string]string{"exp": "peak_exp", "kills": "kills"}[metric]
	if col == "" {
		return nil, fmt.Errorf("bad metric %q", metric)
	}
	since, err := periodStart(period, now)
	if err != nil {
		return nil, err
	}
	key := period + "/" + metric
	s.mu.Lock()
	if c, ok := s.cache[key]; ok && now.Sub(c.at) < s.cacheTTL {
		s.mu.Unlock()
		return c.rows, nil
	}
	s.mu.Unlock()
	rows, err := s.query(col, since)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.cache[key] = cached{at: now, rows: rows}
	s.mu.Unlock()
	return rows, nil
}

func (s *Store) query(col string, since int64) ([]Life, error) {
	// col comes from a fixed whitelist above, never from user input.
	rs, err := s.db.Query(`SELECT name,character,peak_exp,peak_level,kills,survive_secs,map,ended_at FROM lives
		WHERE ended_at >= ? ORDER BY `+col+` DESC, ended_at ASC LIMIT 50`, since)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	out := []Life{}
	for rs.Next() {
		var l Life
		if err := rs.Scan(&l.Name, &l.Character, &l.PeakExp, &l.PeakLevel, &l.Kills, &l.SurviveSecs, &l.Map, &l.EndedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rs.Err()
}
