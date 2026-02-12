import type { MapConfig } from "../types/match";

export interface Point {
  x: number;
  y: number;
}

export function worldToRadar(worldX: number, worldY: number, config: MapConfig): Point {
  return {
    x: (worldX - config.posX) / config.scale,
    y: (config.posY - worldY) / config.scale,
  };
}

export function radarToCanvas(
  radarX: number,
  radarY: number,
  config: MapConfig,
  canvasWidth: number,
  canvasHeight: number
): Point {
  return {
    x: (radarX / config.radarWidth) * canvasWidth,
    y: (radarY / config.radarHeight) * canvasHeight,
  };
}

// CS2 world xy -> CS2 radar xy -> canvas xy
export function worldToCanvas(
  worldX: number,
  worldY: number,
  config: MapConfig,
  canvasWidth: number,
  canvasHeight: number
): Point {
  const radar = worldToRadar(worldX, worldY, config);
  return radarToCanvas(radar.x, radar.y, config, canvasWidth, canvasHeight);
}