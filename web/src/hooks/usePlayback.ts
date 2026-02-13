import { useState, useRef, useCallback, useEffect } from "react";
import type { Round, Snapshot, PlayerState } from "../types/match";
import { interpolateSnapshot } from "../utils/interpolation";

export interface PlaybackState {
  isPlaying: boolean;
  speed: number;
  roundIndex: number;
  currentTime: number;
  roundDuration: number;
  players: PlayerState[];
  currentSnapshot: Snapshot | null;
}

export interface PlaybackControls {
  play: () => void;
  pause: () => void;
  togglePlay: () => void;
  setSpeed: (speed: number) => void;
  setRound: (index: number) => void;
  seek: (time: number) => void;
}

const SPEEDS = [0.5, 1, 2, 4];

export function usePlayback(rounds: Round[]): [PlaybackState, PlaybackControls] {
  const [isPlaying, setIsPlaying] = useState(true);
  const [speed, setSpeedState] = useState(1);
  const [roundIndex, setRoundIndex] = useState(0);
  const [currentTime, setCurrentTime] = useState(0);

  const rafRef = useRef<number>(0);
  const lastFrameRef = useRef<number>(0);

  const round = rounds[roundIndex] as Round | undefined;
  const snapshots = round?.snapshots ?? [];
  const roundDuration =
    snapshots.length > 0 ? snapshots[snapshots.length - 1]!.timeInRound : 0;

  const { players, currentSnapshot } = getInterpolatedFrame(snapshots, currentTime);

  const tick = useCallback(() => {
    const now = performance.now();
    const dt = (now - lastFrameRef.current) / 1000;
    lastFrameRef.current = now;

    setCurrentTime((prev) => {
      const next = prev + dt * speed;
      if (next >= roundDuration) {
        setIsPlaying(false);
        return roundDuration;
      }
      return next;
    });

    rafRef.current = requestAnimationFrame(tick);
  }, [speed, roundDuration]);

  useEffect(() => {
    if (isPlaying) {
      lastFrameRef.current = performance.now();
      rafRef.current = requestAnimationFrame(tick);
    }
    return () => cancelAnimationFrame(rafRef.current);
  }, [isPlaying, tick]);

  const play = useCallback(() => {
    if (currentTime >= roundDuration) setCurrentTime(0);
    setIsPlaying(true);
  }, [currentTime, roundDuration]);

  const pause = useCallback(() => setIsPlaying(false), []);

  const togglePlay = useCallback(() => {
    if (isPlaying) pause();
    else play();
  }, [isPlaying, play, pause]);

  const setSpeed = useCallback((s: number) => {
    if (SPEEDS.includes(s)) setSpeedState(s);
  }, []);

  const setRound = useCallback(
    (index: number) => {
      if (index < 0 || index >= rounds.length) return;
      setRoundIndex(index);
      setCurrentTime(0);
      setIsPlaying(true);
    },
    [rounds.length]
  );

  const seek = useCallback(
    (time: number) => setCurrentTime(Math.max(0, Math.min(time, roundDuration))),
    [roundDuration]
  );

  return [
    { isPlaying, speed, roundIndex, currentTime, roundDuration, players, currentSnapshot },
    { play, pause, togglePlay, setSpeed, setRound, seek },
  ];
}

function getInterpolatedFrame(
  snapshots: Snapshot[],
  time: number
): { players: PlayerState[]; currentSnapshot: Snapshot | null } {
  if (snapshots.length === 0) return { players: [], currentSnapshot: null };
  if (snapshots.length === 1)
    return { players: snapshots[0]!.players, currentSnapshot: snapshots[0]! };

  if (time <= snapshots[0]!.timeInRound)
    return { players: snapshots[0]!.players, currentSnapshot: snapshots[0]! };

  const last = snapshots[snapshots.length - 1]!;
  if (time >= last.timeInRound) return { players: last.players, currentSnapshot: last };

  let lo = 0;
  let hi = snapshots.length - 1;
  while (lo < hi - 1) {
    const mid = (lo + hi) >> 1;
    if (snapshots[mid]!.timeInRound <= time) lo = mid;
    else hi = mid;
  }

  const prev = snapshots[lo]!;
  const next = snapshots[hi]!;
  const segmentDuration = next.timeInRound - prev.timeInRound;
  const t = segmentDuration > 0 ? (time - prev.timeInRound) / segmentDuration : 0;

  return {
    players: interpolateSnapshot(prev, next, t),
    currentSnapshot: t < 0.5 ? prev : next,
  };
}