package main

// VersusGame owns two linked Attack-mode Game instances for head-to-head play.
// Each cleared match sends penalty balls to the opponent (floor(cleared/3) balls).
type VersusGame struct {
	P1 *Game
	P2 *Game
}

// NewVersusGame creates two linked games and registers them with the hub.
func NewVersusGame(hub *Hub) *VersusGame {
	p1 := &Game{
		grid:      NewAttackGrid(),
		mode:      modeVersus,
		playerID:  1,
		dropTicks: attackDropInterval,
		hub:       hub,
		moveIn:    hub.MoveIn,
	}
	p2 := &Game{
		grid:      NewAttackGrid(),
		mode:      modeVersus,
		playerID:  2,
		dropTicks: attackDropInterval,
		hub:       hub,
		moveIn:    hub.MoveIn2,
	}
	p1.opponent = p2
	p2.opponent = p1
	hub.SetMode("versus")
	return &VersusGame{P1: p1, P2: p2}
}

func (v *VersusGame) LogicTick() {
	v.P1.LogicTick()
	v.P2.LogicTick()
}

func (v *VersusGame) BroadcastInitial() {
	v.P1.BroadcastInitial()
	v.P2.BroadcastInitial()
}

// Reset reinitializes both boards and clears penalty queues.
func (v *VersusGame) Reset() {
	v.P1.Reset(modeVersus)
	v.P2.Reset(modeVersus)
}
