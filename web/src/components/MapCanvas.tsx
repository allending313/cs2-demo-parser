import { useRef, useState, useEffect, useCallback } from "react";
import type { MapConfig, PlayerState } from "../types/match";
import { worldToCanvas } from "../utils/coordinates";
import { getCSSColor } from "../utils/style";
import { drawPlayerDot, drawPlayerName, drawViewCone, drawX } from "../utils/draw";

interface MapCanvasProps {
  mapConfig: MapConfig;
  mapImageUrl: string;
  players: PlayerState[];
  width: number;
  height: number;
}

export default function MapCanvas({
  mapConfig,
  mapImageUrl,
  players,
  width,
  height,
}: MapCanvasProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const imageRef = useRef<HTMLImageElement | null>(null);
  const [imageLoaded, setImageLoaded] = useState(false);

  useEffect(() => {
    setImageLoaded(false);
    const img = new Image();
    img.src = mapImageUrl;
    img.onload = () => {
      imageRef.current = img;
      setImageLoaded(true);
    };
  }, [mapImageUrl]);

  const draw = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    ctx.clearRect(0, 0, width, height);

    if (imageLoaded && imageRef.current) {
      ctx.drawImage(imageRef.current, 0, 0, width, height);
    } else {
      ctx.fillStyle = "#1a1a2e";
      ctx.fillRect(0, 0, width, height);
    }

    const ctColor = getCSSColor("--color-ct");
    const tColor = getCSSColor("--color-t");
    const ctDim = getCSSColor("--color-ct-dim");
    const tDim = getCSSColor("--color-t-dim");

    for (const player of players) {
      const { x, y } = worldToCanvas(player.x, player.y, mapConfig, width, height);
      const color = player.team === "ct" ? ctColor : tColor;
      const deadColor = player.team === "ct" ? ctDim : tDim;

      if (!player.isAlive) {
        drawX(ctx, x, y, deadColor);
        continue;
      }

      drawViewCone(ctx, x, y, player.yaw, color);
      drawPlayerDot(ctx, x, y, color);
      drawPlayerName(ctx, x, y, player.name, color);
    }
  }, [players, mapConfig, mapImageUrl, imageLoaded, width, height]);

  useEffect(() => {
    const id = requestAnimationFrame(draw);
    return () => cancelAnimationFrame(id);
  }, [draw]);

  return (
    <canvas
      ref={canvasRef}
      width={width}
      height={height}
      className="block h-full w-full"
    />
  );
}
