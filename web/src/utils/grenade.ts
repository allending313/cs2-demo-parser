import type { GrenadeEvent, GrenadeType, TrajectoryPoint } from "../types/match";
import { lerp } from "./interpolation";

export interface ActiveGrenade {
  type: GrenadeType;
  x: number;
  y: number;
  state: "inflight" | "effect";
  trail: TrajectoryPoint[];
}

const EFFECT_TYPES: Set<GrenadeType> = new Set(["smoke", "molotov", "incendiary", "decoy"]);

export function getActiveGrenades(
  grenades: GrenadeEvent[],
  currentTime: number,
): ActiveGrenade[] {
  const active: ActiveGrenade[] = [];

  for (const g of grenades) {
    if (currentTime < g.throwTime) continue;

    const hasDetonation = g.detonateTime > 0;

    if (!hasDetonation || currentTime < g.detonateTime) {
      const pos = interpolateTrajectory(g.trajectory, currentTime);
      const trail = getTrailPoints(g.trajectory, currentTime);
      active.push({ type: g.type, x: pos.x, y: pos.y, state: "inflight", trail });
      continue;
    }

    if (EFFECT_TYPES.has(g.type) && g.effectDuration > 0) {
      if (currentTime < g.detonateTime + g.effectDuration) {
        active.push({ type: g.type, x: g.detonateX, y: g.detonateY, state: "effect", trail: [] });
      }
    }
  }

  return active;
}

function getTrailPoints(
  pts: TrajectoryPoint[],
  time: number,
): TrajectoryPoint[] {
  if (pts.length === 0) return [];

  const trail: TrajectoryPoint[] = [];
  for (const p of pts) {
    if (p.t > time) break;
    trail.push(p);
  }

  // Add interpolated current position as final point
  const pos = interpolateTrajectory(pts, time);
  trail.push({ t: time, x: pos.x, y: pos.y });

  return trail;
}

function interpolateTrajectory(
  pts: TrajectoryPoint[],
  time: number,
): { x: number; y: number } {
  if (pts.length === 0) return { x: 0, y: 0 };
  if (pts.length === 1 || time <= pts[0]!.t) return { x: pts[0]!.x, y: pts[0]!.y };

  const last = pts[pts.length - 1]!;
  if (time >= last.t) return { x: last.x, y: last.y };

  for (let i = 0; i < pts.length - 1; i++) {
    const a = pts[i]!;
    const b = pts[i + 1]!;
    if (time >= a.t && time < b.t) {
      const t = (time - a.t) / (b.t - a.t);
      return { x: lerp(a.x, b.x, t), y: lerp(a.y, b.y, t) };
    }
  }

  return { x: last.x, y: last.y };
}
