import type { Round } from "../types/match";

interface RoundSelectorProps {
  rounds: Round[];
  currentIndex: number;
  onSelect: (index: number) => void;
}

export default function RoundSelector({ rounds, currentIndex, onSelect }: RoundSelectorProps) {
  return (
    <div className="flex items-center justify-center gap-0.5 overflow-x-auto py-1.5">
      {rounds.map((round, i) => {
        const isActive = i === currentIndex;
        const base = "shrink-0 cursor-pointer rounded border-none px-2 py-1 text-[13px] font-medium tabular-nums transition-colors";

        if (isActive) {
          const activeBg = round.winner === "ct" ? "bg-ct" : "bg-t";
          return (
            <button
              key={round.number}
              className={`${base} ${activeBg} font-bold text-bg`}
              onClick={() => onSelect(i)}
            >
              {String(round.number).padStart(2, "0")}
            </button>
          );
        }

        return (
          <button
            key={round.number}
            className={`${base} bg-transparent text-text-muted hover:bg-surface-hover hover:text-text-primary`}
            onClick={() => onSelect(i)}
          >
            {String(round.number).padStart(2, "0")}
          </button>
        );
      })}
    </div>
  );
}