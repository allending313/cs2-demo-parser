package parser

import (
	"errors"
	"fmt"
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

	p.RegisterEventHandler(func(e events.GrenadeProjectileThrow) {
		collector.onGrenadeThrow(e, p)
	})

	p.RegisterEventHandler(func(e events.GrenadeProjectileDestroy) {
		collector.onGrenadeDestroy(e, p)
	})

	p.RegisterEventHandler(func(e events.SmokeStart) {
		collector.onSmokeStart(e)
	})

	p.RegisterEventHandler(func(e events.SmokeExpired) {
		collector.onSmokeExpired(e, p)
	})

	p.RegisterEventHandler(func(e events.InfernoStart) {
		collector.onInfernoStart(e, p)
	})

	p.RegisterEventHandler(func(e events.InfernoExpired) {
		collector.onInfernoExpired(e, p)
	})

	p.RegisterEventHandler(func(e events.DecoyStart) {
		collector.onDecoyStart(e, p)
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

// buildTeams constructs team rosters from snapshot data across all rounds.
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
