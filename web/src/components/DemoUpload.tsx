import { useState, useCallback, useRef } from "react";
import { useLocation } from "wouter";

type UploadState =
  | { step: "idle" }
  | { step: "parsing"; id: string; progress: number }
  | { step: "error"; message: string };

export default function DemoUpload() {
  const [state, setState] = useState<UploadState>({ step: "idle" });
  const [dragging, setDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [, navigate] = useLocation();

  const uploadFile = useCallback(
    async (file: File) => {
      if (!file.name.endsWith(".dem")) {
        setState({ step: "error", message: "File must be a .dem file" });
        return;
      }

      const form = new FormData();
      form.append("demo", file);

      setState({ step: "parsing", id: "", progress: 0 });

      let res: Response;
      try {
        res = await fetch("/api/parse", { method: "POST", body: form });
      } catch {
        setState({ step: "error", message: "Failed to connect to server" });
        return;
      }

      if (!res.ok) {
        const body = await res.json().catch(() => null);
        setState({
          step: "error",
          message: body?.error ?? `Upload failed (${res.status})`,
        });
        return;
      }

      const job: { id: string } = await res.json();
      setState({ step: "parsing", id: job.id, progress: 0 });

      const pollInterval = setInterval(async () => {
        try {
          const statusRes = await fetch(`/api/match/${job.id}/status`);
          const status = await statusRes.json();

          if (status.status === "ready") {
            clearInterval(pollInterval);
            navigate(`/match/${job.id}`);
          } else if (status.status === "error") {
            clearInterval(pollInterval);
            setState({
              step: "error",
              message: status.error || "Parse failed",
            });
          } else {
            setState({
              step: "parsing",
              id: job.id,
              progress: status.progress ?? 0,
            });
          }
        } catch {
          clearInterval(pollInterval);
          setState({ step: "error", message: "Lost connection to server" });
        }
      }, 500);
    },
    [navigate]
  );

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setDragging(false);
      const file = e.dataTransfer.files[0];
      if (file) uploadFile(file);
    },
    [uploadFile]
  );

  const handleFileChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (file) uploadFile(file);
    },
    [uploadFile]
  );

  return (
    <div className="flex h-screen items-center justify-center bg-bg text-text-primary">
      <div className="w-full max-w-md px-4">
        {state.step === "parsing" ? (
          <div className="text-center">
            <h2 className="mb-4 text-lg font-semibold">Parsing demo...</h2>
            <div className="mx-auto h-2 w-64 overflow-hidden rounded-full bg-surface">
              <div
                className="h-full rounded-full bg-ct transition-all duration-300"
                style={{ width: `${Math.round(state.progress * 100)}%` }}
              />
            </div>
            <p className="mt-2 text-sm text-text-muted">
              {Math.round(state.progress * 100)}%
            </p>
          </div>
        ) : (
          <>
            {state.step === "error" && (
              <div className="mb-4 rounded-md border border-t-dim bg-surface px-4 py-3 text-sm text-t">
                {state.message}
              </div>
            )}
            <div
              className={`flex cursor-pointer flex-col items-center rounded-lg border-2 border-dashed px-6 py-12 transition-colors ${
                dragging
                  ? "border-ct bg-surface"
                  : "border-border hover:border-text-muted"
              }`}
              onDragOver={(e) => {
                e.preventDefault();
                setDragging(true);
              }}
              onDragLeave={() => setDragging(false)}
              onDrop={handleDrop}
              onClick={() => fileInputRef.current?.click()}
            >
              <p className="text-lg font-semibold">Drop a .dem file here</p>
              <p className="mt-1 text-sm text-text-muted">or click to browse</p>
              <input
                ref={fileInputRef}
                type="file"
                accept=".dem"
                className="hidden"
                onChange={handleFileChange}
              />
            </div>
          </>
        )}
      </div>
    </div>
  );
}
