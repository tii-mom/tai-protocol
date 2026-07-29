/**
 * TAI Protocol - AI Agent Executor (Phase 0: Platform Pool Model)
 *
 * Architecture:
 *   Pet → TAI Backend (proxy + TAI deduction) → 3api.shop (platform key)
 *
 * The bot NEVER calls 3api directly. All compute goes through the TAI backend
 * which handles: auth, TAI balance check, proxy to 3api, usage recording.
 * This keeps the economic loop entirely within TAI's control.
 */

export interface AgentContext {
  petId: string;
  ownerTgId: string;
  intelligence: number;    // apt_int (0-100)
  skills: string[];        // equipped skill categories
  taiBalance: number;      // pet's TAI balance (off-chain ledger)
  dailyBudget: number;     // user-set daily spending limit (TAI)
  dailySpent: number;      // TAI spent today
}

export interface BountyTask {
  id: string;
  title: string;
  description: string;
  difficulty: 'D' | 'C' | 'B' | 'A' | 'S';
  requiredSkills: string[];
  rewardTAI: number;
  rewardUSDT: number;
  estimatedTokens: number;  // expected token usage
}

export interface TaskResult {
  success: boolean;
  output: string;
  tokensUsed: number;
  taiCost: number;
  durationMs: number;
  model: string;
}

// Difficulty → model mapping (mirrors backend threeapi/exchange.go)
const DIFFICULTY_MODEL: Record<string, string> = {
  D: 'gpt-4o-mini',
  C: 'qwen-turbo',
  B: 'gpt-4o',
  A: 'claude-sonnet-4-20250514',
  S: 'claude-opus-4-20250514',
};

// Rough TAI cost per 1000 tokens by difficulty
const TAI_PER_1K: Record<string, number> = {
  D: 1.0, C: 0.8, B: 5.0, A: 5.0, S: 20.0,
};

/**
 * AgentExecutor - executes bounty tasks via TAI backend proxy.
 * The backend handles 3api communication and TAI deduction.
 */
export class AgentExecutor {
  private backendBase: string;

  constructor() {
    this.backendBase = process.env.BACKEND_URL || 'http://localhost:8080';
  }

  /** Check if pet can afford and qualifies for a task */
  canExecute(ctx: AgentContext, task: BountyTask): { eligible: boolean; reason?: string } {
    // Intelligence gate
    const diffMap: Record<string, number> = { D: 0, C: 40, B: 60, A: 80, S: 95 };
    const requiredInt = diffMap[task.difficulty] || 0;
    if (ctx.intelligence < requiredInt) {
      return { eligible: false, reason: `INT ${ctx.intelligence} < ${requiredInt}` };
    }

    // Skill gate
    for (const skill of task.requiredSkills) {
      if (!ctx.skills.includes(skill)) {
        return { eligible: false, reason: `missing skill: ${skill}` };
      }
    }

    // Cost estimation
    const rate = TAI_PER_1K[task.difficulty] || 1.0;
    const estimatedCost = (task.estimatedTokens / 1000) * rate;

    if (ctx.taiBalance < estimatedCost) {
      return { eligible: false, reason: `TAI ${ctx.taiBalance.toFixed(1)} < cost ${estimatedCost.toFixed(1)}` };
    }

    if (ctx.dailySpent + estimatedCost > ctx.dailyBudget) {
      return { eligible: false, reason: `daily budget exceeded` };
    }

    return { eligible: true };
  }

