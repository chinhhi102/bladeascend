package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// dumpErrors in các sự kiện lỗi client gần đây trong telemetry.db.
func dumpErrors() {
	db, err := sql.Open("sqlite", "data/telemetry.db?mode=ro")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("===== Sự kiện theo loại (14 ngày) =====")
	since := time.Now().Add(-14 * 24 * time.Hour).UnixMilli()
	rows, err := db.Query(`SELECT kind, COUNT(*), MAX(ts) FROM events WHERE ts >= ? GROUP BY kind ORDER BY 2 DESC`, since)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var kind string
		var n int
		var maxts int64
		if err := rows.Scan(&kind, &n, &maxts); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  %-24s %6d  lần cuối %s\n", kind, n, time.UnixMilli(maxts).Format("02/01 15:04"))
	}
	rows.Close()

	fmt.Println("\n===== Lỗi client gần nhất =====")
	rows, err = db.Query(`SELECT e.ts, e.kind, e.data, e.session_id, s.browser, s.os, s.build
		FROM events e LEFT JOIN sessions s ON s.session_id = e.session_id
		WHERE e.kind IN ('error','webgl_context_lost','gpu_pressure') ORDER BY e.ts DESC LIMIT 40`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var ts int64
		var kind, data, sess string
		var browser, os, build sql.NullString
		if err := rows.Scan(&ts, &kind, &data, &sess, &browser, &os, &build); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("---- %s  [%s]  sess=%s  %s/%s build=%s\n",
			time.UnixMilli(ts).Format("02/01 15:04:05"), kind, sess, browser.String, os.String, build.String)
		fmt.Printf("     %s\n", data)
	}
}
