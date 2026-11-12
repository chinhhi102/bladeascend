package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// dumpSamples in các mẫu đo quanh lúc lỗi.
func dumpSamples(id string) {
	db, err := sql.Open("sqlite", "data/telemetry.db?mode=ro")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM samples WHERE session_id = ?`, id).Scan(&n)
	fmt.Printf("\n===== %d mẫu của phiên %s =====\n", n, id)
	rows, err := db.Query(`SELECT ts, fps_avg, fps_p1, frame_p95, frame_max, over33, long_tasks, draw_calls,
		sprites, tex_mb, tex_uploads, zone, realm, swords, chars_view, tien_view, swords_draw, vfx, zoning, stalls
		FROM samples WHERE session_id = ? ORDER BY ts DESC LIMIT 25`, id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	fmt.Printf("%-9s %5s %5s %6s %6s %5s %5s %6s %6s %6s %5s %4s %5s %5s %5s %4s %5s %4s %5s\n",
		"giờ", "fps", "p1", "f_p95", "f_max", ">33", "task", "draw", "sprite", "texMB", "upl", "khu", "cảnh", "kiếm", "ngươi", "tiên", "vẽ_k", "vfx", "đổikhu")
	for rows.Next() {
		var ts int64
		var fps, fp1, fp95, fmax, texmb float64
		var over33, tasks, draws, sprites, upl, zone, realm, swords, chars, tien, swd, vfx, zoning, stalls sql.NullInt64
		if err := rows.Scan(&ts, &fps, &fp1, &fp95, &fmax, &over33, &tasks, &draws, &sprites, &texmb, &upl,
			&zone, &realm, &swords, &chars, &tien, &swd, &vfx, &zoning, &stalls); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%-9s %5.0f %5.0f %6.1f %6.1f %5d %5d %6d %6d %6.0f %5d %4d %5d %5d %5d %4d %5d %4d %5d\n",
			time.UnixMilli(ts).Format("15:04:05"), fps, fp1, fp95, fmax, over33.Int64, tasks.Int64, draws.Int64,
			sprites.Int64, texmb, upl.Int64, zone.Int64, realm.Int64, swords.Int64, chars.Int64, tien.Int64,
			swd.Int64, vfx.Int64, zoning.Int64)
	}
}
