package admin

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"bladeascend/server/internal/telemetry"
)

// Bộ dọn dữ liệu (janitor) — ADMIN_SPEC §13 (v1.6).
//
// ── Nguyên tắc ───────────────────────────────────────────────────────────────────────────
//
// "Dữ liệu chỉ được giữ khi còn dùng vào việc gì." Nhưng xoá là việc không lùi lại được,
// nên gói này được viết quanh ba quy tắc:
//
//  1. **Đếm trước, xoá sau.** Mọi việc xoá đều có bản `dry-run` chạy đúng cùng câu lệnh
//     `WHERE`, chỉ đổi `DELETE` thành `COUNT(*)`. Hai câu lệnh khác nhau là hai sự thật
//     khác nhau, và bản xem trước sẽ nói dối đúng vào lúc người ta tin nó.
//  2. **Gộp trước khi xoá.** Dòng lỗi thô bị xoá sau 30 ngày, nhưng CHỮ KÝ của nó (số lần,
//     build đầu/cuối) giữ mãi — xoá thẳng là mất luôn câu "lỗi này đã có từ tháng trước".
//  3. **Không bao giờ đụng** art, `assets/library` (quy tắc 13), mạng chơi đang nằm trên
//     bảng xếp hạng toàn cục, lượt thích, và CHỮ góp ý.
//
// ── Vì sao không dùng một thư viện dọn chung ─────────────────────────────────────────────
//
// Dữ liệu nằm ở BA tệp SQLite khác nhau (bảng xếp hạng + quản trị, telemetry, metrics) với
// ba vòng đời khác nhau. Một `DELETE ... WHERE ts < ?` chung sẽ bỏ sót đúng những ngoại lệ
// đáng giá nhất — "giữ mãi mạng trên bảng xếp hạng" là một câu truy vấn con, không phải
// một con số ngày.

// Hạn giữ theo bảng §13. Để thành hằng số có tên để trang quản trị và nhật ký nói cùng
// một con số với mã.
const (
	KeepRawTelemetryDays = 14
	KeepLivesDays        = 90
	KeepErrorsDays       = 30
	KeepAdminSessDays    = 30
	KeepExpiredBanDays   = 30
	KeepAuditDays        = 365
	// VacuumEveryDays: VACUUM mỗi tuần (§13), và chỉ khi 0 người thật online.
	VacuumEveryDays = 7
	// DBWarnBytes: cảnh báo dashboard khi tệp DB vượt mức này (§13: 1 GB).
	DBWarnBytes = 1 << 30
)

// JanitorJob là một việc dọn: tên để người đọc hiểu, và số dòng sẽ xoá / đã xoá.
type JanitorJob struct {
	// Key: mã ổn định (`telemetry_samples`…), dùng cho giao diện và nhật ký.
	Key string `json:"key"`
	// What: câu tiếng Việt mô tả việc này dọn cái gì.
	What string `json:"what"`
	// Keep: hạn giữ, dạng chữ ("14 ngày").
	Keep string `json:"keep"`
	// Rows: số dòng sẽ xoá (dry-run) hoặc đã xoá (chạy thật).
	Rows int64 `json:"rows"`
	// Err: lỗi nếu có — một việc hỏng không được làm hỏng cả lượt dọn.
	Err string `json:"err,omitempty"`
}

// JanitorResult là kết quả một lượt dọn.
type JanitorResult struct {
	At     int64        `json:"at"`
	DryRun bool         `json:"dry_run"`
	Jobs   []JanitorJob `json:"jobs"`
	Rows   int64        `json:"rows"`
	// Vacuumed: lượt này có chạy VACUUM không, và tiết kiệm được bao nhiêu byte.
	Vacuumed   bool  `json:"vacuumed"`
	FreedBytes int64 `json:"freed_bytes"`
	// Skipped: lý do bỏ qua VACUUM (có người online, chưa tới hạn tuần).
	VacuumNote string `json:"vacuum_note,omitempty"`
}

