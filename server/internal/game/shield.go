package game

// Lá chắn vòng kế ngoài cùng (spec §5.2 v4). Vòng ngoài cùng là vùng mài: kiếm địch được
// phép chạm vào đó. Mọi thứ sâu hơn bị bịt kín — vòng kế ngoài còn kiếm thì kiếm địch
// không chạm tới thân được. Mài gãy vòng ngoài thì lá chắn lùi vào 1 lớp; chỉ khi còn
// đúng 1 vòng thì §5.1 mới chém được thân.
//
// Lá chắn là LUẬT CHÉM, không phải tường chắn đường đi. Trước đây nó đẩy người đang đi ra
// xa `reach + shield_r` (hàng trăm đơn vị), nên đi tới gần ai là nhân vật khựng lại giữa
// khoảng không, không thấy gì chặn mà vẫn không tiến được. Giờ hai người đi xuyên qua nhau
// thoải mái; lá chắn chỉ quyết định cú chém có vào tới thân hay không.

// ShieldR is the radius enemy blades may not reach past: the blade tip of the
// SECOND-outermost ring. 0 means no shield left — the player is down to a single ring and
// §5.1 can finally land a body hit.
func (g *Game) ShieldR(p *Player) float64 {
	geo := p.Geom
	if p.Ring.Count <= 0 || g.Cfg.Sword.ShieldRing == "" || geo.Rings < 2 {
		return 0
	}
	return geo.TipDist(geo.Rings-2, p.Ring.Tilt) + geo.Width/2
}

// Shielded reports whether p still has a sealed inner zone, i.e. whether §5.1 body hits on
// p must be ignored this tick.
func (g *Game) Shielded(p *Player) bool { return g.ShieldR(p) > 0 }
