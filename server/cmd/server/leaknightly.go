package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Kết quả bài kiểm rò rỉ HẰNG ĐÊM — GAME_SPEC §13.9 (v6.54).
//
// `tools/leaknightly.sh` chạy lúc 03:00 và ghi một tệp JSON; server chỉ ĐỌC nó. Vì sao
// không cho server tự chạy bài đo: bài client cần Chrome headless và Playwright, tức là
// cần cả một môi trường Node — nhét thứ đó vào tiến trình đang phục vụ người chơi là đổi
// bản chất của tiến trình đó. Một tệp JSON là ranh giới đủ rộng và đủ rõ.
//
// Việc duy nhất server GHI vào tệp là cờ `ack`: "đã có người xem lại". Đó phải là hành
// động của một người đang đăng nhập trang quản trị, không phải một biến môi trường gõ cho
// qua — nếu không thì cổng chặn deploy chỉ là một chướng ngại cần gõ thêm một chữ.

// LeakNightly là nội dung `server/data/leak-nightly.json`.
type LeakNightly struct {
	At int64 `json:"at"`
	OK bool  `json:"ok"`
	// Verdict: chữ của §13.9 v6.57 — "xanh" · "vang" (cảnh báo, KHÔNG chặn) · "do" (chặn
	// deploy) · "chua_ket_luan". `OK` vẫn là thứ `tools/deploy.sh` đọc để chặn, nên vàng
	// giữ OK = true; `Verdict` chỉ nói rõ thêm vàng khác xanh ở chỗ nào.
	Verdict string `json:"verdict,omitempty"`
	// Skipped: lý do bỏ qua đêm đó (có người chơi, máy bận). Rỗng = đã chạy thật.
	Skipped string `json:"skipped"`
	Note    string `json:"note"`
	Log     string `json:"log"`
	// Ack: đã có người xem lại kết quả đỏ. Chỉ trang quản trị đặt được.
	Ack           bool   `json:"ack"`
	AckAt         int64  `json:"ack_at,omitempty"`
	ClientMinutes int    `json:"client_minutes"`
	ServerMinutes int    `json:"server_minutes"`
	Path          string `json:"-"`
	// Client: số liệu chi tiết của bài client, đọc từ `leak-client.json` (do
	// `tools/browsertest/leakcheck.mjs` ghi). Tách tệp vì script bash không ghép JSON được
	// mà không kéo theo `jq`; ghép ở đây thì trang quản trị vẫn thấy một khối duy nhất.
	Client *LeakClient `json:"client,omitempty"`
}

// LeakClient là nội dung `server/data/leak-client.json` — phép xét theo dốc của §13.9 v6.57.
//
// Trang quản trị phải hiện ĐỦ năm thứ spec đòi: 5 dốc khối, dốc trung bình, KTC, mốc kỷ
// lục, và chuỗi dấu 7 đêm gần nhất. Một dòng "ĐỎ/XANH" không nói được build nào bắt đầu
// trượt, mà đó là câu hỏi đầu tiên mỗi lần cổng đỏ.
type LeakClient struct {
	At      int64  `json:"at"`
	Build   string `json:"build"`
	Verdict string `json:"verdict"`
	// Scenario: phiên bản KỊCH BẢN (v6.58). Mốc kỷ lục chỉ có nghĩa trong phạm vi một
	// kịch bản, nên trang quản trị phải hiện nó cạnh con số — không thì hai đêm chạy hai
	// kịch bản khác nhau trông như hai đêm so được với nhau.
	ScenarioVersion string   `json:"scenario_version"`
	Reasons         []string `json:"reasons"`
	Minutes         float64  `json:"minutes"`
	Warmup          float64  `json:"warmup_min"`
	Blocks          int      `json:"blocks"`
	Note            string   `json:"note,omitempty"`

	Scenario struct {
		ZoneSwitches       int `json:"zone_switches"`
		ZoneSwitchesWindow int `json:"zone_switches_window"`
		Deaths             int `json:"deaths"`
		DeathsWindow       int `json:"deaths_window"`
		EndZone            int `json:"end_zone"`
		Cycles             int `json:"cycles"`
	} `json:"scenario"`

	Metrics map[string]struct {
		Blocks   []float64 `json:"blocks"`
		Mean     float64   `json:"mean"`
		CILo     float64   `json:"ci_lo"`
		CIHi     float64   `json:"ci_hi"`
		Baseline float64   `json:"baseline"`
		Rise60   float64   `json:"rise60"`
		RisePct  float64   `json:"rise_pct"`
		Red      bool      `json:"red"`
	} `json:"metrics"`

	Ratchet map[string]struct {
		Record      float64 `json:"record"`
		Value       float64 `json:"value"`
		ExceededPct float64 `json:"exceeded_pct"`
		Gate        bool    `json:"gate"`
		Lowered     bool    `json:"lowered"`
		First       bool    `json:"first"`
	} `json:"ratchet"`

	Signs map[string]struct {
		Signs  []string `json:"signs"`
		Streak int      `json:"streak"`
	} `json:"signs"`

	Ledger []struct {
		Key    string  `json:"key"`
		Slope  float64 `json:"slope"`
		CILo   float64 `json:"ci_lo"`
		CIHi   float64 `json:"ci_hi"`
		Mean   int     `json:"mean"`
		Rising bool    `json:"rising"`
	} `json:"ledger"`

	GPU struct {
		PeakMB   float64 `json:"peak_mb"`
		BudgetMB float64 `json:"budget_mb"`
	} `json:"gpu"`
}