// JanitorDBs là ba tệp SQLite mà janitor được phép đụng vào.
//
// Truyền vào chứ không tự mở: `cmd/server` đã giữ sẵn kết nối, và mở thêm một kết nối thứ
// hai tới cùng một tệp SQLite trong lúc server đang ghi là cách chắc chắn nhất để gặp
// `database is locked` đúng lúc 04:30 sáng.
type JanitorDBs struct {
	// Main: bảng xếp hạng + mọi bảng `admin_*` + feedback + bans (cùng một tệp, §8).
	Main *sql.DB
	// Telemetry / Metrics: nil khi chạy với -telemetry=false.
	Telemetry *sql.DB
	Metrics   *sql.DB
	// MainPath / TelemetryPath / MetricsPath: đường dẫn tệp, để đo dung lượng và VACUUM.
	MainPath, TelemetryPath, MetricsPath string
	// FeedbackDir: thư mục ảnh góp ý.
	FeedbackDir string
}

// Janitor chạy việc dọn. Không giữ trạng thái ngoài lần chạy gần nhất.
type Janitor struct {
	db JanitorDBs
	// HumansOnline: số người thật đang chơi. VACUUM chỉ chạy khi nó trả 0 (§13).
	HumansOnline func() int
	store        *Store
}

func NewJanitor(st *Store, db JanitorDBs, humans func() int) *Janitor {
	if humans == nil {
		humans = func() int { return 0 }
	}
	return &Janitor{db: db, HumansOnline: humans, store: st}
}

// cut trả về mốc unix của "n ngày trước".
func cut(now time.Time, days int) int64 { return now.AddDate(0, 0, -days).Unix() }

// task mô tả một việc dọn bằng MỘT câu `WHERE`, dùng chung cho đếm và xoá.
type task struct {
	key, what, keep string
	db              *sql.DB
	table           string
	where           string
	args            []any
}

// plan dựng danh sách việc theo đúng bảng §13.
func (j *Janitor) plan(now time.Time) []task {
	var ts []task
	add := func(t task) {
		if t.db != nil {
			ts = append(ts, t)
		}
	}
	// 1. Mẫu telemetry THÔ: 14 ngày. Đã cuộn vào metrics phút/giờ/ngày (§12) nên xoá thô
	//    không mất thông tin — chỉ mất độ phân giải 10 giây của những ngày rất cũ.
	add(task{key: "telemetry_samples", what: "Mẫu telemetry client (10 s/lần)", keep: "14 ngày",
		db: j.db.Telemetry, table: "samples", where: "server_ts < ?", args: []any{cut(now, KeepRawTelemetryDays)}})
	add(task{key: "telemetry_server_samples", what: "Mẫu telemetry server (10 s/lần)", keep: "14 ngày",
		db: j.db.Telemetry, table: "server_samples", where: "ts < ?", args: []any{cut(now, KeepRawTelemetryDays)}})
	// Phiên telemetry không còn mẫu nào thì cũng không còn ai hỏi tới.
	add(task{key: "telemetry_sessions", what: "Phiên telemetry không còn mẫu", keep: "14 ngày",
		db: j.db.Telemetry, table: "sessions",
		where: "started_at < ? AND session_id NOT IN (SELECT DISTINCT session_id FROM samples)",
		args:  []any{cut(now, KeepRawTelemetryDays)}})
	// 2. Lỗi: 30 ngày cho dòng THÔ; chữ ký đã gộp vào `error_sigs` và giữ mãi (xem rollupErrors).
	add(task{key: "telemetry_events", what: "Dòng lỗi thô (chữ ký đã gộp, giữ mãi)", keep: "30 ngày",
		db: j.db.Telemetry, table: "events", where: "server_ts < ?", args: []any{cut(now, KeepErrorsDays)}})
	// 3. Metrics: phút 14 ngày, giờ 180 ngày, ngày giữ mãi (§12.2).
	add(task{key: "metrics_minute", what: "Số liệu hiệu năng theo PHÚT", keep: "14 ngày",
		db: j.db.Metrics, table: "metrics_minute", where: "ts < ?", args: []any{cut(now, 14)}})
	add(task{key: "metrics_hour", what: "Số liệu hiệu năng theo GIỜ", keep: "180 ngày",
		db: j.db.Metrics, table: "metrics_hour", where: "ts < ?", args: []any{cut(now, 180)}})
	// 4. Mạng chơi: 90 ngày, NHƯNG giữ mãi mạng đang nằm trên bảng xếp hạng toàn cục.
	//
	//    `lives_rollup` đã giữ thống kê gộp theo ngày (xem rollupLives), nên xoá dòng chi
	//    tiết không mất đường cong lịch sử — chỉ mất từng mạng một.
	add(task{key: "lives", what: "Mạng chơi cũ (giữ mãi mạng trên bảng xếp hạng)", keep: "90 ngày",
		db: j.db.Main, table: "lives", where: keepBoardLives, args: []any{cut(now, KeepLivesDays)}})
	// 5. Phiên admin hết hạn + lần đăng nhập: 30 ngày.
	add(task{key: "admin_sessions", what: "Phiên đăng nhập quản trị đã hết hạn", keep: "30 ngày",
		db: j.db.Main, table: "admin_sessions", where: "expires_at < ?", args: []any{cut(now, KeepAdminSessDays)}})
	add(task{key: "admin_login_attempts", what: "Nhật ký lần đăng nhập quản trị", keep: "30 ngày",
		db: j.db.Main, table: "admin_login_attempts", where: "at < ?", args: []any{cut(now, KeepAdminSessDays)}})
	// 6. Cấm đã hết hạn: 30 ngày SAU KHI hết. `until_at IS NULL` là cấm vĩnh viễn — không đụng.
	add(task{key: "bans", what: "Lệnh cấm đã hết hạn", keep: "30 ngày sau khi hết",
		db: j.db.Main, table: "bans", where: "until_at IS NOT NULL AND until_at < ?",
		args: []any{cut(now, KeepExpiredBanDays)}})
	// 7. Nhật ký quản trị: 1 năm.
	add(task{key: "admin_audit", what: "Nhật ký thao tác quản trị", keep: "1 năm",
		db: j.db.Main, table: "admin_audit", where: "at < ?", args: []any{cut(now, KeepAuditDays)}})
	return ts
}

