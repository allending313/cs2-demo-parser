export type Team = "ct" | "t";
export type BombState = "carried" | "planted" | "dropped" | "defused" | "exploded";
export type WinReason = "elimination" | "bomb_defused" | "bomb_exploded" | "time";

export interface TeamInfo {
  name: string;
  players: TeamPlayer[];
}

export interface TeamPlayer {
  steamId: string;
  name: string;
}

export interface PlayerState {
  steamId: string;
  name: string;
  team: Team;
  x: number;
  y: number;
  z: number;
  yaw: number;
  hp: number;
  armor: number;
  isAlive: boolean;
  weapon: string;
  hasDefuser: boolean;
  flashAlpha: number;
}

export interface BombInfo {
  x: number;
  y: number;
  state: BombState;
  carrier: string | null;
}

export interface Snapshot {
  tick: number;
  timeInRound: number;
  bomb: BombInfo;
  players: PlayerState[];
}

export interface KillEvent {
  tick: number;
  timeInRound: number;
  attacker: string;
  victim: string;
  weapon: string;
  headshot: boolean;
  wallbang: boolean;
  attackerX: number;
  attackerY: number;
  victimX: number;
  victimY: number;
}

export interface Round {
  number: number;
  winner: Team;
  winReason: WinReason;
  endTScore: number;
  endCTScore: number;
  snapshots: Snapshot[];
  kills: KillEvent[];
}

export interface MapConfig {
  posX: number;
  posY: number;
  scale: number;
  radarWidth: number;
  radarHeight: number;
}

export interface MatchData {
  id: string;
  map: string;
  tickRate: number;
  duration: number;
  teams: {
    ct: TeamInfo;
    t: TeamInfo;
  };
  rounds: Round[];
  mapConfig: MapConfig;
}