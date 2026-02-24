package parser

import (
	"math"

	demoinfocs "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	common "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
	events "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"

	models "github.com/allending313/cs2-demo-parser/internal/model"
)

const (
	// Max trajectory waypoints per grenade. Beyond this we downsample to
	// keep the JSON lean — the frontend lerps between points anyway.
	maxTrajectoryPoints = 10

	smokeDuration      = 18.0
	molotovMaxDuration = 7.0
	decoyDuration      = 15.0
)

// inflightGrenade tracks a grenade between throw and detonation so we can
// sample its trajectory without storing every tick.
type inflightGrenade struct {
	event      models.GrenadeEvent
	trajectory []models.TrajectoryPoint
	entityID   int
}

func (c *roundCollector) onGrenadeThrow(e events.GrenadeProjectileThrow, p demoinfocs.Parser) {
	if c.current == nil || e.Projectile == nil {
		return
	}

	gs := p.GameState()
	tick := gs.IngameTick()
	pos := e.Projectile.Position()

	ge := models.GrenadeEvent{
		Type:      grenadeTypeString(e.Projectile.WeaponInstance),
		ThrowTick: tick,
		ThrowTime: c.ticksToSeconds(tick, p),
		ThrowX:    pos.X,
		ThrowY:    pos.Y,
	}

	if e.Projectile.Thrower != nil {
		ge.Thrower = e.Projectile.Thrower.SteamID64
	}

	c.inflight[e.Projectile.Entity.ID()] = &inflightGrenade{
		event:    ge,
		entityID: e.Projectile.Entity.ID(),
		trajectory: []models.TrajectoryPoint{{
			TimeInRound: ge.ThrowTime,
			X:           pos.X,
			Y:           pos.Y,
		}},
	}
}

func (c *roundCollector) onGrenadeDestroy(e events.GrenadeProjectileDestroy, p demoinfocs.Parser) {
	if c.current == nil || e.Projectile == nil {
		return
	}

	id := e.Projectile.Entity.ID()
	ig, ok := c.inflight[id]
	if !ok {
		return
	}

	gs := p.GameState()
	tick := gs.IngameTick()
	pos := e.Projectile.Position()

	ig.event.DetonateTick = tick
	ig.event.DetonateTime = c.ticksToSeconds(tick, p)
	ig.event.DetonateX = pos.X
	ig.event.DetonateY = pos.Y

	// Append the detonation point so the trajectory covers the full flight path
	ig.trajectory = append(ig.trajectory, models.TrajectoryPoint{
		TimeInRound: ig.event.DetonateTime,
		X:           pos.X,
		Y:           pos.Y,
	})
	ig.event.Trajectory = downsampleTrajectory(ig.trajectory, maxTrajectoryPoints)

	c.grenades = append(c.grenades, ig.event)
	delete(c.inflight, id)
}

func (c *roundCollector) onHeExplode(e events.HeExplode, p demoinfocs.Parser) {
	if c.current == nil {
		return
	}

	id := e.GrenadeEntityID
	ig, ok := c.inflight[id]
	if !ok {
		return
	}

	gs := p.GameState()
	tick := gs.IngameTick()

	ig.event.DetonateTick = tick
	ig.event.DetonateTime = c.ticksToSeconds(tick, p)
	ig.event.DetonateX = e.Position.X
	ig.event.DetonateY = e.Position.Y

	ig.trajectory = append(ig.trajectory, models.TrajectoryPoint{
		TimeInRound: ig.event.DetonateTime,
		X:           e.Position.X,
		Y:           e.Position.Y,
	})
	ig.event.Trajectory = downsampleTrajectory(ig.trajectory, maxTrajectoryPoints)

	c.grenades = append(c.grenades, ig.event)
	delete(c.inflight, id)
}

