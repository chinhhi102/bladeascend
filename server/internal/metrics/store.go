package metrics

import (
	"database/sql"
	"fmt"
	"time"
)

// Bảng theo ADMIN_SPEC §12.2. Khoá chính (ts, metric, scope) nên chạy lại bản cuộn hai lần
// không nhân đôi dữ liệu — `INSERT OR REPLACE`.
const schema = `
PRAGMA journal_mode=WAL;
CREATE TABLE IF NOT EXISTS metrics_minute (
	ts INTEGER NOT NULL, metric TEXT NOT NULL, scope TEXT NOT NULL,
	build TEXT NOT NULL DEFAULT '',
	count INTEGER NOT NULL, min REAL NOT NULL, max REAL NOT NULL, sum REAL NOT NULL,
	p95 REAL NOT NULL, hist BLOB,
	PRIMARY KEY (ts, metric, scope)
) WITHOUT ROWID;
CREATE TABLE IF NOT EXISTS metrics_hour (
	ts INTEGER NOT NULL, metric TEXT NOT NULL, scope TEXT NOT NULL,
	build TEXT NOT NULL DEFAULT '',
	count INTEGER NOT NULL, min REAL NOT NULL, max REAL NOT NULL, sum REAL NOT NULL,
	p95 REAL NOT NULL, hist BLOB,
	PRIMARY KEY (ts, metric, scope)
) WITHOUT ROWID;
CREATE TABLE IF NOT EXISTS metrics_day (
	ts INTEGER NOT NULL, metric TEXT NOT NULL, scope TEXT NOT NULL,
	build TEXT NOT NULL DEFAULT '',
	count INTEGER NOT NULL, min REAL NOT NULL, max REAL NOT NULL, sum REAL NOT NULL,
	p95 REAL NOT NULL, hist BLOB,
	PRIMARY KEY (ts, metric, scope)
) WITHOUT ROWID;
CREATE TABLE IF NOT EXISTS deploy_marks (
	ts INTEGER PRIMARY KEY, build TEXT NOT NULL, note TEXT NOT NULL DEFAULT ''
);`

// Giữ dữ liệu bao lâu (§12.2): phút 14 ngày, giờ 180 ngày, ngày giữ mãi.
const (
	keepMinutes = 14 * 24 * time.Hour
	keepHours   = 180 * 24 * time.Hour
)

// Row là một dòng số liệu đã gom.
type Row struct {
	TS            int64
	Metric, Scope string
	Build         string
	Count         int64
	Min, Max, Sum float64
	P95           float64
	Hist          Hist
}

// Avg tiện cho trang quản trị: tổng chia số mẫu.
func (r Row) Avg() float64 {
	if r.Count == 0 {
		return 0
	}
	return r.Sum / float64(r.Count)
}

// DB mở (hoặc tạo) file số liệu và dựng bảng.
type DB struct{ db *sql.DB }

// Open dùng *sql.DB có sẵn — gọi từ cmd/server với file riêng `metrics.db`.
func Open(db *sql.DB) (*DB, error) {
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("schema metrics: %w", err)
	}
	return &DB{db: db}, nil
}

// SQL trả handle thô (cho test và cho trang quản trị dùng chung kết nối).
func (d *DB) SQL() *sql.DB { return d.db }

// MarkDeploy ghi một mốc deploy để trang Hiệu năng vẽ vạch dọc (§12.2).
func (d *DB) MarkDeploy(ts int64, build, note string) error {
	_, err := d.db.Exec(`INSERT OR REPLACE INTO deploy_marks (ts, build, note) VALUES (?,?,?)`, ts, build, note)
	return err
}

