package admin

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"bladeascend/server/internal/telemetry"
)

// Bộ dọn dữ liệu — ADMIN_SPEC §13, checklist cuối §13:
// "dry-run in đúng số dòng sẽ xoá; chạy thật xoá đúng số đó; mạng trên bảng xếp hạng toàn
//
//	cục còn nguyên; DB giảm dung lượng sau VACUUM."
func openTestDBs(t *testing.T) (*Store, JanitorDBs) {
	t.Helper()
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.db")
	main, err := sql.Open("sqlite", mainPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { main.Close() })
	// Bảng `lives` + `hidden_lives` bình thường do `internal/store` dựng; ở đây chỉ cần
	// đúng những cột mà janitor đụng tới.
	if _, err := main.Exec(`
		CREATE TABLE lives (id INTEGER PRIMARY KEY AUTOINCREMENT, guest_id TEXT, name TEXT,
			character TEXT, peak_exp REAL, peak_level INTEGER, kills INTEGER, survive_secs REAL,
			map TEXT, ended_at INTEGER, score REAL DEFAULT 0, hits INTEGER DEFAULT 0);
		CREATE TABLE hidden_lives (guest_id TEXT, ended_at INTEGER, reason TEXT, at INTEGER,
			by_ip TEXT, PRIMARY KEY (guest_id, ended_at));`); err != nil {
		t.Fatal(err)
	}
	st, err := New(main)
	if err != nil {
		t.Fatal(err)
	}
	// Telemetry mở bằng `sql.Open` THÔ, không qua `telemetry.Open`.
	//
	// `telemetry.Open` chạy một goroutine dọn nền, và việc đầu tiên nó làm là xoá mọi thứ
	// quá 30 ngày — tức là xoá đúng những dòng cũ mà bài test vừa chèn vào để thử. Hai bộ
	// dọn cùng chạy trên một tệp thì bài test đo nhầm bộ.
	telPath := filepath.Join(dir, "tel.db")
	tel, err := sql.Open("sqlite", telPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { tel.Close() })
	if _, err := tel.Exec(`
		CREATE TABLE events (id INTEGER PRIMARY KEY AUTOINCREMENT, session_id TEXT NOT NULL,
			ts INTEGER NOT NULL, server_ts INTEGER NOT NULL, kind TEXT NOT NULL, data TEXT NOT NULL DEFAULT '');
		CREATE TABLE samples (id INTEGER PRIMARY KEY AUTOINCREMENT, session_id TEXT NOT NULL, server_ts INTEGER NOT NULL);
		CREATE TABLE sessions (session_id TEXT PRIMARY KEY, started_at INTEGER NOT NULL);
		CREATE TABLE server_samples (id INTEGER PRIMARY KEY AUTOINCREMENT, ts INTEGER NOT NULL);`); err != nil {
		t.Fatal(err)
	}
	return st, JanitorDBs{Main: main, MainPath: mainPath, Telemetry: tel, TelemetryPath: telPath}
}