func (c *roundCollector) onSmokeStart(e events.SmokeStart, p demoinfocs.Parser) {
	if c.current == nil {
		return
	}

	gs := p.GameState()
	tick := gs.IngameTick()
	detonateTime := c.ticksToSeconds(tick, p)

	// Find the inflight smoke grenade closest to this position and finalize it.
	// GrenadeProjectileDestroy fires too late for smokes (when the cloud clears),
	// so we use SmokeStart as the actual detonation event.
	bestID := -1
	bestDist := math.MaxFloat64
	for id, ig := range c.inflight {
		if ig.event.Type != "smoke" {
			continue
		}
		dx := ig.event.ThrowX - e.Position.X
		dy := ig.event.ThrowY - e.Position.Y
		dist := dx*dx + dy*dy
		if dist < bestDist {
			bestDist = dist
			bestID = id
		}
	}

	key := quantizePos(e.Position.X, e.Position.Y)

	if bestID >= 0 {
		ig := c.inflight[bestID]
		ig.event.DetonateTick = tick
		ig.event.DetonateTime = detonateTime
		ig.event.DetonateX = e.Position.X
		ig.event.DetonateY = e.Position.Y
		ig.event.EffectDuration = smokeDuration

		ig.trajectory = append(ig.trajectory, models.TrajectoryPoint{
			TimeInRound: detonateTime,
			X:           e.Position.X,
			Y:           e.Position.Y,
		})
		ig.event.Trajectory = downsampleTrajectory(ig.trajectory, maxTrajectoryPoints)

		idx := len(c.grenades)
		c.grenades = append(c.grenades, ig.event)
		delete(c.inflight, bestID)
		c.smokeByPos[key] = idx
	} else {
		idx := c.findGrenadeByTypeAndPos("smoke", e.Position.X, e.Position.Y)
		if idx >= 0 {
			c.grenades[idx].EffectDuration = smokeDuration
			c.smokeByPos[key] = idx
		}
	}
}

func (c *roundCollector) onSmokeExpired(e events.SmokeExpired, p demoinfocs.Parser) {
	if c.current == nil {
		return
	}

	key := quantizePos(e.Position.X, e.Position.Y)
	idx, ok := c.smokeByPos[key]
	if !ok || idx >= len(c.grenades) {
		return
	}

	// Compute actual duration from detonation to expiry
	expireTime := c.ticksToSeconds(p.GameState().IngameTick(), p)
	if det := c.grenades[idx].DetonateTime; det > 0 {
		c.grenades[idx].EffectDuration = expireTime - det
	}
	delete(c.smokeByPos, key)
}

func (c *roundCollector) onInfernoStart(e events.InfernoStart, p demoinfocs.Parser) {
	if c.current == nil {
		return
	}

	// Compute inferno center from its fires to match against the correct grenade
	var cx, cy float64
	fires := e.Inferno.Fires().Active().List()
	if len(fires) > 0 {
		for _, f := range fires {
			cx += f.Vector.X
			cy += f.Vector.Y
		}
		cx /= float64(len(fires))
		cy /= float64(len(fires))
	}

	idx := c.findGrenadeByTypeAndPos("molotov", cx, cy)
	if idx < 0 {
		idx = c.findGrenadeByTypeAndPos("incendiary", cx, cy)
	}
	if idx >= 0 {
		c.grenades[idx].EffectDuration = molotovMaxDuration
		c.infernoByUID[int64(e.Inferno.Entity.ID())] = idx
	}
}

func (c *roundCollector) onInfernoExpired(e events.InfernoExpired, p demoinfocs.Parser) {
	if c.current == nil {
		return
	}

	uid := e.Inferno.Entity.ID()
	idx, ok := c.infernoByUID[int64(uid)]
	if !ok || idx >= len(c.grenades) {
		return
	}

	expireTime := c.ticksToSeconds(p.GameState().IngameTick(), p)
	if det := c.grenades[idx].DetonateTime; det > 0 {
		duration := expireTime - det
		if duration > molotovMaxDuration {
			duration = molotovMaxDuration
		}
		c.grenades[idx].EffectDuration = duration
	}
	delete(c.infernoByUID, int64(uid))
}

func (c *roundCollector) onDecoyStart(e events.DecoyStart, p demoinfocs.Parser) {
	if c.current == nil {
		return
	}

	// Patch the closest unmatched decoy by position
	idx := c.findGrenadeByTypeAndPos("decoy", e.Position.X, e.Position.Y)
	if idx >= 0 {
		c.grenades[idx].EffectDuration = decoyDuration
	}
}

// findGrenadeByTypeAndPos returns the index of the unmatched grenade of the
// given type whose detonation position is closest to (x, y), or -1.
// This avoids mis-matching when multiple grenades of the same type detonate
// close together in time.
func (c *roundCollector) findGrenadeByTypeAndPos(typ string, x, y float64) int {
	bestIdx := -1
	bestDist := math.MaxFloat64
	for i := len(c.grenades) - 1; i >= 0; i-- {
		g := c.grenades[i]
		if g.Type != typ || g.EffectDuration != 0 {
			continue
		}
		dx := g.DetonateX - x
		dy := g.DetonateY - y
		dist := dx*dx + dy*dy
		if dist < bestDist {
			bestDist = dist
			bestIdx = i
		}
	}
	return bestIdx
}

