export function getCSSColor(variable: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(variable).trim();
}