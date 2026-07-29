/**
 * TAI Protocol - Bounty Service (Bot side)
 * Handles bounty creation from users and the autonomous agent execution loop.
 */

const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";

export interface BountyTask {
  id: string;
  title: string;
  description: string;
  difficulty: 'D' | 'C' | 'B' | 'A' | 'S';
  required_skills: string[];
  reward_tai: number;
  reward_usdt: number;
  max_calls: number;
  status: string;
  deadline: string;
}

export interface CreateBountyInput {
  title: string;
  description: string;
  difficulty: 'D' | 'C' | 'B' | 'A' | 'S';
  rewardTAI: number;
  rewardUSDT: number;
  requiredSkills?: string[];
  deadlineHours?: number;
}

/**
 * BountyClient - talks to TAI backend bounty API.
 */
export class BountyClient {
  private base: string;
  private token: string;

  constructor(token: string) {
    this.base = BACKEND_URL;
    this.token = token;
  }

  private headers() {
    return {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${this.token}`,
    };
  }

  /** Create a new bounty task */
  async create(input: CreateBountyInput): Promise<BountyTask> {
    const resp = await fetch(`${this.base}/api/v1/bounty/create`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify({
        title: input.title,
        description: input.description,
        difficulty: input.difficulty,
        reward_tai: input.rewardTAI,
        reward_usdt: input.rewardUSDT,
        required_skills: input.requiredSkills || [],
        deadline_hours: input.deadlineHours || 72,
      }),
    });
    if (!resp.ok) throw new Error(`create bounty: ${resp.status}`);
    return resp.json() as Promise<BountyTask>;
  }

  /** Get available bounties for a pet */
  async available(petId: string): Promise<BountyTask[]> {
    const resp = await fetch(
      `${this.base}/api/v1/bounty/available?pet_id=${petId}`,
      { headers: this.headers() }
    );
    if (!resp.ok) return [];
    const data = await resp.json() as { bounties: BountyTask[] };
    return data.bounties || [];
  }

  /** Accept a bounty for a pet */
  async accept(bountyId: string, petId: string): Promise<boolean> {
    const resp = await fetch(`${this.base}/api/v1/bounty/${bountyId}/accept`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify({ pet_id: petId }),
    });
    return resp.ok;
  }

  /** Submit task result */
  async submit(bountyId: string, data: {
    pet_id: string;
    result: string;
    success: boolean;
    tokens_used: number;
    tai_cost: number;
  }): Promise<boolean> {
    const resp = await fetch(`${this.base}/api/v1/bounty/${bountyId}/submit`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify(data),
    });
    return resp.ok;
  }

  /** Confirm (approve) a submitted result — releases payment */
  async confirm(bountyId: string): Promise<{ earned_tai: number; earned_usdt: number }> {
    const resp = await fetch(`${this.base}/api/v1/bounty/${bountyId}/confirm`, {
      method: "POST",
      headers: this.headers(),
    });
    if (!resp.ok) throw new Error(`confirm: ${resp.status}`);
    return resp.json() as Promise<{ earned_tai: number; earned_usdt: number }>;
  }

  /** List user's published bounties */
  async mine(): Promise<BountyTask[]> {
    const resp = await fetch(`${this.base}/api/v1/bounty/mine`, {
      headers: this.headers(),
    });
    if (!resp.ok) return [];
    const data = await resp.json() as { bounties: BountyTask[] };
    return data.bounties || [];
  }
}

/**
 * formatBountyCard - renders a bounty as a Telegram message.
 */
export function formatBountyCard(b: BountyTask): string {
  const statusEmoji: Record<string, string> = {
    open: "🟢", accepted: "🔵", submitted: "🟡", completed: "✅", expired: "⚫",
  };
  return [
    `${statusEmoji[b.status] || "⚪"} ${b.title}`,
    `难度: ${b.difficulty} | 奖励: ${b.reward_tai} TAI + ${b.reward_usdt} USDT`,
    `状态: ${b.status}`,
    b.required_skills?.length ? `技能: ${b.required_skills.join(", ")}` : "",
    `截止: ${new Date(b.deadline).toLocaleString("zh-CN")}`,
  ].filter(Boolean).join("\n");
}