  /**
   * Execute a bounty task.
   * Calls TAI backend /api/v1/pet/execute which:
   *   1. Verifies pet's TAI balance
   *   2. Proxies to 3api with platform key
   *   3. Deducts TAI from pet's ledger
   *   4. Returns AI response + cost
   */
  async execute(ctx: AgentContext, task: BountyTask): Promise<TaskResult> {
    const startTime = Date.now();
    const model = DIFFICULTY_MODEL[task.difficulty] || 'gpt-4o-mini';

    try {
      const resp = await fetch(`${this.backendBase}/api/v1/pet/execute`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          pet_id: ctx.petId,
          task_id: task.id,
          model,
          messages: [
            { role: 'system', content: this.systemPrompt(ctx) },
            { role: 'user', content: this.taskPrompt(task) },
          ],
          max_tokens: 4096,
          temperature: 0.7,
        }),
      });

      if (!resp.ok) {
        const err = await resp.text();
        return {
          success: false,
          output: `Backend error ${resp.status}: ${err}`,
          tokensUsed: 0,
          taiCost: 0,
          durationMs: Date.now() - startTime,
          model,
        };
      }

      const data = await resp.json() as {
        content: string;
        tokens_used: number;
        tai_cost: number;
        model: string;
      };

      return {
        success: true,
        output: data.content,
        tokensUsed: data.tokens_used,
        taiCost: data.tai_cost,
        durationMs: Date.now() - startTime,
        model: data.model,
      };
    } catch (err: any) {
      return {
        success: false,
        output: `Network error: ${err.message}`,
        tokensUsed: 0,
        taiCost: 0,
        durationMs: Date.now() - startTime,
        model,
      };
    }
  }

  private systemPrompt(ctx: AgentContext): string {
    return [
      `You are a TAI Protocol mecha-pet agent (ID: ${ctx.petId}).`,
      `Intelligence: ${ctx.intelligence}/100. Skills: ${ctx.skills.join(', ') || 'none'}.`,
      `You autonomously execute bounty tasks to earn TAI tokens for your owner.`,
      `Be precise, efficient, and thorough. Output in the language the task requires.`,
    ].join('\n');
  }

  private taskPrompt(task: BountyTask): string {
    return [
      `## Task: ${task.title}`,
      '',
      task.description,
      '',
      '---',
      `Difficulty: ${task.difficulty} | Skills: ${task.requiredSkills.join(', ') || 'none'}`,
      `Reward: ${task.rewardTAI} TAI + ${task.rewardUSDT} USDT`,
      '',
      'Complete this task thoroughly. Provide structured, actionable output.',
    ].join('\n');
  }
}

/**
 * DailyAgentLoop - 7x24 autonomous operation per pet.
 * Polls TAI backend for matching bounties, executes, submits results.
 */
export class DailyAgentLoop {
  private executor: AgentExecutor;
  private running = false;
  private backendBase: string;

  constructor() {
    this.executor = new AgentExecutor();
    this.backendBase = process.env.BACKEND_URL || 'http://localhost:8080';
  }

  async start(petContexts: AgentContext[]) {
    this.running = true;
    console.log(`[AgentLoop] started for ${petContexts.length} pets`);

    while (this.running) {
      for (const ctx of petContexts) {
        try {
          await this.processPet(ctx);
        } catch (err) {
          console.error(`[AgentLoop] pet ${ctx.petId} error:`, err);
        }
      }
      await this.sleep(60_000);
    }
  }

  private async processPet(ctx: AgentContext) {
    // Fetch available bounties for this pet
    const resp = await fetch(
      `${this.backendBase}/api/v1/bounty/available?pet_id=${ctx.petId}`
    );
    if (!resp.ok) return;

    const { bounties } = await resp.json() as { bounties: BountyTask[] };
    if (!bounties?.length) return;

    for (const task of bounties) {
      const { eligible } = this.executor.canExecute(ctx, task);
      if (!eligible) continue;

      // Accept
      await fetch(`${this.backendBase}/api/v1/bounty/${task.id}/accept`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ pet_id: ctx.petId }),
      });

      // Execute (backend proxies to 3api + deducts TAI)
      const result = await this.executor.execute(ctx, task);

      // Submit
      await fetch(`${this.backendBase}/api/v1/bounty/${task.id}/submit`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          pet_id: ctx.petId,
          result: result.output,
          success: result.success,
          tokens_used: result.tokensUsed,
          tai_cost: result.taiCost,
        }),
      });

      // Update local context
      ctx.taiBalance -= result.taiCost;
      ctx.dailySpent += result.taiCost;

      // One task per cycle per pet
      break;
    }
  }

  stop() {
    this.running = false;
  }

  private sleep(ms: number) {
    return new Promise((r) => setTimeout(r, ms));
  }
}
