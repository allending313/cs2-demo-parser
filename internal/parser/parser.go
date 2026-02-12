package parser

import (
	"errors"
	"fmt"
	"math"
	"os"

	demoinfocs "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	common "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
	events "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
	msg "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/msg"

	models "github.com/allending313/cs2-demo-parser/internal/model"
)

const (
	snapshotsPerSecond     = 5
	postRoundBufferSeconds = 3.0
)

// ProgressFunc is called periodically with a value between 0 and 1.
type ProgressFunc func(progress float32)

func ParseDemo(filePath, matchID string, onProgress ProgressFunc) (*models.Match, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening demo: %w", err)
	}
	defer f.Close()

	p := demoinfocs.NewParser(f)
	defer p.Close()

	match := &models.Match{ID: matchID}
	collector := newRoundCollector(match)

	p.RegisterNetMessageHandler(func(srvInfo *msg.CSVCMsg_ServerInfo) {
		match.Map = srvInfo.GetMapName()
	})

	// Start capturing after freeze time, not at round start.
	// RoundStart still fires during freeze time
	// We only use it to finalize any lingering post-round buffer from the previous round.
	p.RegisterEventHandler(func(e events.RoundStart) {
		collector.finalizePendingRound()
	})

	p.RegisterEventHandler(func(e events.RoundFreezetimeEnd) {
		collector.onFreezetimeEnd(p)
	})

	p.RegisterEventHandler(func(e events.RoundEnd) {
		collector.onRoundEnd(e, p)
	})

	p.RegisterEventHandler(func(e events.Kill) {
		collector.onKill(e, p)
	})

	p.RegisterEventHandler(func(e events.BombPlanted) {
		collector.bombState = "planted"
		collector.bombCarrier = 0
	})

	p.RegisterEventHandler(func(e events.BombDefused) {
		collector.bombState = "defused"
	})

	p.RegisterEventHandler(func(e events.BombExplode) {
		collector.bombState = "exploded"
	})

	p.RegisterEventHandler(func(e events.BombPickup) {
		if e.Player != nil {
			collector.bombState = "carried"
			collector.bombCarrier = e.Player.SteamID64
		}
	})

	p.RegisterEventHandler(func(e events.BombDropped) {
		collector.bombState = "dropped"
		collector.bombCarrier = 0
	})

	p.RegisterEventHandler(func(e events.FrameDone) {
		collector.onFrame(p)

		if onProgress != nil {
			onProgress(p.Progress())
		}
	})

	if err := p.ParseToEnd(); err != nil {
		if !errors.Is(err, demoinfocs.ErrUnexpectedEndOfDemo) {
			return nil, fmt.Errorf("parsing demo: %w", err)
		}
	}

	collector.finalizePendingRound()
	collector.flush()

	match.TickRate = p.TickRate()
	match.Duration = p.CurrentTime().Seconds()
	match.Teams = buildTeams(match.Rounds)

	return match, nil
}

type roundCollector struct {
	match *models.Match

	current          *models.Round
	snapshots        []models.Snapshot
	kills            []models.KillEvent
	roundStartTick   int
	lastSnapshotTick int
	sampleInterval   int

	// After RoundEnd fires, keep recording snapshots for a short buffer
	// so the viewer doesn't cut off immediately on the last kill.
	pendingEnd     bool
	roundEndTick   int
	postRoundTicks int

	bombState   string
	bombCarrier uint64
}

func newRoundCollector(match *models.Match) *roundCollector {
	return &roundCollector{match: match}
}

func (c *roundCollector) onFreezetimeEnd(p demoinfocs.Parser) {
	c.finalizePendingRound()

	gs := p.GameState()
	c.current = &models.Round{
		Number: gs.TotalRoundsPlayed() + 1,
	}
	c.snapshots = nil
	c.kills = nil
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
}

func (c *roundCollector) onRoundEnd(e events.RoundEnd, p demoinfocs.Parser) {
	if c.current == nil {
		return
	}

	c.current.Winner = teamToString(e.Winner)
	c.current.WinReason = roundEndReasonToString(e.Reason)

	if e.WinnerState != nil {
		score := e.WinnerState.Score() + 1 // score hasn't updated yet so we add 1
		if e.Winner == common.TeamCounterTerrorists {
			c.current.EndCTScore = score
			if e.LoserState != nil {
				c.current.EndTScore = e.LoserState.Score()
			}
		} else {
			c.current.EndTScore = score
			if e.LoserState != nil {
				c.current.EndCTScore = e.LoserState.Score()
			}
		}
	}

	c.pendingEnd = true
	c.roundEndTick = p.GameState().IngameTick()
}

func (c *roundCollector) finalizePendingRound() {
	if c.current == nil || !c.pendingEnd {
		return
	}

	c.current.Snapshots = c.snapshots
	c.current.Kills = c.kills
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

// flush saves any in-progress round that never received a RoundEnd (demo ends mid-round).
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

// buildTeams constructs team rosters from snapshot data across all rounds.
// This is more reliable than querying the parser's demo end state which can miss disconnected players.
func buildTeams(rounds []models.Round) models.Teams {
	type playerRecord struct {
		steamID uint64
		name    string
		team    string
	}

	seen := make(map[uint64]*playerRecord)

	for i := range rounds {
		for _, snap := range rounds[i].Snapshots {
			for _, ps := range snap.Players {
				if ps.SteamID == 0 || ps.Team == "" {
					continue
				}
				seen[ps.SteamID] = &playerRecord{
					steamID: ps.SteamID,
					name:    ps.Name,
					team:    ps.Team,
				}
			}
		}
	}

	var teams models.Teams
	for _, rec := range seen {
		info := models.PlayerInfo{
			SteamID: rec.steamID,
			Name:    rec.name,
		}
		switch rec.team {
		case "ct":
			teams.CT.Players = append(teams.CT.Players, info)
		case "t":
			teams.T.Players = append(teams.T.Players, info)
		}
	}

	return teams
}

func teamToString(team common.Team) string {
	switch team {
	case common.TeamCounterTerrorists:
		return "ct"
	case common.TeamTerrorists:
		return "t"
	default:
		return ""
	}
}

func roundEndReasonToString(reason events.RoundEndReason) string {
	switch reason {
	case events.RoundEndReasonTerroristsWin, events.RoundEndReasonCTWin,
		events.RoundEndReasonTerroristsStopped, events.RoundEndReasonCTStoppedEscape:
		return "elimination"
	case events.RoundEndReasonBombDefused:
		return "bomb_defused"
	case events.RoundEndReasonTargetBombed:
		return "bomb_exploded"
	case events.RoundEndReasonTargetSaved:
		return "time"
	default:
		return "other"
	}
}
