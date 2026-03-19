export const ROUND_TIME = 115;

export function formatTime(seconds: number): string {
  const clamped = Math.max(0, seconds);
  const m = Math.floor(clamped / 60);
  const s = Math.floor(clamped % 60);
  return `${m}:${String(s).padStart(2, "0")}`;
}

export function grenadeLabel(name: string): string {
  const lower = name.toLowerCase();
  if (lower.includes("flash")) return "F";
  if (lower.includes("smoke")) return "S";
  if (lower.includes("he")) return "HE";
  if (lower.includes("molotov") || lower.includes("incendiary")) return "M";
  if (lower.includes("decoy")) return "D";
  return name;
}

export function armorLabel(armor: number, hasHelmet: boolean): string {
  if (armor <= 0) return "";
  return hasHelmet ? "A+H" : "A";
}