export const PLAYER_RADIUS = 6;
export const FONT = "700 11px Stratum2, sans-serif";
const VIEW_CONE_LENGTH = 10;
const VIEW_CONE_ANGLE = Math.PI / 6;

const MAX_NAME_LENGTH = 16;
const NAME_PADDING_X = 4;
const NAME_PADDING_Y = 2;
const NAME_BG_RADIUS = 3;
const NAME_BG_OPACITY = "40";
const NAME_OFFSET_Y = PLAYER_RADIUS + 6;

function truncateName(name: string): string {
  if (name.length <= MAX_NAME_LENGTH) return name;
  return name.slice(0, MAX_NAME_LENGTH - 1) + "\u2026";
}

export function drawPlayerDot(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  color: string,
) {
  ctx.beginPath();
  ctx.arc(x, y, PLAYER_RADIUS, 0, Math.PI * 2);
  ctx.fillStyle = color;
  ctx.fill();
  ctx.lineWidth = 1.5;
  ctx.strokeStyle = "rgba(0,0,0,0.5)";
  ctx.stroke();
}

export function drawPlayerName(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  name: string,
  color: string,
) {
  const label = truncateName(name);
  ctx.font = FONT;
  ctx.textAlign = "center";
  ctx.textBaseline = "bottom";

  const metrics = ctx.measureText(label);
  const textW = metrics.width;
  const textH = 11;
  const bgW = textW + NAME_PADDING_X * 2;
  const bgH = textH + NAME_PADDING_Y * 2;
  const bgX = x - bgW / 2;
  const bgY = y - NAME_OFFSET_Y - bgH;

  ctx.beginPath();
  ctx.roundRect(bgX, bgY, bgW, bgH, NAME_BG_RADIUS);
  ctx.fillStyle = color + NAME_BG_OPACITY;
  ctx.fill();

  ctx.fillStyle = "#fff";
  ctx.fillText(label, x, y - NAME_OFFSET_Y);
}

export function drawViewCone(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  yawDegrees: number,
  color: string,
) {
  const yawRad = (-yawDegrees * Math.PI) / 180;

  const tipX = x + Math.cos(yawRad) * VIEW_CONE_LENGTH;
  const tipY = y + Math.sin(yawRad) * VIEW_CONE_LENGTH;

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

  ctx.fillStyle = color + "80"; // 50% opacity
  ctx.fill();
}

export function drawX(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  color: string,
) {
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
