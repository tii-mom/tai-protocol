import type { Persona } from "./types.js";

const PERSONAS: Persona[] = [
  { name: "赤羽", style: "aggressive", bias: 0.08 },
  { name: "青石", style: "conservative", bias: -0.03 },
  { name: "沐风", style: "random", bias: 0.01 },
  { name: "海棠", style: "conservative", bias: 0.04 },
  { name: "北辰", style: "aggressive", bias: 0.06 },
  { name: "银沙", style: "random", bias: -0.06 },
  { name: "星尘", style: "aggressive", bias: 0.02 },
  { name: "雾灯", style: "conservative", bias: -0.08 },
];

export function getPersonas(): Persona[] {
  // 返回副本，避免运行中意外改写静态人设。
  return PERSONAS.map((persona) => ({ ...persona }));
}
