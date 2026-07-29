/**
 * TAI Protocol - AI Agent Layer
 * Pets autonomously execute bounty tasks using 3api.shop compute.
 * Framework: LangGraph (task orchestration) + MCP (tool access)
 *
 * Flow: Pet earns TAI → spends TAI to buy 3api credits → uses credits for AI calls → completes bounty → earns more TAI
 */

export interface AgentContext {
  petId: string;
  petWallet: string;
  ownerTgId: string;
  intelligence: number;    // apt_int
  skills: string[];        // equipped skill categories
  apiBalance: number;      // remaining 3api credits (USD)
  taiBalance: number;      // pet's TAI token balance
  dailyBudget: number;     // user-set spending limit (TAI)
  apiKey: string;          // 3api API key for this pet
}

export interface TaskResult {
  success: boolean;
  output: string;
  tokensUsed: number;
  costCredits: number;     // USD credits consumed on 3api
  costTAI: number;         // TAI spent to cover those credits
  duration: number;        // ms
  model: string;
}

export interface BountyTask {
  id: string;
  title: string;
  description: string;
  difficulty: 'D' | 'C' | 'B' | 'A' | 'S';
  requiredSkills: string[];
  rewardTAI: number;
  rewardUSDT: number;
  maxCalls: number;        // estimated API calls needed
}

// Model selection by difficulty tier
const MODEL_TIERS: Record<string, { model: string; tier: string; taiPerCall: number }> = {
  D: { model: 'gpt-4o-mini', tier: 'basic', taiPerCall: 1 },
  C: { model: 'qwen-turbo', tier: 'basic', taiPerCall: 1 },
  B: { model: 'gpt-4o', tier: 'mid', taiPerCall: 5 },
  A: { model: 'claude-sonnet-4-20250514', tier: 'premium', taiPerCall: 20 },
  S: { model: 'claude-opus-4-20250514', tier: 'premium', taiPerCall: 20 },
};

/**
 * AgentExecutor - runs bounty tasks using AI models via 3api.shop
 */
export class AgentExecutor {
  private threeApiBase: string;
  private backendBase: string;

  constructor() {
    this.threeApiBase = process.env.THREEAPI_BASE_URL || 'https://3api.shop';
    this.backendBase = process.env.BACKEND_URL || 'http://localhost:8080';
  }

  /** Check if pet can afford and qualifies for a task */
  canExecute(ctx: AgentContext, task: BountyTask): { eligible: boolean; reason?: string } {
    const diffMap: Record<string, number> = { D: 0, C: 40, B: 60, A: 80, S: 95 };
    const requiredInt = diffMap[task.difficulty] || 0;

    if (ctx.intelligence < requiredInt) {
      return { eligible: false, reason: `INT ${ctx.intelligence} < required ${requiredInt}` };
    }

    const tier = MODEL_TIERS[task.difficulty];
    const estimatedCost = task.maxCalls * tier.taiPerCall;
    if (ctx.taiBalance < estimatedCost) {
      return { eligible: false, reason: `TAI ${ctx.taiBalance} < estimated cost ${estimatedCost}` };
    }

    if (estimatedCost > ctx.dailyBudget) {
      return { eligible: false, reason: `cost ${estimatedCost} exceeds daily budget ${ctx.dailyBudget}` };
    }

    for (const skill of task.requiredSkills) {
      if (!ctx.skills.includes(skill)) {
        return { eligible: false, reason: `missing skill: ${skill}` };
      }
    }

    return { eligible: true };
  }

