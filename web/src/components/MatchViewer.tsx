import { useMemo } from "react";
import type { MatchData } from "../types/match";
import { usePlayback } from "../hooks/usePlayback";
import { formatTime } from "../utils/format";
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

  // Show the score at the start of this round, not the end
  const ctScore = round ? round.endCTScore - (round.winner === "ct" ? 1 : 0) : 0;
  const tScore = round ? round.endTScore - (round.winner === "t" ? 1 : 0) : 0;

  return (
    <div className="flex h-screen flex-col bg-bg text-text-primary">
      {/* Timer */}
      <div className="mx-auto mt-3 w-20 rounded-md border border-border bg-surface py-3 text-center text-base font-semibold tabular-nums">
        {formatTime(playback.currentTime)}
      </div>

      {/* Main area */}
      <div className="flex min-h-0 flex-1 items-center justify-center gap-4 px-4 py-3">
        <TeamPanel side="t" team={match.teams.t} score={tScore} players={tPlayers} />

        <div
          className="shrink-0 overflow-hidden rounded shadow-lg"
          style={{ width: MAP_SIZE, height: MAP_SIZE }}
        >
          <MapCanvas
            mapConfig={match.mapConfig}
            mapImageUrl={radarImageUrl}
            players={playback.players}
            width={MAP_SIZE}
            height={MAP_SIZE}
          />
        </div>

        <TeamPanel side="ct" team={match.teams.ct} score={ctScore} players={ctPlayers} />
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