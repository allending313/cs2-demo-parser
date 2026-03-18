import { useCallback, useMemo, useRef } from "react";
import { formatTime, ROUND_TIME } from "../utils/format";
import type { KillEvent, PlayerState } from "../types/match";

interface PlaybackControlsProps {
  isPlaying: boolean;
  speed: number;
  currentTime: number;
  duration: number;
  kills: KillEvent[];
  players: PlayerState[];
  onTogglePlay: () => void;
  onSetSpeed: (speed: number) => void;
  onSeek: (time: number) => void;
}

const SPEEDS = [0.5, 1, 2, 4];

export default function PlaybackControls({
  isPlaying,
  speed,
  currentTime,
  duration,
  kills,
  players,
  onTogglePlay,
  onSetSpeed,
  onSeek,
}: PlaybackControlsProps) {
  const trackRef = useRef<HTMLDivElement>(null);

  const killMarkers = useMemo(() => {
    const teamBySteamId = new Map<string, "ct" | "t">();
    for (const p of players) teamBySteamId.set(p.steamId, p.team);

    return kills.map((kill) => {
      const attackerTeam = teamBySteamId.get(kill.attacker);
      const victimTeam = teamBySteamId.get(kill.victim);
      const isTeamkillOrSuicide =
        !attackerTeam || kill.attacker === kill.victim || attackerTeam === victimTeam;
      const colorClass = isTeamkillOrSuicide
        ? "bg-text-muted"
        : attackerTeam === "ct"
          ? "bg-ct"
          : "bg-t";
      const pct = duration > 0 ? (kill.timeInRound / duration) * 100 : 0;

      return { pct, colorClass, timeInRound: kill.timeInRound };
    });
  }, [kills, players, duration]);

  const cycleSpeed = useCallback(() => {
    const idx = SPEEDS.indexOf(speed);
    const next = SPEEDS[(idx + 1) % SPEEDS.length]!;
    onSetSpeed(next);
  }, [speed, onSetSpeed]);

  const handleTrackClick = useCallback(
    (e: React.MouseEvent<HTMLDivElement>) => {
      const track = trackRef.current;
      if (!track || duration <= 0) return;
      const rect = track.getBoundingClientRect();
      const ratio = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width));
      onSeek(ratio * duration);
    },
    [duration, onSeek]
  );

  const progress = duration > 0 ? (currentTime / duration) * 100 : 0;

  return (
    <div className="flex items-center gap-3 px-4 py-2">
      <button
        className="min-w-9 cursor-pointer rounded border border-border bg-transparent px-2 py-0.5 text-xs font-semibold tabular-nums text-text-muted transition-colors hover:border-text-muted hover:text-text-primary"
        onClick={cycleSpeed}
        title="Playback speed"
      >
        x{speed}
      </button>

      <button
        className="flex cursor-pointer items-center border-none bg-transparent p-1 text-ct transition-opacity hover:opacity-80"
        onClick={onTogglePlay}
        title={isPlaying ? "Pause" : "Play"}
      >
        {isPlaying ? <PauseIcon /> : <PlayIcon />}
      </button>

      <div
        ref={trackRef}
        className="relative h-5 flex-1 cursor-pointer rounded-sm"
        onClick={handleTrackClick}
      >
        {/* Track background */}
        <div className="absolute top-1/2 h-1 w-full -translate-y-1/2 rounded-sm bg-surface-hover">
          <div
            className="pointer-events-none h-full rounded-sm bg-ct"
            style={{ width: `${progress}%` }}
          />
        </div>

        {/* Kill markers */}
        {killMarkers.map((marker, i) => (
          <button
            key={i}
            className={`absolute top-0 h-full w-1.5 -translate-x-1/2 rounded-sm ${marker.colorClass} transition-opacity hover:opacity-70`}
            style={{ left: `${marker.pct}%` }}
            onClick={(e) => {
              e.stopPropagation();
              onSeek(Math.max(0, marker.timeInRound - 1));
            }}
            title={`Kill at ${formatTime(ROUND_TIME - marker.timeInRound)}`}
          />
        ))}

        {/* Playhead */}
        <div
          className="pointer-events-none absolute top-1/2 h-3 w-3 -translate-y-1/2 rounded-full bg-ct shadow-md"
          style={{ left: `${progress}%`, transform: `translate(-50%, -50%)` }}
        />
      </div>

      <span className="min-w-10.5 text-right text-[13px] font-medium tabular-nums text-text-muted">
        {formatTime(ROUND_TIME - currentTime)}
      </span>
    </div>
  );
}

// temp icons for now
function PlayIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 14 14" fill="currentColor">
      <path d="M3 1.5v11l9-5.5z" />
    </svg>
  );
}

// temp icons for now
function PauseIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 14 14" fill="currentColor">
      <rect x="2" y="1" width="3.5" height="12" rx="0.5" />
      <rect x="8.5" y="1" width="3.5" height="12" rx="0.5" />
    </svg>
  );
}