// finalizeInflightGrenades commits any grenades still mid-air at round end.
func (c *roundCollector) finalizeInflightGrenades() {
	for id, ig := range c.inflight {
		ig.event.Trajectory = downsampleTrajectory(ig.trajectory, maxTrajectoryPoints)
		c.grenades = append(c.grenades, ig.event)
		delete(c.inflight, id)
	}
}

// sampleGrenadePositions records the current position of each in-flight
// grenade. Called at the snapshot sampling interval, not every tick.
func (c *roundCollector) sampleGrenadePositions(gs demoinfocs.GameState, p demoinfocs.Parser) {
	for _, proj := range gs.GrenadeProjectiles() {
		if proj == nil {
			continue
		}
		id := proj.Entity.ID()
		ig, ok := c.inflight[id]
		if !ok {
			continue
		}
		pos := proj.Position()
		ig.trajectory = append(ig.trajectory, models.TrajectoryPoint{
			TimeInRound: c.ticksToSeconds(gs.IngameTick(), p),
			X:           pos.X,
			Y:           pos.Y,
		})
	}
}

func grenadeTypeString(w *common.Equipment) string {
	if w == nil {
		return "unknown"
	}
	switch w.Type {
	case common.EqSmoke:
		return "smoke"
	case common.EqFlash:
		return "flash"
	case common.EqHE:
		return "he"
	case common.EqMolotov:
		return "molotov"
	case common.EqIncendiary:
		return "incendiary"
	case common.EqDecoy:
		return "decoy"
	default:
		return "unknown"
	}
}

// quantizePos buckets world coordinates to integers for use as map keys.
// Grenade positions across related events (e.g. SmokeStart/SmokeExpired) can
// differ by small floating-point amounts, so we round to the nearest unit.
func quantizePos(x, y float64) [2]int {
	return [2]int{int(math.Round(x)), int(math.Round(y))}
}

// downsampleTrajectory reduces a trajectory to at most maxPoints using
// largest-triangle-three-buckets, preserving the first and last points.
func downsampleTrajectory(pts []models.TrajectoryPoint, maxPoints int) []models.TrajectoryPoint {
	if len(pts) <= maxPoints {
		return pts
	}

	result := make([]models.TrajectoryPoint, 0, maxPoints)
	result = append(result, pts[0])

	bucketSize := float64(len(pts)-2) / float64(maxPoints-2)
	prevIdx := 0

	for i := 1; i < maxPoints-1; i++ {
		bucketStart := int(float64(i-1)*bucketSize) + 1
		bucketEnd := int(float64(i)*bucketSize) + 1
		if bucketEnd > len(pts)-1 {
			bucketEnd = len(pts) - 1
		}

		// Average of the next bucket (used as the target triangle vertex)
		nextStart := int(float64(i)*bucketSize) + 1
		nextEnd := int(float64(i+1)*bucketSize) + 1
		if nextEnd > len(pts) {
			nextEnd = len(pts)
		}
		var avgX, avgY float64
		for j := nextStart; j < nextEnd; j++ {
			avgX += pts[j].X
			avgY += pts[j].Y
		}
		count := float64(nextEnd - nextStart)
		if count > 0 {
			avgX /= count
			avgY /= count
		}

		// Pick the point in this bucket with the largest triangle area
		bestIdx := bucketStart
		bestArea := -1.0
		for j := bucketStart; j < bucketEnd; j++ {
			area := math.Abs(
				(pts[prevIdx].X-avgX)*(pts[j].Y-pts[prevIdx].Y) -
					(pts[prevIdx].X-pts[j].X)*(avgY-pts[prevIdx].Y),
			)
			if area > bestArea {
				bestArea = area
				bestIdx = j
			}
		}

		result = append(result, pts[bestIdx])
		prevIdx = bestIdx
	}

	result = append(result, pts[len(pts)-1])
	return result
}

// entityID adapts the various grenade event entity references into a
// consistent integer ID.
func entityID(e interface{ Entity() interface{ ID() int } }) int {
	return e.Entity().ID()
}
