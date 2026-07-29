/**
 * TAI Protocol - Bot Services Layer
 * Handles business logic between TG commands and backend API / TON chain.
 */

const API_BASE = process.env.API_BASE_URL || "http://localhost:8080/api/v1";

export class PetService {
  /** Claim starter pet for new user */
  async claimPet(tgUserId: number): Promise<any> {
    // TODO: POST /pet/claim with tg auth
    return { id: "TODO", name: "未命名", quality: "N", species: "cat" };
  }

  /** Get pet status card (formatted for TG message) */
  async getPetCard(petId: string): Promise<string> {
    // TODO: GET /pet/:id → format as text card
    return [
      "🤖 铁甲龙 Lv.5",
      "━━━━━━━━━━━━━━",
      "品质: SSR | 代数: 0 (创世)",
      "智力: 87 | 攻击: 65 | 速度: 72",
      "技能格: [审查之炮] [雷达阵列] [空] [空]",
      "状态: 💼 执行赏金任务中...",
      "今日收入: 3.2 USDT",
      "能量: ████████░░ 80%",
    ].join("\n");
  }

  /** Rename pet */
  async renamePet(petId: string, name: string): Promise<boolean> {
    // TODO: PUT /pet/:id/name
    return true;
  }
}

export class MarketService {
  /** Get market overview for display */
  async getMarketSummary(): Promise<string> {
    // TODO: GET /market/kline + /market/ranking
    return [
      "📈 TAI 市场概览",
      "━━━━━━━━━━━━━━",
      "24h 成交额: 1,247 TON",
      "宠物均价: 3.2 TON (↑12%)",
      "今日最高: 暗影龙 #001 → 25 TON",
      "在售: 47 只宠物 | 23 本兽决",
    ].join("\n");
  }
}

export class BountyService {
  /** Get available bounties matching user's pets */
  async getAvailableBounties(userId: string): Promise<any[]> {
    // TODO: GET /bounty/available
    return [];
  }

  /** Auto-assign bounty to best matching pet */
  async autoAssign(userId: string, bountyId: string): Promise<any> {
    // TODO: POST /bounty/:id/accept
    return { ok: true };
  }
}

export class EarningsService {
  /** Get daily earnings report */
  async getDailyReport(userId: string): Promise<string> {
    // TODO: GET /user/me/earnings
    return [
      "💰 今日收益报告",
      "━━━━━━━━━━━━━━",
      "赏金任务: +2.4 USDT (2笔)",
      "广告收入: +0.8 USDT",
      "算力消耗: -0.6 USDT (3api.shop)",
      "━━━━━━━━━━━━━━",
      "净收入: +2.6 USDT → 已入账",
    ].join("\n");
  }
}
