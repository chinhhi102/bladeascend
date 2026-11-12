package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// dumpToday: mọi sự kiện lỗi + phiên trong 2 ngày gần nhất.
func dumpToday() {
	db, err := sql.Open("sqlite", "data/telemetry.db?mode=ro")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	since := time.Now().Add(-48 * time.Hour).UnixMilli()

	fmt.Println("\n===== Lỗi 48 giờ qua (đầy đủ) =====")
	rows, err := db.Query(`SELECT e.ts, e.kind, e.data, e.session_id, s.browser, s.os, s.mobile, s.build,
		s.gpu, s.quality, s.automated, s.fps_median, s.dur_secs
		FROM events e LEFT JOIN sessions s ON s.session_id = e.session_id
		WHERE e.kind IN ('error','webgl_context_lost','gpu_pressure') AND e.ts >= ?
		ORDER BY e.ts DESC`, since)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var ts int64
		var kind, data, sess string
		var browser, os, build, gpu, quality sql.NullString
		var mobile, automated sql.NullInt64
		var fps, dur sql.NullFloat64
		if err := rows.Scan(&ts, &kind, &data, &sess, &browser, &os, &mobile, &build, &gpu, &quality, &automated, &fps, &dur); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("---- %s [%s] sess=%s %s/%s mobile=%d auto=%d build=%s q=%s fps_med=%.0f dur=%.0fs\n     gpu=%s\n     %s\n",
			time.UnixMilli(ts).Format("02/01 15:04:05"), kind, sess, browser.String, os.String,
			mobile.Int64, automated.Int64, build.String, quality.String, fps.Float64, dur.Float64, gpu.String, data)
	}
}