// Stale: kết quả cũ hơn 48 giờ. Bài chạy hằng đêm, nên quá hai đêm không có kết quả nghĩa
// là cron hỏng — và một cổng im lặng ngừng chạy còn tệ hơn một cổng đỏ.
func (l *LeakNightly) Stale(now time.Time) bool {
	return l.At > 0 && now.Unix()-l.At > 48*3600
}

// Warning là câu cảnh báo cho dashboard; rỗng = không có gì phải nói.
func (l *LeakNightly) Warning(now time.Time) string {
	switch {
	case l == nil:
		return ""
	case !l.OK && !l.Ack:
		return "Bài kiểm rò rỉ bộ nhớ đêm qua ĐỎ: " + l.Note + " — deploy kế tiếp bị chặn tới khi xem lại."
	case l.Stale(now):
		return "Bài kiểm rò rỉ chưa chạy quá 48 giờ — kiểm lại mục cron (tools/leaknightly.sh --install)."
	case l.Verdict == "vang":
		// Thẻ VÀNG của v6.57: dốc dương 3 đêm liên tiếp trong khi KTC từng đêm vẫn ôm qua
		// 0. Không chặn deploy — nhưng phải hiện ra, vì đây là lưới duy nhất bắt được rò
		// rỉ chậm dưới 0,3 MB/phút, và đêm thứ tư còn dương là chuyển ĐỎ.
		return "Bài kiểm rò rỉ VÀNG: dốc bộ nhớ dương 3 đêm liên tiếp — mở việc điều tra. Đêm thứ 4 vẫn dương sẽ chuyển ĐỎ (không chặn deploy lúc này)."
	}
	return ""
}

func leakNightlyPath(dbPath string) string {
	return filepath.Join(filepath.Dir(dbPath), "leak-nightly.json")
}

// readLeakNightly đọc tệp kết quả. Trả nil khi chưa chạy lần nào hoặc tệp hỏng.
//
// Tệp hỏng cũng trả nil chứ không báo lỗi: nó được một script shell ghi, và một lần ghi
// dở dang không được phép làm trang quản trị không mở được.
func readLeakNightly(dbPath string) *LeakNightly {
	p := leakNightlyPath(dbPath)
	b, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var l LeakNightly
	if json.Unmarshal(b, &l) != nil {
		return nil
	}
	l.Path = p
	// Số liệu chi tiết của bài client nằm ở tệp bên cạnh. Thiếu hoặc hỏng thì bỏ qua: thẻ
	// cảnh báo vẫn phải hiện được dù không có bảng số.
	if cb, err := os.ReadFile(filepath.Join(filepath.Dir(p), "leak-client.json")); err == nil {
		var c LeakClient
		if json.Unmarshal(cb, &c) == nil {
			l.Client = &c
		}
	}
	return &l
}

// handleAdminLeakAck đánh dấu "đã xem lại" để deploy kế tiếp đi tiếp được.
func (s *server) handleAdminLeakAck(w http.ResponseWriter, r *http.Request) {
	l := readLeakNightly(s.dbPath)
	if l == nil {
		http.Error(w, "chưa có kết quả nào", http.StatusNotFound)
		return
	}
	l.Ack, l.AckAt = true, time.Now().Unix()
	// Ghi lại đúng những gì `leaknightly.sh` ghi, cộng cờ ack. Số liệu chi tiết của bài
	// client sống ở tệp riêng của nó — chép sang đây là tạo ra bản thứ hai của cùng một
	// con số, và hai bản thì sớm muộn lệch nhau.
	out := *l
	out.Client = nil
	b, err := json.MarshalIndent(out, "", "  ")
	if err == nil {
		// Ghi qua tệp tạm rồi đổi tên: `tools/deploy.sh` có thể đọc đúng lúc này, và đọc
		// một tệp JSON mới viết được một nửa thì nó coi như không có cổng.
		tmp := l.Path + ".tmp"
		if err = os.WriteFile(tmp, b, 0o644); err == nil {
			err = os.Rename(tmp, l.Path)
		}
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.adminAuth.Store().Audit("leak_ack", "", "xác nhận đã xem kết quả rò rỉ đỏ", clientIP(r),
		map[string]any{"at": l.At, "note": l.Note})
	log.Printf("rò rỉ hằng đêm: đã xác nhận xem lại (%s)", l.Note)
	writeJSON(w, l)
}
