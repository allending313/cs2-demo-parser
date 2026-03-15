import { describe, it, expect } from "vitest";
import { truncateName, getGrenadeColor } from "./draw";
import type { GrenadeType, Team } from "../types/match";

describe("truncateName", () => {
  it("returns short names unchanged", () => {
    expect(truncateName("Player1")).toBe("Player1");
  });

  it("returns names at the max length unchanged", () => {
    const name = "A".repeat(16);
    expect(truncateName(name)).toBe(name);
  });

  it("truncates long names with an ellipsis", () => {
    const name = "A".repeat(20);
    expect(truncateName(name)).toBe("A".repeat(15) + "\u2026");
  });

  it("handles empty string", () => {
    expect(truncateName("")).toBe("");
  });
});

describe("getGrenadeColor", () => {
  it("returns team-specific color for smoke grenades", () => {
    expect(getGrenadeColor("smoke", "ct")).toBe("#a2c6ff");
    expect(getGrenadeColor("smoke", "t")).toBe("#ffdf93");
  });

  it("returns team-specific color for fire grenades", () => {
    const fireTypes: GrenadeType[] = ["molotov", "incendiary"];
    for (const type of fireTypes) {
      expect(getGrenadeColor(type, "ct")).toBe("#6ea8ff");
      expect(getGrenadeColor(type, "t")).toBe("#ff8c3a");
    }
  });

  it("returns default color when team is null", () => {
    expect(getGrenadeColor("smoke", null)).toBe("#cccccc");
    expect(getGrenadeColor("he", null)).toBe("#ff4444");
  });

  it("returns default color for non-team-colored grenades regardless of team", () => {
    const teams: Team[] = ["ct", "t"];
    for (const team of teams) {
      expect(getGrenadeColor("flash", team)).toBe("#ffd700");
      expect(getGrenadeColor("he", team)).toBe("#ff4444");
      expect(getGrenadeColor("decoy", team)).toBe("#888888");
    }
  });
});
