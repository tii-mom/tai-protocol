import "dotenv/config";

import { getPersonas } from "./personas.js";
import {
  buildMarketState,
  generateTickPlan,
  INITIAL_PRICE,
  roundPrice,
  toBackendOrderPayload,
} from "./strategy.js";
import type { BackendOrderPayload, Persona, TickPlan } from "./types.js";

interface AppConfig {
  backendUrl: string;
  adminToken: string;
  tickIntervalMs: number;
  requestTimeoutMs: number;
  maxRetries: number;
  retryBaseDelayMs: number;
}

interface RuntimeState {
  config: AppConfig;
  personas: Persona[];
  dailyOpenDate: string;
  dailyOpenPrice: number | null;
  lastKnownPrice: number;
  stopped: boolean;
  timer: NodeJS.Timeout | null;
}

const DEFAULT_TICK_INTERVAL_MS = 5 * 60 * 1000;

async function startMarketMaker(): Promise<void> {
  const config = readConfig();
  const state: RuntimeState = {
    config,
    personas: getPersonas(),
    dailyOpenDate: getLocalDateKey(),
    dailyOpenPrice: null,
    lastKnownPrice: INITIAL_PRICE,
    stopped: false,
    timer: null,
  };

  const stop = () => {
    state.stopped = true;
    if (state.timer) clearTimeout(state.timer);
    console.log("[market-maker] 收到停止信号，已停止后续 tick");
  };

  process.once("SIGINT", stop);
  process.once("SIGTERM", stop);

  console.log(
    `[market-maker] 启动完成 backend=${config.backendUrl} tick=${config.tickIntervalMs}ms personas=${state.personas.length}`,
  );

  await runAndScheduleNext(state);
}

async function runAndScheduleNext(state: RuntimeState): Promise<void> {
  if (state.stopped) return;

  await runTick(state).catch((error: unknown) => {
    console.error(`[market-maker] tick 执行失败: ${formatError(error)}`);
  });

  if (!state.stopped) {
    state.timer = setTimeout(() => {
      void runAndScheduleNext(state);
    }, state.config.tickIntervalMs);
  }
}

async function runTick(state: RuntimeState): Promise<void> {
  const now = new Date();
  const dateKey = getLocalDateKey(now);
  if (dateKey !== state.dailyOpenDate) {
    state.dailyOpenDate = dateKey;
    state.dailyOpenPrice = null;
  }

  const currentPrice = await readCurrentPriceWithFallback(state);
  if (state.dailyOpenPrice === null) {
    state.dailyOpenPrice = currentPrice;
  }

  const marketState = buildMarketState(currentPrice, state.dailyOpenPrice);
  const plan = generateTickPlan(state.personas, marketState);
  logTick(now, plan);

  const payloads = plan.orders.map(toBackendOrderPayload);
  const results = await Promise.allSettled(
    payloads.map((payload) =>
      withRetry(
        `提交订单 ${payload.bot_name}/${payload.side}`,
        () => postOrder(state.config, payload),
        state.config,
      ),
    ),
  );

  results.forEach((result, index) => {
    const payload = payloads[index];
    if (result.status === "fulfilled") {
      console.log(
        `[market-maker] 订单成功 bot=${payload.bot_name} side=${payload.side} quantity=${payload.quantity} price=${payload.price.toFixed(8)}`,
      );
    } else {
      console.error(
        `[market-maker] 订单失败 bot=${payload.bot_name} side=${payload.side} quantity=${payload.quantity} price=${payload.price.toFixed(8)} error=${formatError(result.reason)}`,
      );
    }
  });
}

async function readCurrentPriceWithFallback(state: RuntimeState): Promise<number> {
  try {
    const price = await withRetry("读取当前价格", () => fetchCurrentPrice(state.config), state.config);
    state.lastKnownPrice = price;
    return price;
  } catch (error) {
    // 价格接口短暂不可用时沿用上一轮价格，避免单点故障导致进程退出。
    console.error(
      `[market-maker] 当前价格读取失败，使用上一轮价格 ${state.lastKnownPrice.toFixed(8)}: ${formatError(error)}`,
    );
    return state.lastKnownPrice;
  }
}

async function fetchCurrentPrice(config: AppConfig): Promise<number> {
  const payload = await requestJson<unknown>(config, "/api/v1/admin/price", { method: "GET" });
  const price = extractPrice(payload);
  if (price === null) {
    throw new Error(`价格接口返回缺少 price 字段: ${safeJson(payload)}`);
  }
  return roundPrice(price);
}

