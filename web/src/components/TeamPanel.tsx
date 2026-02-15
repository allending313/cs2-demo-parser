import type { Team, TeamInfo, PlayerState } from "../types/match";

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

          return (
            <div
              key={state.steamId}
              className={`flex items-center justify-between rounded px-2 py-1.5 text-[13px] transition-opacity ${rowBg} ${dead ? "opacity-40" : ""}`}
            >
              <span className="truncate font-medium text-text-primary">{state.name}</span>
              <span className="ml-2 shrink-0 text-xs font-semibold tabular-nums text-text-muted">
                {state.isAlive ? `${state.hp} HP` : "DEAD"}
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
}