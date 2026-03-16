import { describe, it, expect } from "vitest";
import { formatTime, ROUND_TIME } from "./format";

describe("formatTime", () => {
  it("formats whole minutes", () => {
    expect(formatTime(60)).toBe("1:00");
    expect(formatTime(120)).toBe("2:00");
  });

  it("zero-pads seconds", () => {
    expect(formatTime(5)).toBe("0:05");
    expect(formatTime(61)).toBe("1:01");
  });

  it("floors fractional seconds", () => {
    expect(formatTime(59.9)).toBe("0:59");
  });

  it("clamps negative values to 0:00", () => {
    expect(formatTime(-1)).toBe("0:00");
    expect(formatTime(-100)).toBe("0:00");
  });

  it("formats the full round time", () => {
    expect(formatTime(ROUND_TIME)).toBe("1:55");
  });
});