// keepBoardLives là mệnh đề `WHERE` của việc dọn `lives`.
//
// Hai lớp bảo vệ, và cả hai đều cần:
//   - `hidden_lives`: mạng đã bị ẩn vẫn là BẰNG CHỨNG (xem store.go), không được xoá.
//   - ba bảng xếp hạng toàn cục: mạng đang hiện trên bảng thì xoá đi là bảng tự nhiên mất
//     một dòng, và không ai hiểu vì sao kỷ lục biến mất.
//
// `boardKeep` = 50, đúng bằng `LIMIT` của `store.Top`: giữ đúng tập mà bảng xếp hạng có
// thể hiện ra, không hơn. Giữ rộng hơn (vd top 1 000) nghe an toàn nhưng nó làm việc dọn
// thành vô nghĩa trên một DB nhỏ — mọi dòng đều lọt vào "top 1 000" và không bao giờ bị xoá.
const keepBoardLives = `ended_at < ?
	AND (guest_id, ended_at) NOT IN (SELECT guest_id, ended_at FROM hidden_lives)
	AND id NOT IN (SELECT id FROM lives ORDER BY score DESC LIMIT 50)
	AND id NOT IN (SELECT id FROM lives ORDER BY hits DESC LIMIT 50)
	AND id NOT IN (SELECT id FROM lives ORDER BY peak_exp DESC LIMIT 50)`