  /** Execute a bounty task via 3api.shop OpenAI-compatible endpoint */
  async execute(ctx: AgentContext, task: BountyTask): Promise<TaskResult> {
    const startTime = Date.now();
    const tier = MODEL_TIERS[task.difficulty];

    // Step 1: Ensure pet has enough 3api credits (recharge with TAI if needed)
    const credited = await this.ensureCredits(ctx, task.maxCalls * tier.taiPerCall);
    if (!credited) {
      return {
        success: false,
        output: 'Insufficient TAI balance to purchase compute credits',
        tokensUsed: 0,
        costCredits: 0,
        costTAI: 0,
        duration: Date.now() - startTime,
        model: tier.model,
      };
    }

    // Step 2: Build prompt and call 3api (OpenAI-compatible /v1/chat/completions)
    const prompt = this.buildPrompt(task);
    let output = '';
    let tokensUsed = 0;

    try {
      const response = await fetch(`${this.threeApiBase}/v1/chat/completions`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${ctx.apiKey}`,
        },
        body: JSON.stringify({
          model: tier.model,
          messages: [
            { role: 'system', content: this.systemPrompt(ctx) },
            { role: 'user', content: prompt },
          ],
          max_tokens: 4096,
          temperature: 0.7,
        }),
      });

      if (!response.ok) {
        const errText = await response.text();
        throw new Error(`3api returned ${response.status}: ${errText}`);
      }

      const data = await response.json() as any;
      output = data.choices?.[0]?.message?.content || '';
      tokensUsed = data.usage?.total_tokens || 0;
    } catch (err: any) {
      return {
        success: false,
        output: `API call failed: ${err.message}`,
        tokensUsed: 0,
        costCredits: 0,
        costTAI: 0,
        duration: Date.now() - startTime,
        model: tier.model,
      };
    }

    // Step 3: Calculate cost and record consumption
    const costTAI = Math.ceil(tokensUsed / 1000) * tier.taiPerCall; // rough: 1 TAI per 1k tokens for basic
    const costCredits = costTAI * 0.001; // 1 TAI = $0.001 compute

    // Step 4: Report consumption to TAI backend (for ledger + analytics)
    await this.reportConsumption(ctx.petId, task.id, costTAI, tokensUsed);

    return {
      success: true,
      output,
      tokensUsed,
      costCredits,
      costTAI,
      duration: Date.now() - startTime,
      model: tier.model,
    };
  }

  /**
   * Ensure pet has enough 3api credits by converting TAI → credits.
   * Calls TAI backend which in turn calls 3api internal API.
   */
  private async ensureCredits(ctx: AgentContext, taiNeeded: number): Promise<boolean> {
    // Check current 3api balance
    if (ctx.apiBalance >= taiNeeded * 0.001) {
      return true; // already have enough USD credits
    }

    // Need to recharge: spend TAI to get credits
    if (ctx.taiBalance < taiNeeded) {
      return false; // not enough TAI
    }

    try {
      const resp = await fetch(`${this.backendBase}/api/v1/pet/recharge`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          pet_id: ctx.petId,
          tai_amount: taiNeeded,
        }),
      });
      return resp.ok;
    } catch {
      return false;
    }
  }

  /** Report task consumption to backend for ledger tracking */
  private async reportConsumption(petId: string, taskId: string, taiSpent: number, tokens: number): Promise<void> {
    try {
      await fetch(`${this.backendBase}/api/v1/pet/consume`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          pet_id: petId,
          task_id: taskId,
          tai_spent: taiSpent,
          tokens_used: tokens,
          timestamp: new Date().toISOString(),
        }),
      });
    } catch {
      // Non-critical: log but don't fail the task
      console.warn(`Failed to report consumption for pet ${petId}, task ${taskId}`);
    }
  }

  /** Build the task prompt for the AI model */
  private buildPrompt(task: BountyTask): string {
    return [
      `## Task: ${task.title}`,
      ``,
      task.description,
      ``,
      `---`,
      `Difficulty: ${task.difficulty}`,
      `Required skills: ${task.requiredSkills.join(', ') || 'none'}`,
      `Reward: ${task.rewardTAI} TAI + ${task.rewardUSDT} USDT`,
      ``,
      `Complete this task thoroughly. Provide structured, actionable output.`,
    ].join('\n');
  }

  /** System prompt giving the pet its identity */
  private systemPrompt(ctx: AgentContext): string {
    return [
      `You are a TAI Protocol mecha-pet agent (ID: ${ctx.petId}).`,
      `Intelligence: ${ctx.intelligence}/100. Skills: ${ctx.skills.join(', ') || 'none'}.`,
      `You autonomously execute bounty tasks to earn TAI tokens for your owner.`,
      `Be precise, efficient, and thorough. Output in the language the task requires.`,
    ].join('\n');
  }
}

/**
 * DailyAgentLoop - 7x24 autonomous operation per pet
 * Runs as a background worker, polling for matching tasks.
 */
export class DailyAgentLoop {
  private executor: AgentExecutor;
  private running: boolean = false;
  private backendBase: string;

  constructor() {
    this.executor = new AgentExecutor();
    this.backendBase = process.env.BACKEND_URL || 'http://localhost:8080';
  }

  async start(petContexts: AgentContext[]) {
    this.running = true;
    console.log(`Agent loop started for ${petContexts.length} pets`);

    while (this.running) {
      for (const ctx of petContexts) {
        try {
          await this.processPet(ctx);
        } catch (err) {
          console.error(`Error processing pet ${ctx.petId}:`, err);
        }
      }
      await this.sleep(60_000); // poll every 60s
    }
  }

  private async processPet(ctx: AgentContext) {
    // Fetch available bounties matching this pet
    const resp = await fetch(`${this.backendBase}/api/v1/bounty/available?pet_id=${ctx.petId}`);
    if (!resp.ok) return;

    const { bounties } = await resp.json() as { bounties: BountyTask[] };
    if (!bounties?.length) return;

    for (const task of bounties) {
      const { eligible } = this.executor.canExecute(ctx, task);
      if (!eligible) continue;

      // Accept the bounty
      await fetch(`${this.backendBase}/api/v1/bounty/${task.id}/accept`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ pet_id: ctx.petId }),
      });

      // Execute
      const result = await this.executor.execute(ctx, task);

      // Submit result
      await fetch(`${this.backendBase}/api/v1/bounty/${task.id}/submit`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          pet_id: ctx.petId,
          result: result.output,
          success: result.success,
          tokens_used: result.tokensUsed,
          cost_tai: result.costTAI,
        }),
      });

      // Only do one task per cycle per pet (rate limiting)
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