// DeployMarks đọc các mốc deploy trong khoảng.
func (d *DB) DeployMarks(from, to int64) ([]struct {
	TS    int64
	Build string
	Note  string
}, error) {
	rows, err := d.db.Query(`SELECT ts,build,note FROM deploy_marks WHERE ts BETWEEN ? AND ? ORDER BY ts`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		TS    int64
		Build string
		Note  string
	}
	for rows.Next() {
		var m struct {
			TS    int64
			Build string
			Note  string
		}
		if err := rows.Scan(&m.TS, &m.Build, &m.Note); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// WriteMinute ghi cả lô một phút trong MỘT giao dịch (§12.2: ghi theo lô 1 lần/phút).
func (d *DB) WriteMinute(rows []Row) error { return d.write("metrics_minute", rows) }

func (d *DB) write(table string, rows []Row) error {
	if len(rows) == 0 {
		return nil
	}
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // Commit ở dưới; Rollback sau Commit là no-op
	st, err := tx.Prepare(`INSERT OR REPLACE INTO ` + table +
		` (ts,metric,scope,build,count,min,max,sum,p95,hist) VALUES (?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer st.Close()
	for _, r := range rows {
		if _, err := st.Exec(r.TS, r.Metric, r.Scope, r.Build, r.Count, r.Min, r.Max, r.Sum, r.P95, EncodeHist(r.Hist)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Query đọc các dòng của một bảng trong khoảng, lọc theo chỉ số/phạm vi nếu có.
func (d *DB) Query(table string, from, to int64, metric, scope string, limit int) ([]Row, error) {
	if !validTable(table) {
		return nil, fmt.Errorf("bảng không hợp lệ: %s", table)
	}
	q := `SELECT ts,metric,scope,build,count,min,max,sum,p95,hist FROM ` + table + ` WHERE ts BETWEEN ? AND ?`
	args := []any{from, to}
	if metric != "" {
		q += ` AND metric = ?`
		args = append(args, metric)
	}
	if scope != "" {
		q += ` AND scope = ?`
		args = append(args, scope)
	}
	q += ` ORDER BY ts LIMIT ?`
	args = append(args, limit)
	rows, err := d.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Row
	for rows.Next() {
		var r Row
		var blob []byte
		if err := rows.Scan(&r.TS, &r.Metric, &r.Scope, &r.Build, &r.Count, &r.Min, &r.Max, &r.Sum, &r.P95, &blob); err != nil {
			return nil, err
		}
		r.Hist = DecodeHist(blob)
		out = append(out, r)
	}
	return out, rows.Err()
}

// Catalog liệt kê những cặp (chỉ số, phạm vi) đang có, để trang quản trị dựng ô chọn mà
// không cần biết trước tên.
func (d *DB) Catalog(since int64) (map[string][]string, error) {
	rows, err := d.db.Query(`SELECT DISTINCT metric, scope FROM metrics_minute WHERE ts >= ?
		UNION SELECT DISTINCT metric, scope FROM metrics_hour WHERE ts >= ? ORDER BY metric, scope`, since, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string][]string{}
	for rows.Next() {
		var m, s string
		if err := rows.Scan(&m, &s); err != nil {
			return nil, err
		}
		out[m] = append(out[m], s)
	}
	return out, rows.Err()
}

func validTable(t string) bool {
	return t == "metrics_minute" || t == "metrics_hour" || t == "metrics_day"
}

// Rollup cuộn phút → giờ và giờ → ngày cho mọi khoảng ĐÃ ĐÓNG trước `now`, rồi dọn hạn.
//
// Cuộn lại một khoảng đã cuộn là vô hại (khoá chính + INSERT OR REPLACE), nên hàm này gọi
// được bao nhiêu lần tuỳ thích — kể cả sau khi server tắt rồi bật lại giữa giờ.
func (d *DB) Rollup(now time.Time) error {
	if err := d.rollInto("metrics_minute", "metrics_hour", now.Truncate(time.Hour).Unix(), time.Hour); err != nil {
		return err
	}
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if err := d.rollInto("metrics_hour", "metrics_day", day.Unix(), 24*time.Hour); err != nil {
		return err
	}
	return d.prune(now)
}

// rollInto gom mọi dòng của `src` có ts < `before` thành các khoảng `step` trong `dst`.
func (d *DB) rollInto(src, dst string, before int64, step time.Duration) error {
	rows, err := d.db.Query(`SELECT ts,metric,scope,build,count,min,max,sum,hist FROM `+src+
		` WHERE ts < ? AND ts >= COALESCE((SELECT MAX(ts) FROM `+dst+`), 0) ORDER BY ts`, before)
	if err != nil {
		return err
	}
	type key struct {
		ts            int64
		metric, scope string
	}
	agg := map[key]*Row{}
	for rows.Next() {
		var ts int64
		var metric, scope, build string
		var count int64
		var min, max, sum float64
		var blob []byte
		if err := rows.Scan(&ts, &metric, &scope, &build, &count, &min, &max, &sum, &blob); err != nil {
			rows.Close()
			return err
		}
		k := key{ts - ts%int64(step.Seconds()), metric, scope}
		r := agg[k]
		if r == nil {
			r = &Row{TS: k.ts, Metric: metric, Scope: scope, Build: build}
			agg[k] = r
		}
		// Build của khoảng gộp: lấy build CUỐI CÙNG trong khoảng. Một giờ có deploy thì hai
		// build cùng góp số liệu; vạch deploy trên biểu đồ mới là chỗ nói chính xác lúc nào.
		r.Build = build
		dist := Dist{Count: count, Min: min, Max: max, Sum: sum, Hist: DecodeHist(blob)}
		cur := Dist{Count: r.Count, Min: r.Min, Max: r.Max, Sum: r.Sum, Hist: r.Hist}
		cur.Merge(dist)
		r.Count, r.Min, r.Max, r.Sum, r.Hist = cur.Count, cur.Min, cur.Max, cur.Sum, cur.Hist
		r.P95 = cur.P95()
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	out := make([]Row, 0, len(agg))
	for _, r := range agg {
		out = append(out, *r)
	}
	return d.write(dst, out)
}

func (d *DB) prune(now time.Time) error {
	if _, err := d.db.Exec(`DELETE FROM metrics_minute WHERE ts < ?`, now.Add(-keepMinutes).Unix()); err != nil {
		return err
	}
	_, err := d.db.Exec(`DELETE FROM metrics_hour WHERE ts < ?`, now.Add(-keepHours).Unix())
	return err
}
