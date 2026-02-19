package models

type Match struct {
	ID        string     `json:"id"`
	Map       string     `json:"map"`
	TickRate  float64    `json:"tickRate"`
	Duration  float64    `json:"duration"`
	Teams     Teams      `json:"teams"`
	Rounds    []Round    `json:"rounds"`
	MapConfig *MapConfig `json:"mapConfig,omitempty"`
}

type Teams struct {
	CT TeamInfo `json:"ct"`
	T  TeamInfo `json:"t"`
}

type TeamInfo struct {
	Name    string       `json:"name"`
	Players []PlayerInfo `json:"players"`
}

type PlayerInfo struct {
	SteamID uint64 `json:"steamId"`
	Name    string `json:"name"`
}

type Round struct {
	Number     int            `json:"number"`
	Winner     string         `json:"winner"`
	WinReason  string         `json:"winReason"`
	EndTScore  int            `json:"endTScore"`
	EndCTScore int            `json:"endCTScore"`
	Snapshots  []Snapshot     `json:"snapshots"`
	Kills      []KillEvent    `json:"kills"`
	Grenades   []GrenadeEvent `json:"grenades"`
}

type Snapshot struct {
	Tick        int           `json:"tick"`
	TimeInRound float64       `json:"timeInRound"`
	Bomb        *BombState    `json:"bomb,omitempty"`
	Players     []PlayerState `json:"players"`
}

type BombState struct {
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	State   string  `json:"state"`
	Carrier uint64  `json:"carrier,omitempty"`
}

type PlayerState struct {
	SteamID    uint64  `json:"steamId"`
	Name       string  `json:"name"`
	Team       string  `json:"team"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Z          float64 `json:"z"`
	Yaw        float64 `json:"yaw"`
	HP         int     `json:"hp"`
	Armor      int     `json:"armor"`
	IsAlive    bool    `json:"isAlive"`
	Weapon     string  `json:"weapon"`
	HasDefuser bool    `json:"hasDefuser"`
	FlashAlpha float64 `json:"flashAlpha"`
}

type KillEvent struct {
	Tick        int     `json:"tick"`
	TimeInRound float64 `json:"timeInRound"`
	Attacker    uint64  `json:"attacker"`
	Victim      uint64  `json:"victim"`
	Weapon      string  `json:"weapon"`
	Headshot    bool    `json:"headshot"`
	Wallbang    bool    `json:"wallbang"`
	AttackerX   float64 `json:"attackerX"`
	AttackerY   float64 `json:"attackerY"`
	VictimX     float64 `json:"victimX"`
	VictimY     float64 `json:"victimY"`
}

type GrenadeEvent struct {
	Type    string `json:"type"`
	Thrower uint64 `json:"thrower"`

	ThrowTick int     `json:"throwTick"`
	ThrowTime float64 `json:"throwTime"`
	ThrowX    float64 `json:"throwX"`
	ThrowY    float64 `json:"throwY"`

	DetonateTick int     `json:"detonateTick"`
	DetonateTime float64 `json:"detonateTime"`
	DetonateX    float64 `json:"detonateX"`
	DetonateY    float64 `json:"detonateY"`

	EffectDuration float64 `json:"effectDuration,omitempty"`

	Trajectory []TrajectoryPoint `json:"trajectory,omitempty"`
}

type TrajectoryPoint struct {
	TimeInRound float64 `json:"t"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
}

type MapConfig struct {
	Name           string  `json:"name"`
	DisplayName    string  `json:"displayName"`
	PosX           float64 `json:"posX"`
	PosY           float64 `json:"posY"`
	Scale          float64 `json:"scale"`
	RadarFile      string  `json:"radarFile"`
	LowerRadarFile *string `json:"lowerRadarFile"`
	RadarWidth     int     `json:"radarWidth"`
	RadarHeight    int     `json:"radarHeight"`
}
