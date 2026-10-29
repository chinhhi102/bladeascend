package game

// Luật chém thân theo số vòng (spec v4.5 §5.2). Hai người đi XUYÊN QUA vòng kiếm của nhau
// tự do, không bị chặn, không bật ngược. Kiếm đối thủ chỉ chém được vào thân khi nạn nhân
// còn tối đa `body_hit_max_rings` vòng (mặc định 1, tức ≤ 10 kiếm). Còn từ 2 vòng trở lên
// thì thân "được che" — phải mài rụng vòng ngoài trước.
//
// Bản trước hiểu đây là tường chắn đường đi và đẩy người đang đi ra xa hàng trăm đơn vị, nên
// đi tới gần ai là nhân vật khựng lại giữa khoảng không. Giờ nó thuần là luật chém.

// Shielded reports whether p still has enough rings to keep blades off their body, i.e.
// whether §5.1 body hits on p must be ignored this tick.
func (g *Game) Shielded(p *Player) bool {
	return p.Ring.Count > 0 && p.Geom.Rings > max(g.Cfg.Sword.BodyHitMaxRings, 1)
}

// ShieldR is how far out p's body is still covered: the blade tip of the last ring that is
// not reachable yet. 0 once §5.1 can land. Bot AI dùng để biết đứng mài ở đâu.
func (g *Game) ShieldR(p *Player) float64 {
	if !g.Shielded(p) {
		return 0
	}
	geo := p.Geom
	return geo.TipDist(geo.Rings-2, p.Ring.Tilt) + geo.Width/2
}
