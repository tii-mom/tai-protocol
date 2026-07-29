import type {
  BackendOrderPayload,
  GeneratedOrder,
  MarketState,
  OrderSide,
  Persona,
  PersonaStyle,
  StrategyConfig,
  TickPlan,
} from "./types.js";

export const DEFAULT_STRATEGY_CONFIG: StrategyConfig = {
  basePrice: 0.001,
  minPriceChangePct: 0.02,
  maxPriceChangePct: 0.05,
  buyRatio: 0.7,
  minOrdersPerTick: 3,
  maxOrdersPerTick: 5,
  minQuantity: 50,
  maxQuantity: 500,
  maxDailyGain: 1.5,
  maxDailyDrop: 0.6,
};

export const INITIAL_PRICE = DEFAULT_STRATEGY_CONFIG.basePrice;

const MIN_PRICE = 0.00000001;

export function roundPrice(price: number): number {
  return Number(Math.max(price, MIN_PRICE).toFixed(8));
}

export function buildMarketState(
  currentPrice: number,
  dailyOpenPrice: number,
  config: StrategyConfig = DEFAULT_STRATEGY_CONFIG,
): MarketState {
  const safeCurrentPrice = roundPrice(Number.isFinite(currentPrice) && currentPrice > 0 ? currentPrice : config.basePrice);
  const safeOpenPrice = roundPrice(Number.isFinite(dailyOpenPrice) && dailyOpenPrice > 0 ? dailyOpenPrice : config.basePrice);
  const dailyChangePct = (safeCurrentPrice - safeOpenPrice) / safeOpenPrice;

  return {
    currentPrice: safeCurrentPrice,
    dailyOpenPrice: safeOpenPrice,
    dailyChangePct,
    // 单日跌幅超过 60% 时暂停买单。
    canBuy: dailyChangePct >= -config.maxDailyDrop,
    // 单日涨幅超过 150% 时暂停卖单。
    canSell: dailyChangePct <= config.maxDailyGain,
  };
}

export function generateTickPlan(
  personas: Persona[],
  marketState: MarketState,
  config: StrategyConfig = DEFAULT_STRATEGY_CONFIG,
): TickPlan {
  const orderCount = randomInt(config.minOrdersPerTick, config.maxOrdersPerTick);
  const sides = buildOrderSides(orderCount, marketState.canBuy, marketState.canSell, config);
  const pressure = calculatePressure(sides);
  const priceChangePct = calculatePriceChangePct(pressure, config);
  const targetPrice = roundPrice(marketState.currentPrice * (1 + priceChangePct));
  const shuffledPersonas = shuffle(personas);

  if (sides.length === 0 || shuffledPersonas.length === 0) {
    return {
      referencePrice: marketState.currentPrice,
      targetPrice,
      priceChangePct,
      pressure,
      dailyChangePct: marketState.dailyChangePct,
      canBuy: marketState.canBuy,
      canSell: marketState.canSell,
      orders: [],
    };
  }

  const orders = sides.map((side, index) => {
    const persona = shuffledPersonas[index % shuffledPersonas.length];
    return createOrder(side, persona, targetPrice, config);
  });

  return {
    referencePrice: marketState.currentPrice,
    targetPrice,
    priceChangePct,
    pressure,
    dailyChangePct: marketState.dailyChangePct,
    canBuy: marketState.canBuy,
    canSell: marketState.canSell,
    orders,
  };
}

export function toBackendOrderPayload(order: GeneratedOrder): BackendOrderPayload {
  return {
    side: order.side,
    item_type: order.item_type,
    item_id: order.item_id,
    price: order.price,
    quantity: order.quantity,
    currency: order.currency,
    bot_name: order.persona.name,
    bot_style: order.persona.style,
  };
}

function buildOrderSides(
  orderCount: number,
  canBuy: boolean,
  canSell: boolean,
  config: StrategyConfig,
): OrderSide[] {
  if (!canBuy && !canSell) return [];
  if (canBuy && !canSell) return Array.from({ length: orderCount }, () => "buy");
  if (!canBuy && canSell) return Array.from({ length: orderCount }, () => "sell");

  const buyCount = Math.max(1, Math.min(orderCount - 1, Math.round(orderCount * config.buyRatio)));
  const sellCount = orderCount - buyCount;
  return shuffle([
    ...Array.from({ length: buyCount }, () => "buy" as const),
    ...Array.from({ length: sellCount }, () => "sell" as const),
  ]);
}

function calculatePressure(sides: OrderSide[]): number {
  if (sides.length === 0) return 0;
  const buyCount = sides.filter((side) => side === "buy").length;
  const sellCount = sides.length - buyCount;
  return (buyCount - sellCount) / sides.length;
}

function calculatePriceChangePct(pressure: number, config: StrategyConfig): number {
  if (pressure === 0) return 0;
  const direction = pressure > 0 ? 1 : -1;
  const magnitude = randomFloat(config.minPriceChangePct, config.maxPriceChangePct);
  return direction * magnitude;
}

function createOrder(
  side: OrderSide,
  persona: Persona,
  targetPrice: number,
  config: StrategyConfig,
): GeneratedOrder {
  const quantity = randomInt(config.minQuantity, config.maxQuantity);
  const priceOffset = getPersonaOffset(persona.style, side) + clamp(persona.bias, -0.1, 0.1);

  return {
    side,
    price: roundPrice(targetPrice * (1 + priceOffset)),
    quantity,
    persona,
    item_type: "token",
    item_id: "TAI",
    currency: "USDT",
  };
}

function getPersonaOffset(style: PersonaStyle, side: OrderSide): number {
  switch (style) {
    case "aggressive": {
      // 激进型更愿意贴近成交方向：买单略高，卖单略低。
      const offset = randomFloat(0, 0.02);
      return side === "buy" ? offset : -offset;
    }
    case "conservative": {
      // 保守型拉开价差：买单更低，卖单更高。
      const offset = randomFloat(0.015, 0.04);
      return side === "buy" ? -offset : offset;
    }
    case "random":
      // 随机型制造噪声，让盘口看起来不那么机械。
      return randomFloat(-0.035, 0.035);
  }
}

function randomInt(min: number, max: number): number {
  const lower = Math.ceil(min);
  const upper = Math.floor(max);
  return Math.floor(Math.random() * (upper - lower + 1)) + lower;
}

function randomFloat(min: number, max: number): number {
  return Math.random() * (max - min) + min;
}

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value));
}

function shuffle<T>(items: T[]): T[] {
  const result = [...items];
  for (let index = result.length - 1; index > 0; index -= 1) {
    const swapIndex = randomInt(0, index);
    [result[index], result[swapIndex]] = [result[swapIndex], result[index]];
  }
  return result;
}
