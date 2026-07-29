export type OrderSide = "buy" | "sell";

export type PersonaStyle = "aggressive" | "conservative" | "random";

export interface Persona {
  name: string;
  style: PersonaStyle;
  bias: number;
}

export interface StrategyConfig {
  basePrice: number;
  minPriceChangePct: number;
  maxPriceChangePct: number;
  buyRatio: number;
  minOrdersPerTick: number;
  maxOrdersPerTick: number;
  minQuantity: number;
  maxQuantity: number;
  maxDailyGain: number;
  maxDailyDrop: number;
}

export interface MarketState {
  currentPrice: number;
  dailyOpenPrice: number;
  dailyChangePct: number;
  canBuy: boolean;
  canSell: boolean;
}

export interface GeneratedOrder {
  side: OrderSide;
  price: number;
  quantity: number;
  persona: Persona;
  item_type: "token";
  item_id: "TAI";
  currency: "USDT";
}

export interface TickPlan {
  referencePrice: number;
  targetPrice: number;
  priceChangePct: number;
  pressure: number;
  dailyChangePct: number;
  canBuy: boolean;
  canSell: boolean;
  orders: GeneratedOrder[];
}

export interface BackendOrderPayload {
  side: OrderSide;
  item_type: "token";
  item_id: "TAI";
  price: number;
  quantity: number;
  currency: "USDT";
  bot_name: string;
  bot_style: PersonaStyle;
}
