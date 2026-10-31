// Command hidelives ẩn khỏi bảng xếp hạng những mạng chơi đã được DUYỆT trong
// `docs/LEADERBOARD_AUDIT.md`.
//
//	cd server && go run ./cmd/hidelives            # xem trước, KHÔNG đổi gì
//	cd server && go run ./cmd/hidelives --apply    # ẩn thật
//
// ── Vì sao có công cụ này thay vì bấm nút ở trang quản trị ──────────────────────────────
//
// 42 lần bấm tay là 42 cơ hội bấm nhầm một dòng, và không có cách nào kiểm lại đã bấm đúng
// những dòng nào. Công cụ này TÍNH LẠI danh sách từ cùng một bộ ngưỡng mỗi lần chạy, nên
// `--apply` hai lần cho ra cùng một kết quả, và `--dry-run` cho xem trước chính xác thứ sẽ
// xảy ra.
//
// Nó đi qua ĐÚNG đường mà nút bấm đi: `store.Hide` + một dòng `admin_audit`. Không có
// đường tắt nào chạm thẳng vào SQL.
//
// **Chỉ ẩn. Không xoá gì, không cấm ai.** Xem `internal/store/hidden.go`.
package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"time"

	"kiemxoay/server/internal/admin"
	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/store"
)

// Ngưỡng — lấy từ `cmd/sim --bots 40 --minutes 20` trên đúng luật đang chạy:
// Đại Thừa 429 s, Chân Tiên 479 s. Lấy MỘT NỬA làm ngưỡng, rộng tay để người chơi giỏi gặp
// may không bị gắn oan. Giống hệt bảng `/admin/#board` dùng.
const (
	simDaiThuaSecs  = 429.1
	simChanTienSecs = 478.7
)

// lanID: `guest_id` do client sinh khi KHÔNG ở secure context — tức là vào bằng địa chỉ
// LAN qua HTTP thường. `crypto.randomUUID` chỉ chạy ở secure context, nên dạng này phân
// biệt được người chơi thật với máy chủ mà không phải đoán theo tên.
var lanID = regexp.MustCompile(`^\d{13}-[0-9a-f]+$`)

// keepGuest: mạng chơi chủ dự án quyết GIỮ LẠI (LEADERBOARD_AUDIT §4).
//
// "Người chơi A" đạt tầng 100 trong 164 giây — nhanh bất thường, nhưng có 15 đòn đánh
// ra và 5 lần hạ người. Ngờ không phải bằng chứng.
const keepGuest = "guest-a1b2c3 (đã ẩn)"

const reason = "Duyệt theo LEADERBOARD_AUDIT 31/10 — lệnh test khi cửa debug còn mở"

func suspectWhy(level int, secs float64) string {
	if secs <= 0 {
		secs = 0.1
	}
	switch {
	case level >= 90 && secs < simChanTienSecs/2:
		return fmt.Sprintf("Tiên giới (tầng %d) trong %.0fs, chơi thật cần ~%.0fs", level, secs, simChanTienSecs)
	case level >= 80 && secs < simDaiThuaSecs/2:
		return fmt.Sprintf("Đại Thừa (tầng %d) trong %.0fs, chơi thật cần ~%.0fs", level, secs, simDaiThuaSecs)
	}
	return ""
}

func main() {
	dbPath := flag.String("db", "", "file SQLite (rỗng = server/data/leaderboard.db)")
	apply := flag.Bool("apply", false, "ẩn THẬT; không có cờ này thì chỉ xem trước")
	flag.Parse()

	path := *dbPath
	if path == "" {
		path = config.FindRepoFile("server/data/leaderboard.db")
	}
	st, err := store.Open(path, time.Second)
	if err != nil {
		log.Fatalf("mở %s: %v", path, err)
	}
	defer st.Close()
	adminStore, err := admin.New(st.DB())
	if err != nil {
		log.Fatal(err)
	}

	rows, err := st.RecentLives(1000)
	if err != nil {
		log.Fatal(err)
	}
	now := time.Now()
	var local, lan, kept, already int
	for _, r := range rows {
		why := suspectWhy(r.Level, r.Survive)
		if why == "" {
			continue
		}
		if r.GuestID == keepGuest {
			kept++
			fmt.Printf("  GIỮ    %s  %s — %s\n", time.Unix(r.EndedAt, 0).Format("02/01 15:04"), r.Name, why)
			continue
		}
		if r.Hidden {
			already++
			continue
		}
		where := "LOCAL"
		if lanID.MatchString(r.GuestID) {
			where, lan = "LAN  ", lan+1
		} else {
			local++
		}
		fmt.Printf("  ẩn %s %s  %-18s tầng %3d  %5.0fs  %s\n",
			where, time.Unix(r.EndedAt, 0).Format("02/01 15:04"), r.Name, r.Level, r.Survive, why)
		if !*apply {
			continue
		}
		if err := st.Hide(r.GuestID, r.EndedAt, reason, "cmd/hidelives", now); err != nil {
			log.Fatalf("ẩn %s/%d: %v", r.GuestID, r.EndedAt, err)
		}
		adminStore.Audit("leaderboard_hide", r.GuestID, reason, "cmd/hidelives",
			map[string]any{"ended_at": r.EndedAt, "name": r.Name, "level": r.Level,
				"survive_secs": r.Survive, "why": why, "source": where})
	}

	fmt.Printf("\n%d mạng LOCAL + %d mạng LAN = %d cần ẩn · %d giữ lại · %d đã ẩn từ trước\n",
		local, lan, local+lan, kept, already)
	if !*apply {
		fmt.Println("\nMới chỉ XEM TRƯỚC. Chạy lại với --apply để ẩn thật.")
		return
	}
	fmt.Println("\nĐã ẩn. Dữ liệu CÒN NGUYÊN — bỏ ẩn bất cứ lúc nào ở /admin/#board.")
}
