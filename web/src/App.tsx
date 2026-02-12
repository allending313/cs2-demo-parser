import { useState, useEffect } from "react";
import type { MatchData } from "./types/match";
import MatchViewer from "./components/MatchViewer";
import "./index.css";

const MATCH_JSON_URL = "/match.json";

function radarUrl(mapName: string): string {
  return `/${mapName}.png`;
}

export default function App() {
  const [match, setMatch] = useState<MatchData | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetch(MATCH_JSON_URL)
      .then((res) => {
        if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
        return res.json();
      })
      .then(setMatch)
      .catch((err) => setError(err.message));
  }, []);

  if (error) {
    return (
      <div className="p-10 font-mono text-text-primary">
        <h2 className="text-lg font-bold">Failed to load match data</h2>
      </div>
    );
  }

  if (!match) {
    return <div className="p-10 text-text-muted">Loading match data...</div>;
  }

  return <MatchViewer match={match} radarImageUrl={radarUrl(match.map)} />;
}