async function postOrder(config: AppConfig, payload: BackendOrderPayload): Promise<unknown> {
  return requestJson<unknown>(config, "/api/v1/market/order", {
    method: "POST",
    body: payload,
  });
}

async function requestJson<T>(
  config: AppConfig,
  path: string,
  options: { method: "GET" | "POST"; body?: unknown },
): Promise<T> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), config.requestTimeoutMs);

  try {
    const response = await fetch(`${config.backendUrl}${path}`, {
      method: options.method,
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${config.adminToken}`,
      },
      body: options.body === undefined ? undefined : JSON.stringify(options.body),
      signal: controller.signal,
    });
    const responseText = await response.text();

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${responseText.slice(0, 240)}`);
    }
    if (responseText.trim().length === 0) {
      return {} as T;
    }

    return JSON.parse(responseText) as T;
  } catch (error) {
    if (error instanceof SyntaxError) {
      throw new Error(`接口返回不是合法 JSON: ${error.message}`);
    }
    throw error;
  } finally {
    clearTimeout(timeout);
  }
}

async function withRetry<T>(label: string, action: () => Promise<T>, config: AppConfig): Promise<T> {
  let lastError: unknown;

  for (let attempt = 1; attempt <= config.maxRetries; attempt += 1) {
    try {
      return await action();
    } catch (error) {
      lastError = error;
      console.warn(`[market-maker] ${label} 失败 attempt=${attempt}/${config.maxRetries} error=${formatError(error)}`);
      if (attempt < config.maxRetries) {
        await delay(config.retryBaseDelayMs * attempt);
      }
    }
  }

  throw lastError instanceof Error ? lastError : new Error(String(lastError));
}

function logTick(now: Date, plan: TickPlan): void {
  const breakerStatus = [
    plan.canBuy ? null : "暂停买单",
    plan.canSell ? null : "暂停卖单",
  ].filter(Boolean).join(",");

  console.log(
    `[market-maker] tick=${now.toISOString()} ref=${plan.referencePrice.toFixed(8)} target=${plan.targetPrice.toFixed(8)} curve=${formatPct(plan.priceChangePct)} pressure=${plan.pressure.toFixed(2)} daily=${formatPct(plan.dailyChangePct)} orders=${plan.orders.length}${breakerStatus ? ` circuit=${breakerStatus}` : ""}`,
  );

  plan.orders.forEach((order) => {
    console.log(
      `[market-maker] 计划订单 bot=${order.persona.name} style=${order.persona.style} side=${order.side} quantity=${order.quantity} price=${order.price.toFixed(8)}`,
    );
  });
}

function extractPrice(payload: unknown): number | null {
  const directPrice = readNumericPrice(payload);
  if (directPrice !== null) return directPrice;
  if (!isRecord(payload)) return null;

  for (const key of ["price", "currentPrice", "current_price", "marketPrice", "market_price"]) {
    const price = readNumericPrice(payload[key]);
    if (price !== null) return price;
  }

  for (const key of ["data", "result", "market"]) {
    const price = extractPrice(payload[key]);
    if (price !== null) return price;
  }

  return null;
}

function readConfig(): AppConfig {
  const backendUrl = process.env.BACKEND_URL?.trim();
  const adminToken = process.env.ADMIN_TOKEN?.trim();

  if (!backendUrl) {
    throw new Error("缺少环境变量 BACKEND_URL");
  }
  if (!adminToken) {
    throw new Error("缺少环境变量 ADMIN_TOKEN");
  }

  return {
    backendUrl: backendUrl.replace(/\/+$/, ""),
    adminToken,
    tickIntervalMs: parsePositiveInteger(process.env.TICK_INTERVAL_MS, DEFAULT_TICK_INTERVAL_MS),
    requestTimeoutMs: 10_000,
    maxRetries: 3,
    retryBaseDelayMs: 800,
  };
}

function readNumericPrice(value: unknown): number | null {
  if (typeof value === "number" && Number.isFinite(value) && value > 0) {
    return value;
  }
  if (typeof value === "string") {
    const parsed = Number(value);
    return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
  }
  return null;
}

function parsePositiveInteger(value: string | undefined, fallback: number): number {
  if (value === undefined) return fallback;
  const parsed = Number(value);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback;
}

function getLocalDateKey(date: Date = new Date()): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function formatPct(value: number): string {
  return `${(value * 100).toFixed(2)}%`;
}

function formatError(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

function safeJson(value: unknown): string {
  try {
    return JSON.stringify(value).slice(0, 240);
  } catch {
    return String(value);
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

void startMarketMaker().catch((error: unknown) => {
  console.error(`[market-maker] 启动失败: ${formatError(error)}`);
  process.exitCode = 1;
});
