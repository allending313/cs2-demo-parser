import { useMemo } from "react";
import type { KillEvent, PlayerState } from "../types/match";

interface KillFeedProps {
  kills: KillEvent[];
  currentTime: number;
  players: PlayerState[];
}

const DISPLAY_DURATION = 5;
const MAX_VISIBLE = 5;

// TODO: show assists on the kill feed
// will need to capture assists in the parser first and add that to the KillEvent field

export default function KillFeed({ kills, currentTime, players }: KillFeedProps) {
  const playerMap = useMemo(() => {
    const map = new Map<string, { name: string; team: string }>();
    for (const p of players) {
      map.set(p.steamId, { name: p.name, team: p.team });
    }
    return map;
  }, [players]);

  const visibleKills = useMemo(() => {
    return kills
      .filter(
        (k) =>
          k.timeInRound <= currentTime &&
          k.timeInRound > currentTime - DISPLAY_DURATION
      )
      .slice(-MAX_VISIBLE);
  }, [kills, currentTime]);

  if (visibleKills.length === 0) return null;

  return (
    <div className="absolute top-2 right-2 z-10 flex flex-col items-end gap-0.5">
      {visibleKills.map((kill, i) => {
        const attacker = playerMap.get(kill.attacker);
        const victim = playerMap.get(kill.victim);

        return (
          <div
            key={`${kill.tick}-${i}`}
            className="flex items-center gap-1.5 rounded bg-black/60 px-2 py-0.5 text-xs"
          >
            <span className={attacker?.team === "ct" ? "text-ct" : "text-t"}>
              {attacker?.name ?? "?"}
            </span>
            <span className="text-text-muted">{kill.weapon}</span>
            <span className={victim?.team === "ct" ? "text-ct" : "text-t"}>
              {victim?.name ?? "?"}
            </span>
          </div>
        );
      })}
    </div>
  );
}