// Plan đếm xem lượt dọn SẼ xoá những gì — không đụng dữ liệu.
func (j *Janitor) Plan(now time.Time) JanitorResult {
	res := JanitorResult{At: now.Unix(), DryRun: true}
	for _, t := range j.plan(now) {
		job := JanitorJob{Key: t.key, What: t.what, Keep: t.keep}
		var n int64
		q := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s`, t.table, t.where)
		if err := t.db.QueryRow(q, t.args...).Scan(&n); err != nil {
			job.Err = err.Error()
		}
		job.Rows = n
		res.Rows += n
		res.Jobs = append(res.Jobs, job)
	}
	res.VacuumNote = j.vacuumNote(now)
	return res
}

// armedKey: cờ "đã được phép xoá thật".
//
// Chừng nào chủ dự án chưa bật, job 04:30 **chỉ chạy bản xem trước** và ghi danh sách vào
// nhật ký. Lý do: một bộ xoá tự động chạy lần đầu lúc 4 giờ sáng, trên dữ liệu thật, mà
// chưa ai nhìn qua danh sách là cách tốt nhất để mất thứ không lấy lại được. Bấm "Dọn ngay"
// một lần là bật — đó mới là một người đọc danh sách rồi đồng ý.
const armedKey = "janitor_armed"

// Armed: job hằng đêm đã được phép xoá thật chưa.
func (j *Janitor) Armed() bool {
	v, ok := j.store.Setting(armedKey)
	return ok && v == "1"
}

// Arm bật phép xoá thật cho các lượt chạy tự động sau này.
func (j *Janitor) Arm() { _ = j.store.SetSetting(armedKey, "1") }

// Run dọn thật. Gộp chữ ký lỗi và thống kê mạng chơi TRƯỚC khi xoá.
func (j *Janitor) Run(now time.Time) JanitorResult {
	res := JanitorResult{At: now.Unix()}
	// Gộp trước, xoá sau — nếu không thì cái giữ mãi bị xoá cùng cái 30 ngày.
	j.rollupErrors(now)
	j.rollupLives(now)
	before := j.totalBytes()
	for _, t := range j.plan(now) {
		job := JanitorJob{Key: t.key, What: t.what, Keep: t.keep}
		q := fmt.Sprintf(`DELETE FROM %s WHERE %s`, t.table, t.where)
		r, err := t.db.Exec(q, t.args...)
		if err != nil {
			job.Err = err.Error()
		} else if n, e := r.RowsAffected(); e == nil {
			job.Rows = n
			res.Rows += n
		}
		res.Jobs = append(res.Jobs, job)
	}
	if note := j.vacuumNote(now); note != "" {
		res.VacuumNote = note
	} else {
		j.vacuumAll()
		j.markVacuum(now)
		res.Vacuumed = true
	}
	res.FreedBytes = before - j.totalBytes()
	return res
}

// vacuumNote trả lý do KHÔNG chạy VACUUM lượt này; rỗng = được chạy.
func (j *Janitor) vacuumNote(now time.Time) string {
	if n := j.HumansOnline(); n > 0 {
		// §13: không chạy khi có người online. VACUUM khoá cả tệp và chép lại toàn bộ —
		// vài giây đứng hình giữa trận là cái giá quá đắt cho vài MB.
		return fmt.Sprintf("bỏ qua VACUUM: đang có %d người chơi", n)
	}
	last := int64(0)
	if v, ok := j.store.Setting("janitor_vacuum_at"); ok {
		last, _ = strconv.ParseInt(v, 10, 64)
	}
	if last > 0 && now.Unix()-last < int64(VacuumEveryDays)*24*3600 {
		return fmt.Sprintf("bỏ qua VACUUM: mới chạy %s trước", time.Duration(now.Unix()-last)*time.Second)
	}
	return ""
}

func (j *Janitor) markVacuum(now time.Time) {
	_ = j.store.SetSetting("janitor_vacuum_at", fmt.Sprint(now.Unix()))
}

func (j *Janitor) vacuumAll() {
	for _, db := range []*sql.DB{j.db.Main, j.db.Telemetry, j.db.Metrics} {
		if db != nil {
			_, _ = db.Exec(`VACUUM`)
		}
	}
}

// rollupErrors gộp dòng lỗi sắp bị xoá thành chữ ký giữ mãi (§13 cột "Sau đó").
func (j *Janitor) rollupErrors(now time.Time) {
	if j.db.Telemetry == nil {
		return
	}
	if _, err := j.db.Telemetry.Exec(errorSigSchema); err != nil {
		return
	}
	// Gộp TRONG GO, không bằng SQL.
	//
	// Chữ ký phải do CHÍNH hàm mà trang "Lỗi" dùng sinh ra (`telemetry.ErrorSig`); viết
	// lại nó bằng SQL là có hai định nghĩa cho cùng một thứ, và tới ngày chúng lệch nhau
	// thì bảng gộp nói một đằng, trang "Lỗi" nói một nẻo. SQLite cũng không có sẵn hàm
	// bỏ-chữ-số nào để làm việc đó.
	rows, err := j.db.Telemetry.Query(
		`SELECT kind, data, ts FROM events WHERE server_ts < ?`, cut(now, KeepErrorsDays))
	if err != nil {
		return
	}
	type agg struct {
		n           int64
		first, last int64
		sample      string
	}
	groups := map[[2]string]*agg{}
	for rows.Next() {
		var kind, data string
		var ts int64
		if rows.Scan(&kind, &data, &ts) != nil {
			continue
		}
		k := [2]string{kind, telemetry.ErrorSig(kind, data)}
		a := groups[k]
		if a == nil {
			a = &agg{first: ts / 1000, last: ts / 1000, sample: data}
			groups[k] = a
		}
		a.n++
		if s := ts / 1000; s < a.first {
			a.first = s
		} else if s > a.last {
			a.last = s
		}
	}
	rows.Close()
	tx, err := j.db.Telemetry.Begin()
	if err != nil {
		return
	}
	for k, a := range groups {
		_, _ = tx.Exec(`INSERT INTO error_sigs (kind, sig, n, first_at, last_at, sample)
			VALUES (?,?,?,?,?,?)
			ON CONFLICT(kind, sig) DO UPDATE SET
				n = n + excluded.n,
				first_at = MIN(first_at, excluded.first_at),
				last_at = MAX(last_at, excluded.last_at)`,
			k[0], k[1], a.n, a.first, a.last, a.sample)
	}
	_ = tx.Commit()
}

// errorSigSchema: bảng chữ ký lỗi giữ mãi.
const errorSigSchema = `
CREATE TABLE IF NOT EXISTS error_sigs (
	kind TEXT NOT NULL,
	sig  TEXT NOT NULL,
	n    INTEGER NOT NULL DEFAULT 0,
	first_at INTEGER NOT NULL,
	last_at  INTEGER NOT NULL,
	sample   TEXT NOT NULL DEFAULT '',
	PRIMARY KEY (kind, sig)
);`

// rollupLives gộp mạng chơi sắp bị xoá thành thống kê theo NGÀY, giữ mãi.
//
// Mất từng mạng một thì chấp nhận được (90 ngày là quá đủ cho mọi câu hỏi về một người cụ
// thể), nhưng mất đường cong "mỗi ngày có bao nhiêu mạng, cảnh giới cao nhất tới đâu" là
// mất khả năng so năm nay với năm ngoái — và không có cách nào dựng lại.
func (j *Janitor) rollupLives(now time.Time) {
	if j.db.Main == nil {
		return
	}
	if _, err := j.db.Main.Exec(livesRollupSchema); err != nil {
		return
	}
	_, _ = j.db.Main.Exec(`
		INSERT INTO lives_rollup (day, n, max_exp, max_score, kills, survive_secs)
		SELECT CAST(ended_at/86400 AS INTEGER), COUNT(*), MAX(peak_exp), MAX(score),
		       SUM(kills), SUM(survive_secs)
		FROM lives WHERE `+keepBoardLives+`
		GROUP BY CAST(ended_at/86400 AS INTEGER)
		ON CONFLICT(day) DO UPDATE SET
			n = n + excluded.n,
			max_exp = MAX(max_exp, excluded.max_exp),
			max_score = MAX(max_score, excluded.max_score),
			kills = kills + excluded.kills,
			survive_secs = survive_secs + excluded.survive_secs`, cut(now, KeepLivesDays))
}

const livesRollupSchema = `
CREATE TABLE IF NOT EXISTS lives_rollup (
	day INTEGER PRIMARY KEY,
	n INTEGER NOT NULL DEFAULT 0,
	max_exp REAL NOT NULL DEFAULT 0,
	max_score REAL NOT NULL DEFAULT 0,
	kills INTEGER NOT NULL DEFAULT 0,
	survive_secs REAL NOT NULL DEFAULT 0
);`

// TableSize là dung lượng một bảng (ước lượng) cho trang "Dọn dữ liệu".
type TableSize struct {
	Name  string `json:"name"`
	Rows  int64  `json:"rows"`
	Bytes int64  `json:"bytes"`
}

// FileSize là một tệp/thư mục có dung lượng đáng theo dõi.
type FileSize struct {
	Name  string `json:"name"`
	Bytes int64  `json:"bytes"`
	Warn  bool   `json:"warn"`
}

// Sizes đo dung lượng từng bảng + từng tệp DB + thư mục ảnh (§13 trang "Dọn dữ liệu").
func (j *Janitor) Sizes() ([]TableSize, []FileSize) {
	var tabs []TableSize
	count := func(db *sql.DB, name string) {
		if db == nil {
			return
		}
		var rows int64
		if err := db.QueryRow(`SELECT COUNT(*) FROM ` + name).Scan(&rows); err != nil {
			return
		}
		// Dung lượng THỰC của một bảng chỉ có qua `dbstat`, mà modernc/sqlite không bật
		// sẵn module đó. Ước lượng theo số dòng × cỡ trung bình đo được trên DB thật: nó
		// đủ để xếp hạng "bảng nào to nhất", và đó là câu hỏi duy nhất trang này cần trả.
		tabs = append(tabs, TableSize{Name: name, Rows: rows, Bytes: rows * rowBytes(name)})
	}
	for _, n := range []string{"lives", "hidden_lives", "feedback", "bans", "admin_audit",
		"admin_sessions", "admin_login_attempts", "likes", "lives_rollup"} {
		count(j.db.Main, n)
	}
	for _, n := range []string{"samples", "events", "sessions", "server_samples", "error_sigs"} {
		count(j.db.Telemetry, n)
	}
	for _, n := range []string{"metrics_minute", "metrics_hour", "metrics_day", "deploy_marks"} {
		count(j.db.Metrics, n)
	}
	sort.Slice(tabs, func(a, b int) bool { return tabs[a].Bytes > tabs[b].Bytes })

	files := []FileSize{
		{Name: "leaderboard.db", Bytes: dbBytes(j.db.MainPath)},
		{Name: "telemetry.db", Bytes: dbBytes(j.db.TelemetryPath)},
		{Name: "metrics.db", Bytes: dbBytes(j.db.MetricsPath)},
		{Name: "ảnh góp ý", Bytes: dirBytes(j.db.FeedbackDir)},
	}
	for i := range files {
		files[i].Warn = files[i].Bytes > DBWarnBytes
	}
	return tabs, files
}

// rowBytes: cỡ trung bình một dòng, đo trên DB thật ngày 12/11. Chỉ để xếp hạng.
func rowBytes(table string) int64 {
	switch table {
	case "samples":
		return 220
	case "events":
		return 400
	case "sessions":
		return 420
	case "server_samples":
		return 90
	case "lives":
		return 160
	case "feedback":
		return 900
	case "metrics_minute", "metrics_hour", "metrics_day":
		return 120
	default:
		return 100
	}
}

// dbBytes cộng cả ba tệp của SQLite: chính, `-wal`, `-shm`.
//
// Bỏ `-wal` là đọc thiếu đúng phần phình nhanh nhất — WAL của telemetry có lúc to hơn cả
// tệp chính giữa hai lần checkpoint.
func dbBytes(path string) int64 {
	if path == "" {
		return 0
	}
	var total int64
	for _, suf := range []string{"", "-wal", "-shm"} {
		if fi, err := os.Stat(path + suf); err == nil {
			total += fi.Size()
		}
	}
	return total
}

func dirBytes(dir string) int64 {
	if dir == "" {
		return 0
	}
	var total int64
	_ = filepath.Walk(dir, func(_ string, fi os.FileInfo, err error) error {
		if err == nil && fi != nil && !fi.IsDir() {
			total += fi.Size()
		}
		return nil
	})
	return total
}

func (j *Janitor) totalBytes() int64 {
	return dbBytes(j.db.MainPath) + dbBytes(j.db.TelemetryPath) + dbBytes(j.db.MetricsPath)
}

// DBWarning trả câu cảnh báo cho dashboard khi tệp DB vượt 1 GB (§13); rỗng = ổn.
func (j *Janitor) DBWarning() string {
	_, files := j.Sizes()
	for _, f := range files {
		if f.Warn {
			return fmt.Sprintf("%s đã %.2f GB — chạy Dọn dữ liệu", f.Name, float64(f.Bytes)/float64(1<<30))
		}
	}
	return ""
}
