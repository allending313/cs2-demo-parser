import { useState, useEffect } from "react";
import type { MatchData } from "../types/match";
import MatchViewer from "./MatchViewer";

function radarUrl(mapName: string): string {
  return `/api/maps/${mapName}/radar.png`;
}

type State =
  | { step: "loading" }
  | { step: "ready"; match: MatchData }
  | { step: "error"; message: string };

export default function MatchPage({ params }: { params: { id: string } }) {
  const [state, setState] = useState<State>({ step: "loading" });

  useEffect(() => {
    let cancelled = false;

    async function fetchMatch() {
      try {
        const res = await fetch(`/api/match/${params.id}`);
        if (!res.ok) {
          const body = await res.json().catch(() => null);
          if (!cancelled)
            setState({
              step: "error",
              message: body?.error ?? `Failed to load match (${res.status})`,
            });
          return;
        }
        const match: MatchData = await res.json();
        if (!cancelled) setState({ step: "ready", match });
      } catch {
        if (!cancelled)
          setState({ step: "error", message: "Failed to connect to server" });
      }
    }

    fetchMatch();
    return () => {
      cancelled = true;
    };
  }, [params.id]);

  if (state.step === "loading") {
    return (
      <div className="flex h-screen items-center justify-center bg-bg text-text-primary">
        <p className="text-lg">Loading match...</p>
      </div>
    );
  }

  if (state.step === "error") {
    return (
      <div className="flex h-screen items-center justify-center bg-bg text-text-primary">
        <div className="text-center">
          <p className="text-lg text-t">{state.message}</p>
          <a href="/" className="mt-4 inline-block text-sm text-text-muted hover:text-text-primary">
            Upload a demo
          </a>
        </div>
      </div>
    );
  }

  return (
    <MatchViewer
      match={state.match}
      radarImageUrl={radarUrl(state.match.map)}
    />
  );
}
