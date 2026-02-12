import type { Snapshot, PlayerState } from "../types/match";

export function lerp(a: number, b: number, t: number): number {
  return a + (b - a) * clamp01(t);
}

export function clamp01(t: number): number {
  return Math.max(0, Math.min(1, t));
}

export function angleLerp(a: number, b: number, t: number): number {
  let delta = ((b - a + 540) % 360) - 180; // shortest signed angle difference, prevents spinning the long way around
  return a + delta * clamp01(t);
}

// Interpolate a player's pos/yaw between two snapshots
export function interpolatePlayer(
  prev: PlayerState,
  next: PlayerState,
  t: number
): PlayerState {
  if (!next.isAlive) return next;

  return {
    ...next,
    x: lerp(prev.x, next.x, t),
    y: lerp(prev.y, next.y, t),
    z: lerp(prev.z, next.z, t),
    yaw: angleLerp(prev.yaw, next.yaw, t),
  };
}

function findPlayer(snapshot: Snapshot, steamId: string): PlayerState | undefined {
  return snapshot.players.find((p) => p.steamId === steamId);
}

// Interpolate all player states between two snapshots.
// next snapshot priority
export function interpolateSnapshot(
  prev: Snapshot,
  next: Snapshot,
  t: number
): PlayerState[] {
  return next.players.map((nextPlayer) => {
    const prevPlayer = findPlayer(prev, nextPlayer.steamId);
    if (!prevPlayer) return nextPlayer;
    return interpolatePlayer(prevPlayer, nextPlayer, t);
  });
}