package game

// EmitBotLeave phát sự kiện bot rời phòng do admin rút (spec §9.5) — client vẽ làn khói ở
// chỗ nó đứng, để người đang nhìn không thấy một nhân vật biến mất trơn.
func (g *Game) EmitBotLeave(p *Player) {
	g.emit(Event{Kind: "BotLeave", A: p.ID, X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}
