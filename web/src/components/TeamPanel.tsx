import type { Team, TeamInfo, PlayerState } from "../types/match";
import { grenadeLabel, armorLabel } from "../utils/format";

interface TeamPanelProps {
  side: Team;
  team: TeamInfo;
  score: number;
  players: PlayerState[];
}

export default function TeamPanel({ side, team, score, players }: TeamPanelProps) {
  const nameColor = side === "ct" ? "text-ct" : "text-t";
  const rowBg = side === "ct" ? "bg-ct-dim" : "bg-t-dim";

  return (
    <div className="flex w-56 shrink-0 flex-col gap-0.5 p-3">
      <div className="mb-1 flex items-center justify-between border-b border-border pb-2.5">
        <span className={`text-sm font-bold uppercase tracking-wide ${nameColor}`}>
          {team.name}
        </span>
        <span className="text-xl font-bold tabular-nums text-text-primary">{score}</span>
      </div>

      <div className="flex flex-col gap-0.5">
        {[...players].sort((a, b) => (a.steamId < b.steamId ? -1 : a.steamId > b.steamId ? 1 : 0)).map((state) => {
          const dead = !state.isAlive;
          const grenadeLabels = (state.grenades ?? []).map(grenadeLabel);
          const armor = armorLabel(state.armor, state.hasHelmet);

          return (
            <div
              key={state.steamId}
              className={`flex flex-col rounded px-2 py-1.5 text-[13px] transition-opacity ${rowBg} ${dead ? "opacity-40" : ""}`}
            >
              <div className="flex items-center justify-between">
                <span className="truncate font-medium text-text-primary">{state.name}</span>
                <span className="ml-2 shrink-0 text-xs font-semibold tabular-nums text-text-muted">
                  {state.isAlive ? `${state.hp} HP` : "DEAD"}
                </span>
              </div>
              {state.isAlive && (
                <div className="mt-0.5 flex items-center gap-2 text-[11px] text-text-muted">
                  <span className="truncate">{state.weapon}</span>
                  {grenadeLabels.length > 0 && (
                    <span className="shrink-0">{grenadeLabels.join(" ")}</span>
                  )}
                  {armor && <span className="shrink-0">{armor}</span>}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}