func TestJanitorDryRunDungBangSoVoiChayThat(t *testing.T) {
	st, dbs := openTestDBs(t)
	now := time.Now()
	old := now.AddDate(0, 0, -200).Unix()
	fresh := now.AddDate(0, 0, -1).Unix()

	// 60 mạng CŨ xếp hạng tăng dần + 1 mạng mới.
	//
	// Luật §13 giữ mãi top 50 của mỗi bảng xếp hạng toàn cục, nên chỉ 10 mạng yếu nhất
	// được phép xoá. Cho cả ba chỉ số (Uy Danh, số chém, Linh khí đỉnh) cùng tăng theo `i`
	// để ba bảng chọn ĐÚNG một tập 50 — nếu không, thứ tự hoà của SQLite là tuỳ ý và con
	// số "sẽ xoá" đổi theo từng lần chạy.
	add := func(guest string, at int64, rank float64) {
		if _, err := dbs.Main.Exec(`INSERT INTO lives (guest_id,name,character,peak_exp,peak_level,
			kills,survive_secs,map,ended_at,score,hits) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
			guest, "N", "a_hon", rank, 3, 1, 60.0, "m", at, rank, int(rank)); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 60; i++ {
		add("cu", old, float64(i))
	}
	add("moi", fresh, 999999) // mạng mới + kỷ lục: phải GIỮ MÃI
	// Nhật ký quá 1 năm + phiên admin hết hạn.
	if _, err := dbs.Main.Exec(`INSERT INTO admin_audit (at,action,target,reason,ip,details)
		VALUES (?,'x','','','',''), (?,'y','','','','')`, now.AddDate(-2, 0, 0).Unix(), fresh); err != nil {
		t.Fatal(err)
	}
	if _, err := dbs.Main.Exec(`INSERT INTO admin_sessions (id,csrf,created_at,expires_at,ip,user_agent)
		VALUES ('a','c',?,?,'','')`, old, old); err != nil {
		t.Fatal(err)
	}

	j := NewJanitor(st, dbs, func() int { return 0 })
	plan := j.Plan(now)
	if !plan.DryRun {
		t.Fatal("Plan phải tự đánh dấu là dry-run")
	}
	byKey := map[string]int64{}
	for _, job := range plan.Jobs {
		if job.Err != "" {
			t.Fatalf("việc %q lỗi: %s", job.Key, job.Err)
		}
		byKey[job.Key] = job.Rows
	}
	// 61 mạng, top 50 được giữ (mạng mới chiếm 1 chỗ trong top đó) → 11 mạng bị xoá.
	if byKey["lives"] != 11 {
		t.Fatalf("dry-run nói xoá %d mạng, muốn 11 (top 50 mỗi bảng và mạng mới phải được giữ)", byKey["lives"])
	}
	if byKey["admin_audit"] != 1 {
		t.Fatalf("nhật ký: dry-run nói %d, muốn 1", byKey["admin_audit"])
	}
	if byKey["admin_sessions"] != 1 {
		t.Fatalf("phiên admin: dry-run nói %d, muốn 1", byKey["admin_sessions"])
	}

	run := j.Run(now)
	for _, job := range run.Jobs {
		if job.Err != "" {
			t.Fatalf("chạy thật, việc %q lỗi: %s", job.Key, job.Err)
		}
		if got, want := job.Rows, byKey[job.Key]; got != want {
			t.Fatalf("việc %q: dry-run nói %d, chạy thật xoá %d", job.Key, want, got)
		}
	}

	// Kỷ lục và mạng mới còn nguyên.
	var left int
	if err := dbs.Main.QueryRow(`SELECT COUNT(*) FROM lives`).Scan(&left); err != nil {
		t.Fatal(err)
	}
	if left != 50 {
		t.Fatalf("còn %d mạng, muốn 50 (đúng tập bảng xếp hạng có thể hiện)", left)
	}
	var keptScore float64
	if err := dbs.Main.QueryRow(`SELECT MAX(score) FROM lives`).Scan(&keptScore); err != nil {
		t.Fatal(err)
	}
	if keptScore != 999999 {
		t.Fatal("mạng đang trên bảng xếp hạng toàn cục bị xoá")
	}
	// Thống kê gộp phải giữ lại dấu vết của những mạng đã xoá.
	var n int64
	if err := dbs.Main.QueryRow(`SELECT COALESCE(SUM(n),0) FROM lives_rollup`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 11 {
		t.Fatalf("thống kê gộp ghi %d mạng, muốn 11", n)
	}
	// Chạy lần hai không được xoá thêm gì (mọi thứ còn lại đều trong hạn).
	again := j.Run(now)
	if again.Rows != 0 {
		t.Fatalf("chạy lần hai còn xoá %d dòng", again.Rows)
	}
}

// Chữ ký lỗi phải được gộp lại TRƯỚC khi xoá dòng thô, và giữ mãi.
func TestJanitorGiuChuKyLoiMai(t *testing.T) {
	st, dbs := openTestDBs(t)
	now := time.Now()
	oldMs := now.AddDate(0, 0, -60).UnixMilli()
	oldSec := now.AddDate(0, 0, -60).Unix()
	for i := 0; i < 7; i++ {
		if _, err := dbs.Telemetry.Exec(`INSERT INTO events (session_id,ts,server_ts,kind,data)
			VALUES ('s',?,?,'error',?)`, oldMs, oldSec,
			"TypeError: x is null\n at f (index-AAA.js:1:2)"); err != nil {
			t.Fatal(err)
		}
	}
	j := NewJanitor(st, dbs, func() int { return 0 })
	if n := j.Plan(now).Jobs; len(n) == 0 {
		t.Fatal("không có việc nào")
	}
	j.Run(now)

	var raw int
	if err := dbs.Telemetry.QueryRow(`SELECT COUNT(*) FROM events`).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	if raw != 0 {
		t.Fatalf("còn %d dòng lỗi thô quá 30 ngày", raw)
	}
	var sig string
	var cnt int64
	if err := dbs.Telemetry.QueryRow(`SELECT sig, n FROM error_sigs`).Scan(&sig, &cnt); err != nil {
		t.Fatalf("chữ ký lỗi phải được giữ lại: %v", err)
	}
	if cnt != 7 {
		t.Fatalf("chữ ký đếm %d lần, muốn 7", cnt)
	}
	if sig != telemetry.ErrorSig("error", "TypeError: x is null\n at f (index-AAA.js:1:2)") {
		t.Fatalf("chữ ký lệch với trang \"Lỗi\": %q", sig)
	}
}

// VACUUM không được chạy khi còn người online (§13).
func TestJanitorKhongVacuumKhiConNguoiChoi(t *testing.T) {
	st, dbs := openTestDBs(t)
	j := NewJanitor(st, dbs, func() int { return 3 })
	res := j.Run(time.Now())
	if res.Vacuumed {
		t.Fatal("VACUUM chạy trong lúc có 3 người đang chơi")
	}
	if res.VacuumNote == "" {
		t.Fatal("phải nói rõ vì sao bỏ qua VACUUM")
	}
	// Không còn ai thì chạy.
	j.HumansOnline = func() int { return 0 }
	if res2 := j.Run(time.Now()); !res2.Vacuumed {
		t.Fatalf("phòng trống mà vẫn bỏ qua VACUUM: %q", res2.VacuumNote)
	}
	// Và mỗi tuần một lần, không phải mỗi lượt dọn.
	if res3 := j.Run(time.Now()); res3.Vacuumed {
		t.Fatal("VACUUM chạy hai lần trong cùng một ngày")
	}
}
