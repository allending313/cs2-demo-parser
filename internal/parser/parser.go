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

const snapshotsPerSecond = 5

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

	p.RegisterEventHandler(func(e events.RoundStart) {
		collector.onRoundStart(p)
	})

	p.RegisterEventHandler(func(e events.RoundEnd) {
		collector.onRoundEnd(e)
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

	// Flush any in-progress round (in case demo ends mid-round)
	collector.flush()

	match.TickRate = p.TickRate()
	match.Duration = p.CurrentTime().Seconds()
	match.Teams = buildTeams(p)

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

	bombState   string
	bombCarrier uint64
}

func newRoundCollector(match *models.Match) *roundCollector {
	return &roundCollector{match: match}
}

func (c *roundCollector) onRoundStart(p demoinfocs.Parser) {
	// Flush previous round if it wasn't ended cleanly
	c.flush()

	gs := p.GameState()
	c.current = &models.Round{
		Number: gs.TotalRoundsPlayed() + 1,
	}
	c.snapshots = nil
	c.kills = nil
	c.roundStartTick = gs.IngameTick()
	c.lastSnapshotTick = 0

	tickRate := p.TickRate()
	if tickRate > 0 {
		c.sampleInterval = int(math.Round(tickRate / float64(snapshotsPerSecond)))
	}
	if c.sampleInterval < 1 {
		c.sampleInterval = 13 // fallback for 64-tick (approx 5 snapshots/sec)
	}

	// Reset bomb state for new round
	c.bombState = ""
	c.bombCarrier = 0
}

func (c *roundCollector) onRoundEnd(e events.RoundEnd) {
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

	c.current.Snapshots = c.snapshots
	c.current.Kills = c.kills
	c.match.Rounds = append(c.match.Rounds, *c.current)
	c.current = nil
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
			// Normalize to 0-255 range: max flash is ~5 seconds
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

// flush saves any in-progress round (e.g. when the demo ends mid-round).
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

func buildTeams(p demoinfocs.Parser) models.Teams {
	gs := p.GameState()
	teams := models.Teams{}

	if ct := gs.TeamCounterTerrorists(); ct != nil {
		teams.CT.Name = ct.ClanName()
		for _, player := range ct.Members() {
			if player == nil {
				continue
			}
			teams.CT.Players = append(teams.CT.Players, models.PlayerInfo{
				SteamID: player.SteamID64,
				Name:    player.Name,
			})
		}
	}

	if t := gs.TeamTerrorists(); t != nil {
		teams.T.Name = t.ClanName()
		for _, player := range t.Members() {
			if player == nil {
				continue
			}
			teams.T.Players = append(teams.T.Players, models.PlayerInfo{
				SteamID: player.SteamID64,
				Name:    player.Name,
			})
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
