import { describe, it, expect } from "vitest";
import { lerp, clamp01, angleLerp } from "./interpolation";

describe("clamp01", () => {
  it("passes through values in [0, 1]", () => {
    expect(clamp01(0)).toBe(0);
    expect(clamp01(0.5)).toBe(0.5);
    expect(clamp01(1)).toBe(1);
  });

  it("clamps below zero", () => {
    expect(clamp01(-0.5)).toBe(0);
    expect(clamp01(-100)).toBe(0);
  });

  it("clamps above one", () => {
    expect(clamp01(1.5)).toBe(1);
    expect(clamp01(999)).toBe(1);
  });
});

describe("lerp", () => {
  it("returns start at t=0", () => {
    expect(lerp(10, 20, 0)).toBe(10);
  });

  it("returns end at t=1", () => {
    expect(lerp(10, 20, 1)).toBe(20);
  });

  it("returns midpoint at t=0.5", () => {
    expect(lerp(0, 100, 0.5)).toBe(50);
  });

  it("clamps t outside [0, 1]", () => {
    expect(lerp(0, 10, -1)).toBe(0);
    expect(lerp(0, 10, 2)).toBe(10);
  });
});

describe("angleLerp", () => {
  it("interpolates between two angles", () => {
    expect(angleLerp(0, 90, 0.5)).toBeCloseTo(45);
  });

  it("takes the shortest path across 360/0 boundary", () => {
    // 360 and 0 are equivalent angles
    expect(angleLerp(350, 10, 0.5) % 360).toBeCloseTo(0);
  });

  it("takes the shortest path in the other direction", () => {
    expect(angleLerp(10, 350, 0.5)).toBeCloseTo(0);
  });

  it("returns start at t=0", () => {
    expect(angleLerp(45, 270, 0)).toBe(45);
  });

  it("returns end at t=1", () => {
    // -90 and 270 are equivalent angles
    expect(((angleLerp(45, 270, 1) % 360) + 360) % 360).toBeCloseTo(270);
  });
});
