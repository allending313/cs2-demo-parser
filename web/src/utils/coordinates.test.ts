import { describe, it, expect } from "vitest";
import { worldToRadar, radarToCanvas, worldToCanvas } from "./coordinates";
import type { MapConfig } from "../types/match";

const config: MapConfig = {
  posX: -2476,
  posY: 3239,
  scale: 4.4,
  radarWidth: 1024,
  radarHeight: 1024,
};

describe("worldToRadar", () => {
  it("converts world origin to expected radar position", () => {
    const result = worldToRadar(config.posX, config.posY, config);
    expect(result).toEqual({ x: 0, y: 0 });
  });

  it("scales offset correctly", () => {
    const result = worldToRadar(config.posX + config.scale, config.posY, config);
    expect(result.x).toBeCloseTo(1);
    expect(result.y).toBeCloseTo(0);
  });

  it("inverts the y axis", () => {
    const result = worldToRadar(config.posX, config.posY - config.scale, config);
    expect(result.x).toBeCloseTo(0);
    expect(result.y).toBeCloseTo(1);
  });
});

describe("radarToCanvas", () => {
  it("maps radar origin to canvas origin", () => {
    const result = radarToCanvas(0, 0, config, 800, 800);
    expect(result).toEqual({ x: 0, y: 0 });
  });

  it("scales to canvas dimensions", () => {
    const result = radarToCanvas(config.radarWidth, config.radarHeight, config, 800, 600);
    expect(result).toEqual({ x: 800, y: 600 });
  });

  it("maps midpoint correctly", () => {
    const result = radarToCanvas(config.radarWidth / 2, config.radarHeight / 2, config, 1000, 1000);
    expect(result).toEqual({ x: 500, y: 500 });
  });
});

describe("worldToCanvas", () => {
  it("composes worldToRadar and radarToCanvas", () => {
    const radar = worldToRadar(0, 0, config);
    const expected = radarToCanvas(radar.x, radar.y, config, 800, 800);
    const result = worldToCanvas(0, 0, config, 800, 800);
    expect(result.x).toBeCloseTo(expected.x);
    expect(result.y).toBeCloseTo(expected.y);
  });
});
