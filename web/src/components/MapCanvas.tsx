import { useRef, useState, useEffect, useCallback } from "react";
import type { MapConfig, PlayerState } from "../types/match";
import { worldToCanvas } from "../utils/coordinates";
import { getCSSColor } from "../utils/style";

interface MapCanvasProps {
  mapConfig: MapConfig;
  mapImageUrl: string;
  players: PlayerState[];
  width: number;
  height: number;
}

const PLAYER_RADIUS = 6;
const FONT = "600 11px -apple-system, BlinkMacSystemFont, sans-serif";
const VIEW_CONE_LENGTH = 10;
const VIEW_CONE_ANGLE = Math.PI / 6;

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
      drawPlayerName(ctx, x, y, player.name);
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

function drawPlayerDot(ctx: CanvasRenderingContext2D, x: number, y: number, color: string) {
  ctx.beginPath();
  ctx.arc(x, y, PLAYER_RADIUS, 0, Math.PI * 2);
  ctx.fillStyle = color;
  ctx.fill();
  ctx.lineWidth = 1.5;
  ctx.strokeStyle = "rgba(0,0,0,0.5)";
  ctx.stroke();
}

function drawPlayerName(ctx: CanvasRenderingContext2D, x: number, y: number, name: string) {
  ctx.font = FONT;
  ctx.textAlign = "center";
  ctx.textBaseline = "bottom";
  ctx.fillStyle = "rgba(0,0,0,0.7)";
  ctx.fillText(name, x + 1, y - PLAYER_RADIUS - 3);
  ctx.fillStyle = "#fff";
  ctx.fillText(name, x, y - PLAYER_RADIUS - 4);
}

function drawViewCone(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  yawDegrees: number,
  color: string
) {
  // Align yaw (convert CS2 degrees to canvas radians)
  // CS2: east=0 degrees, counterclockwise
  // Canvas: east=0 radians, clockwise
  const yawRad = (-yawDegrees * Math.PI) / 180;

  // Calculate tip point (at VIEW_CONE_LENGTH distance)
  const tipX = x + Math.cos(yawRad) * VIEW_CONE_LENGTH;
  const tipY = y + Math.sin(yawRad) * VIEW_CONE_LENGTH;

  // Calculate base points (near player)
  const baseRadius = 8;
  const leftX = x + Math.cos(yawRad - VIEW_CONE_ANGLE) * baseRadius;
  const leftY = y + Math.sin(yawRad - VIEW_CONE_ANGLE) * baseRadius;
  const rightX = x + Math.cos(yawRad + VIEW_CONE_ANGLE) * baseRadius;
  const rightY = y + Math.sin(yawRad + VIEW_CONE_ANGLE) * baseRadius;

  ctx.beginPath();
  ctx.moveTo(leftX, leftY);
  ctx.lineTo(tipX, tipY);
  ctx.lineTo(rightX, rightY);
  ctx.closePath();

  ctx.fillStyle = color + "80" // 50% opacity
  ctx.fill();
}

function drawX(ctx: CanvasRenderingContext2D, x: number, y: number, color: string) {
  const size = 5;
  ctx.strokeStyle = color;
  ctx.lineWidth = 3;
  ctx.beginPath();
  ctx.moveTo(x - size, y - size);
  ctx.lineTo(x + size, y + size);
  ctx.moveTo(x + size, y - size);
  ctx.lineTo(x - size, y + size);
  ctx.stroke();
}