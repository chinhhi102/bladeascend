package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
)

// dumpSession: toàn bộ sự kiện + mẫu đo của một phiên, để dựng lại lúc lỗi xảy ra.
func dumpSession(id string) {
	db, err := sql.Open("sqlite", "data/telemetry.db?mode=ro")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Printf("\n===== Phiên %s =====\n", id)
	var started, ended int64
	var ua, gpu, quality string
	var qualAuto, automated int
	var dur, fps float64
	var n int
	err = db.QueryRow(`SELECT started_at, ended_at, ua, gpu, quality, qual_auto, automated, dur_secs, fps_median, n_samples
		FROM sessions WHERE session_id = ?`, id).Scan(&started, &ended, &ua, &gpu, &quality, &qualAuto, &automated, &dur, &fps, &n)
	if err != nil {
		fmt.Println("  (không có hàng sessions):", err)
	} else {
		fmt.Printf("  bắt đầu %s · kết thúc %s · dur=%.0fs n=%d fps_med=%.0f q=%s auto_q=%d automated=%d\n",
			time.UnixMilli(started).Format("02/01 15:04:05"), time.UnixMilli(ended).Format("02/01 15:04:05"),
			dur, n, fps, quality, qualAuto, automated)
		fmt.Printf("  ua=%s\n  gpu=%s\n", ua, gpu)
	}

	fmt.Println("  --- sự kiện ---")
	rows, err := db.Query(`SELECT ts, kind, data FROM events WHERE session_id = ? ORDER BY ts`, id)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var ts int64
		var kind, data string
		if err := rows.Scan(&ts, &kind, &data); err != nil {
			log.Fatal(err)
		}
		if len(data) > 200 {
			data = data[:200] + "…"
		}
		fmt.Printf("  %s  %-22s %s\n", time.UnixMilli(ts).Format("15:04:05"), kind, data)
	}
	rows.Close()

	fmt.Println("  --- mẫu đo (samples) ---")
	cols, err := db.Query(`SELECT * FROM samples WHERE session_id = ? ORDER BY ts LIMIT 1`, id)
	if err == nil {
		names, _ := cols.Columns()
		fmt.Println("  cột:", names)
		cols.Close()
	}
	_ = os.Stdout
}
