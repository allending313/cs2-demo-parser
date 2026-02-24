import { useMemo, useEffect, useCallback } from "react";
import type { MatchData } from "../types/match";
import { usePlayback } from "../hooks/usePlayback";
import { formatTime } from "../utils/format";
import { getActiveGrenades } from "../utils/grenade";
import MapCanvas from "./MapCanvas";
import TeamPanel from "./TeamPanel";
import RoundSelector from "./RoundSelector";
import PlaybackControls from "./PlaybackControls";

interface MatchViewerProps {
  match: MatchData;
  radarImageUrl: string;
}

const MAP_SIZE = 700;

export default function MatchViewer({ match, radarImageUrl }: MatchViewerProps) {
  const [playback, controls] = usePlayback(match.rounds);
  const round = match.rounds[playback.roundIndex];

  const ctPlayers = useMemo(
    () => playback.players.filter((p) => p.team === "ct"),
    [playback.players]
  );
  const tPlayers = useMemo(
    () => playback.players.filter((p) => p.team === "t"),
    [playback.players]
  );

  // Resolve which org name belongs on which side right now.
  // match.teams is keyed by end-of-match sides, but players swap at halftime.
  const { ctTeam, tTeam } = useMemo(() => {
    const staticCtIds = new Set(match.teams.ct.players.map((p) => p.steamId));
    const swapped = ctPlayers.length > 0 && !staticCtIds.has(ctPlayers[0]!.steamId);
    return swapped
      ? { ctTeam: match.teams.t, tTeam: match.teams.ct }
      : { ctTeam: match.teams.ct, tTeam: match.teams.t };
  }, [ctPlayers, match.teams]);

  const activeGrenades = useMemo(
    () => getActiveGrenades(round?.grenades ?? [], playback.currentTime),
    [round?.grenades, playback.currentTime]
  );

  // Show the score at the start of this round, not the end
  const ctScore = round ? round.endCTScore - (round.winner === "ct" ? 1 : 0) : 0;
  const tScore = round ? round.endTScore - (round.winner === "t" ? 1 : 0) : 0;

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;

      switch (e.key) {
        case " ":
          e.preventDefault();
          controls.togglePlay();
          break;
        case "ArrowLeft":
          e.preventDefault();
          controls.seek(playback.currentTime - 5);
          break;
        case "ArrowRight":
          e.preventDefault();
          controls.seek(playback.currentTime + 5);
          break;
        case "ArrowUp":
          e.preventDefault();
          controls.setRound(playback.roundIndex - 1);
          break;
        case "ArrowDown":
          e.preventDefault();
          controls.setRound(playback.roundIndex + 1);
          break;
        case ".":
          controls.cycleSpeed();
          break;
      }
    },
    [controls, playback.currentTime, playback.roundIndex]
  );

  useEffect(() => {
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [handleKeyDown]);

  return (
    <div className="flex h-screen flex-col bg-bg text-text-primary">
      {/* Timer */}
      <div className="mx-auto mt-3 w-20 rounded-md border border-border bg-surface py-3 text-center text-base font-semibold tabular-nums">
        {formatTime(playback.currentTime)}
      </div>

      {/* Main area */}
      <div className="flex min-h-0 flex-1 items-center justify-center gap-4 px-4 py-3">
        <TeamPanel side="t" team={tTeam} score={tScore} players={tPlayers} />

        <div
          className="shrink-0 overflow-hidden rounded shadow-lg"
          style={{ width: MAP_SIZE, height: MAP_SIZE }}
        >
          <MapCanvas
            mapConfig={match.mapConfig}
            mapImageUrl={radarImageUrl}
            players={playback.players}
            grenades={activeGrenades}
            width={MAP_SIZE}
            height={MAP_SIZE}
          />
        </div>

        <TeamPanel side="ct" team={ctTeam} score={ctScore} players={ctPlayers} />
      </div>

      {/* Bottom bar */}
      <div className="bg-surface">
        <RoundSelector
          rounds={match.rounds}
          currentIndex={playback.roundIndex}
          onSelect={controls.setRound}
        />
        <PlaybackControls
          isPlaying={playback.isPlaying}
          speed={playback.speed}
          currentTime={playback.currentTime}
          duration={playback.roundDuration}
          onTogglePlay={controls.togglePlay}
          onSetSpeed={controls.setSpeed}
          onSeek={controls.seek}
        />
      </div>
    </div>
  );
}