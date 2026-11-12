// Công cụ tạm: in bảng feedback ra màn hình để đọc. Không commit.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "data/leaderboard.db?mode=ro")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(`SELECT id, at, guest_id, name, kind, body, context, image, images, status, note
		FROM feedback ORDER BY at DESC LIMIT 30`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id, at int64
		var guest, name, kind, body, ctx, image, images, status, note string
		if err := rows.Scan(&id, &at, &guest, &name, &kind, &body, &ctx, &image, &images, &status, &note); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("---- #%d  %s  [%s] status=%s\n", id, time.UnixMilli(at).Format("2006-01-02 15:04:05"), kind, status)
		fmt.Printf("  guest=%s name=%q\n", guest, name)
		fmt.Printf("  body: %s\n", body)
		fmt.Printf("  context: %s\n", ctx)
		fmt.Printf("  image=%q images=%q note=%q\n", image, images, note)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	dumpErrors()
	dumpToday()
	dumpSession("9698395a392e7cde")
	dumpSamples("9698395a392e7cde")
}
