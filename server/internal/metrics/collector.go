package metrics

import (
	"log"
	"time"
)

// Collector gom mẫu trong bộ nhớ rồi mỗi phút ghi một lô xuống SQLite (ADMIN_SPEC §12.2).
//
// Luật quan trọng nhất ở đây là **chỉ ghi phút có ít nhất một người thật online**. Bot chạy
// một mình cả đêm không nói lên điều gì về hiệu năng mà người chơi cảm thấy, và 14 ngày dữ
// liệu phút của một server không ai chơi là 14 ngày rác che mất những phút có ý nghĩa.
type Collector struct {
	db    *DB
	build string

	cur     map[key]*Dist
	minute  int64 // mốc đầu phút đang gom (unix)
	humans  bool  // phút này đã thấy người thật chưa
	lastHr  int64
	dropped int
}

type key struct{ metric, scope string }

// New dựng bộ gom. `build` đi vào mọi dòng để biết số này của bản nào.
func New(db *DB, build string) *Collector {
	return &Collector{db: db, build: build, cur: map[key]*Dist{}}
}

// Add ghi một mẫu đơn (gauge: load máy, heap, số người…).
func (c *Collector) Add(metric, scope string, v float64) {
	d := c.at(metric, scope)
	d.Add(v)
}

// AddDist gộp cả một phân phối đã gom sẵn (tick: mọi tick trong 1 giây).
func (c *Collector) AddDist(metric, scope string, src Dist) {
	if src.Count == 0 {
		return
	}
	c.at(metric, scope).Merge(src)
}

func (c *Collector) at(metric, scope string) *Dist {
	k := key{metric, scope}
	d := c.cur[k]
	if d == nil {
		d = &Dist{}
		c.cur[k] = d
	}
	return d
}

// NoteHumans báo số người THẬT đang online lúc lấy mẫu. Phút nào có ≥ 1 thì phút đó được ghi.
func (c *Collector) NoteHumans(n int) {
	if n > 0 {
		c.humans = true
	}
}

// Flush đóng phút hiện tại nếu `now` đã sang phút khác, rồi cuộn giờ/ngày khi sang giờ mới.
// Gọi mỗi giây; tự biết khi nào cần làm gì.
func (c *Collector) Flush(now time.Time) {
	m := now.Truncate(time.Minute).Unix()
	if c.minute == 0 {
		c.minute, c.lastHr = m, now.Truncate(time.Hour).Unix()
		return
	}
	if m == c.minute {
		return
	}
	c.writeMinute(c.minute)
	c.minute, c.cur, c.humans = m, map[key]*Dist{}, false

	if hr := now.Truncate(time.Hour).Unix(); hr != c.lastHr {
		c.lastHr = hr
		if err := c.db.Rollup(now); err != nil {
			log.Printf("metrics: cuộn giờ/ngày: %v", err)
		}
	}
}

// Close ghi nốt phút dở rồi cuộn lần cuối — để 59 giây cuối trước khi tắt server không mất.
func (c *Collector) Close(now time.Time) {
	if c.minute != 0 {
		c.writeMinute(c.minute)
	}
	if err := c.db.Rollup(now); err != nil {
		log.Printf("metrics: cuộn lúc đóng: %v", err)
	}
}

// Dropped: số phút đã bỏ vì không có người thật — trang quản trị in ra để không ai tưởng
// bộ thu gom hỏng khi thấy biểu đồ trống ban đêm.
func (c *Collector) Dropped() int { return c.dropped }

func (c *Collector) writeMinute(ts int64) {
	if len(c.cur) == 0 {
		return
	}
	if !c.humans {
		c.dropped++
		return
	}
	rows := make([]Row, 0, len(c.cur))
	for k, d := range c.cur {
		rows = append(rows, Row{TS: ts, Metric: k.metric, Scope: k.scope, Build: c.build,
			Count: d.Count, Min: d.Min, Max: d.Max, Sum: d.Sum, P95: d.P95(), Hist: d.Hist})
	}
	if err := c.db.WriteMinute(rows); err != nil {
		log.Printf("metrics: ghi phút %d: %v", ts, err)
	}
}
