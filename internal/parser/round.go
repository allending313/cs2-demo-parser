package parser

import (
	"math"

	demoinfocs "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	common "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
	events "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"

	models "github.com/allending313/cs2-demo-parser/internal/model"
)

type roundCollector struct {
	match *models.Match

	current          *models.Round
	snapshots        []models.Snapshot
	kills            []models.KillEvent
	grenades         []models.GrenadeEvent
	roundStartTick   int
	lastSnapshotTick int
	sampleInterval   int

	// After RoundEnd fires, keep recording snapshots for a short buffer
	// so the viewer doesn't cut off immediately on the last kill.
	pendingEnd     bool
	roundEndTick   int
	postRoundTicks int

	// Running score tallies — more reliable than reading from event state
	ctScore int
	tScore  int

	bombState   string
	bombCarrier uint64

	// In-flight grenades keyed by entity ID. Populated on throw, finalized
	// on the corresponding detonation/destroy event.
	inflight map[int]*inflightGrenade

	// For smokes and infernos the start and expire events fire separately,
	// and the throw/destroy events may not carry the same entity reference.
	// We correlate them by position: map a quantized (x,y) to the grenade
	// index in the grenades slice so the expire handler can patch duration.
	smokeByPos   map[[2]int]int
	infernoByUID map[int64]int
}

func newRoundCollector(match *models.Match) *roundCollector {
	return &roundCollector{
		match:        match,
		inflight:     make(map[int]*inflightGrenade),
		smokeByPos:   make(map[[2]int]int),
		infernoByUID: make(map[int64]int),
	}
}

func (c *roundCollector) onFreezetimeEnd(p demoinfocs.Parser) {
	c.finalizePendingRound()

	gs := p.GameState()
	c.current = &models.Round{
		Number: gs.TotalRoundsPlayed() + 1,
	}
	c.snapshots = nil
	c.kills = nil
	c.grenades = nil
	c.pendingEnd = false
	c.roundStartTick = gs.IngameTick()
	c.lastSnapshotTick = 0

	tickRate := p.TickRate()
	if tickRate > 0 {
		c.sampleInterval = int(math.Round(tickRate / float64(snapshotsPerSecond)))
		c.postRoundTicks = int(tickRate * postRoundBufferSeconds)
	}
	if c.sampleInterval < 1 {
		c.sampleInterval = 13
	}

	c.bombState = ""
	c.bombCarrier = 0
	c.inflight = make(map[int]*inflightGrenade)
	c.smokeByPos = make(map[[2]int]int)
	c.infernoByUID = make(map[int64]int)
}

func (c *roundCollector) onRoundEnd(e events.RoundEnd, p demoinfocs.Parser) {
	if c.current == nil {
		return
	}

	c.current.Winner = teamToString(e.Winner)
	c.current.WinReason = roundEndReasonToString(e.Reason)

	// Track scores ourselves rather than reading from the event's
	// WinnerState/LoserState, more consistent this way
	switch e.Winner {
	case common.TeamCounterTerrorists:
		c.ctScore++
	case common.TeamTerrorists:
		c.tScore++
	}
	c.current.EndCTScore = c.ctScore
	c.current.EndTScore = c.tScore

	// Don't finalize yet — continue capturing frames for the post-round
	c.pendingEnd = true
	c.roundEndTick = p.GameState().IngameTick()
}

// finalizePendingRound commits a round whose post-round buffer has expired.
func (c *roundCollector) finalizePendingRound() {
	if c.current == nil || !c.pendingEnd {
		return
	}

	c.finalizeInflightGrenades()
	c.current.Snapshots = c.snapshots
	c.current.Kills = c.kills
	c.current.Grenades = c.grenades
	c.match.Rounds = append(c.match.Rounds, *c.current)
	c.current = nil
	c.pendingEnd = false
}

func (c *roundCollector) onKill(e events.Kill, p demoinfocs.Parser) {
	if c.current == nil {
		return
	}

	gs := p.GameState()
	tick := gs.IngameTick()
	kill := models.KillEvent{
		Tick:        tick,
		TimeInRound: c.ticksToSeconds(tick, p),
		Headshot:    e.IsHeadshot,
		Wallbang:    e.PenetratedObjects > 0,
	}

	if e.Weapon != nil {
		kill.Weapon = e.Weapon.String()
	}

	if e.Killer != nil {
		kill.Attacker = e.Killer.SteamID64
		pos := e.Killer.Position()
		kill.AttackerX = pos.X
		kill.AttackerY = pos.Y
	}

	if e.Victim != nil {
		kill.Victim = e.Victim.SteamID64
		pos := e.Victim.Position()
		kill.VictimX = pos.X
		kill.VictimY = pos.Y
	}

	c.kills = append(c.kills, kill)
}

func (c *roundCollector) onFrame(p demoinfocs.Parser) {
	if c.current == nil {
		return
	}

	gs := p.GameState()
	tick := gs.IngameTick()

	// Cut off snapshot collection once the post-round buffer expires
	if c.pendingEnd && (tick-c.roundEndTick) > c.postRoundTicks {
		c.finalizePendingRound()
		return
	}

	if tick-c.lastSnapshotTick < c.sampleInterval {
		return
	}
	c.lastSnapshotTick = tick

	snapshot := models.Snapshot{
		Tick:        tick,
		TimeInRound: c.ticksToSeconds(tick, p),
	}

	snapshot.Bomb = c.captureBombState(gs)

	for _, player := range gs.Participants().Playing() {
		if player == nil {
			continue
		}

		pos := player.Position()
		ps := models.PlayerState{
			SteamID:    player.SteamID64,
			Name:       player.Name,
			Team:       teamToString(player.Team),
			X:          pos.X,
			Y:          pos.Y,
			Z:          pos.Z,
			Yaw:        float64(player.ViewDirectionX()),
			HP:         player.Health(),
			Armor:      player.Armor(),
			IsAlive:    player.IsAlive(),
			HasDefuser: player.HasDefuseKit(),
		}

		if w := player.ActiveWeapon(); w != nil {
			ps.Weapon = w.String()
		}

		remaining := player.FlashDurationTimeRemaining().Seconds()
		if remaining > 0 {
			ps.FlashAlpha = math.Min(remaining/5.0*255, 255)
		}

		snapshot.Players = append(snapshot.Players, ps)
	}

	c.snapshots = append(c.snapshots, snapshot)
	c.sampleGrenadePositions(gs, p)
}

func (c *roundCollector) captureBombState(gs demoinfocs.GameState) *models.BombState {
	bomb := gs.Bomb()
	if bomb.Carrier != nil {
		return &models.BombState{
			X:       bomb.Carrier.Position().X,
			Y:       bomb.Carrier.Position().Y,
			State:   "carried",
			Carrier: bomb.Carrier.SteamID64,
		}
	}

	state := c.bombState
	if state == "" {
		state = "dropped"
	}

	pos := bomb.LastOnGroundPosition
	return &models.BombState{
		X:     pos.X,
		Y:     pos.Y,
		State: state,
	}
}

func (c *roundCollector) flush() {
	if c.current == nil {
		return
	}

	c.current.Snapshots = c.snapshots
	c.current.Kills = c.kills
	c.match.Rounds = append(c.match.Rounds, *c.current)
	c.current = nil
}

func (c *roundCollector) ticksToSeconds(tick int, p demoinfocs.Parser) float64 {
	elapsed := tick - c.roundStartTick
	tickRate := p.TickRate()
	if tickRate <= 0 {
		return 0
	}
	return float64(elapsed) / tickRate
}
