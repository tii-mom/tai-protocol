import { Bot, Context, SessionFlavor, session } from "grammy";
import dotenv from "dotenv";
import { verifyWebAppInitData } from "./auth.js";

dotenv.config();

const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";

// Session: stores JWT token after auth
interface SessionData {
  step: "idle" | "naming";
  token: string | null;
  petId: string | null;
}

type BotContext = Context & SessionFlavor<SessionData> & { initData?: string };

const bot = new Bot<BotContext>(process.env.TG_BOT_TOKEN || "");

bot.use(session({ initial: (): SessionData => ({ step: "idle", token: null, petId: null }) }));

// ─── Auth helper ───────────────────────────────────────────────────

async function authenticate(ctx: BotContext): Promise<string | null> {
  const sess = ctx.session as SessionData;
  if (sess.token) return sess.token;

  const webAppInitData = ctx.initData;
  let initData: string;

  if (webAppInitData) {
    if (!verifyWebAppInitData(webAppInitData, process.env.TG_BOT_TOKEN || "")) {
      console.error("Invalid or expired WebApp initData");
      return null;
    }
    initData = webAppInitData;
  } else {
    initData = buildBotInitData(ctx);
  }

  try {
    const resp = await fetch(`${BACKEND_URL}/api/v1/user/auth/tg`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        init_data: initData,
      }),
    });

    if (!resp.ok) {
      console.error(`Auth failed: ${resp.status}`);
      return null;
    }

    const data = await resp.json() as { token: string; is_new: boolean };
    sess.token = data.token;
    return data.token;
  } catch (err) {
    console.error("Auth error:", err);
    return null;
  }
}

// Build a minimal initData-like payload for bot commands.
// In production Mini App, this comes from Telegram.WebApp.initData with real signature.
function buildBotInitData(ctx: Context): string {
  const user = ctx.from;
  if (!user) return "";
  const userObj = {
    id: user.id,
    first_name: user.first_name,
    last_name: user.last_name || "",
    username: user.username || "",
  };
  // For bot commands, we use a trusted internal format
  // Backend should accept this when request comes from the bot server
  return `user=${encodeURIComponent(JSON.stringify(userObj))}&auth_date=${Math.floor(Date.now() / 1000)}&hash=bot-trusted`;
}

// ─── Commands ──────────────────────────────────────────────────────

bot.command("start", async (ctx: BotContext) => {
  const tgUser = ctx.from;
  if (!tgUser) return;

  const token = await authenticate(ctx);
  if (!token) {
    await ctx.reply("认证失败，请稍后重试。");
    return;
  }

  // Claim starter pet
  try {
    const resp = await fetch(`${BACKEND_URL}/api/v1/pet/claim`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`,
      },
    });

    const data = await resp.json() as any;

    if (resp.status === 409) {
      // Already claimed
      await ctx.reply(
        `⚙️ 欢迎回来, ${tgUser.first_name}!\n\n` +
        `你的机甲宠物已经在等你了。\n` +
        `发送 /pet 查看状态，/earn 查看收益。`
      );
      return;
    }

    if (resp.ok && data.pet) {
      const sess = ctx.session as SessionData;
      sess.petId = data.pet.id;
      sess.step = "naming";

      await ctx.reply(
        `⚙️ 欢迎, ${tgUser.first_name}!\n\n` +
        `🤖 你获得了一只 ${data.pet.species} 机甲宠物！\n` +
        `品质: ${data.pet.quality} | 资质: HP${data.pet.apt_hp} ATK${data.pet.apt_atk} DEF${data.pet.apt_def} SPD${data.pet.apt_spd} INT${data.pet.apt_int}\n` +
        `赠送: ${data.bonus}\n\n` +
        `给它取个名字吧 (直接发送名字):`
      );
    } else {
      await ctx.reply("领取宠物失败: " + (data.error || "unknown error"));
    }
  } catch (err) {
    await ctx.reply("网络错误，请稍后重试。");
  }
});

bot.command("pet", async (ctx: BotContext) => {
  const token = await authenticate(ctx);
  if (!token) return;

  try {
    const resp = await fetch(`${BACKEND_URL}/api/v1/user/me/pets`, {
      headers: { "Authorization": `Bearer ${token}` },
    });
    const data = await resp.json() as { pets: any[] };

    if (!data.pets?.length) {
      await ctx.reply("你还没有宠物。发送 /start 领取你的初始机甲！");
      return;
    }

    const pet = data.pets[0];
    await ctx.reply(
      `🤖 ${pet.name || pet.species}\n` +
      `━━━━━━━━━━━━━━\n` +
      `品质: ${pet.quality} | Lv.${pet.level}\n` +
      `状态: ${pet.status}\n` +
      `能量: ${pet.energy}/100 | 心情: ${pet.mood}/100\n` +
      `TAI余额: ${pet.tai_balance || 0}\n` +
      `━━━━━━━━━━━━━━\n` +
      `发送 /earn 查看收益 | /market 交易市场`
    );
  } catch {
    await ctx.reply("获取宠物信息失败。");
  }
});

bot.command("market", async (ctx: Context) => {
  // TODO: Open Mini App via reply_markup with web_app
  await ctx.reply("📈 交易市场即将开放，敬请期待！");
});

bot.command("earn", async (ctx: BotContext) => {
  const token = await authenticate(ctx);
  if (!token) return;

  try {
    const resp = await fetch(`${BACKEND_URL}/api/v1/user/me/earnings`, {
      headers: { "Authorization": `Bearer ${token}` },
    });
    const data = await resp.json() as any;
    await ctx.reply(
      `💰 收益面板\n` +
      `━━━━━━━━━━━━━━\n` +
      `累计TAI: ${data.total_tai || 0}\n` +
      `累计USDT: ${data.total_usdt || 0}\n` +
      `━━━━━━━━━━━━━━\n` +
      `宠物正在自动执行赏金任务中...`
    );
  } catch {
    await ctx.reply("获取收益信息失败。");
  }
});

bot.command("breed", async (ctx: Context) => {
  await ctx.reply("🧬 繁殖系统开发中，即将上线！");
});

bot.command("bounty", async (ctx: BotContext) => {
  const token = await authenticate(ctx);
  if (!token) return;

  const sess = ctx.session as SessionData;
  const petId = sess.petId || "";

  try {
    const resp = await fetch(
      `${BACKEND_URL}/api/v1/bounty/available?pet_id=${petId}`,
      { headers: { "Authorization": `Bearer ${token}` } }
    );
    const data = await resp.json() as { bounties: any[] };

    if (!data.bounties?.length) {
      await ctx.reply(
        "🎯 当前没有可用的赏金任务。\n\n" +
        "发送 /task 发布一个任务，让宠物去执行！"
      );
      return;
    }

    const list = data.bounties.slice(0, 5).map((b: any, i: number) =>
      `${i + 1}. [${b.difficulty}] ${b.title}\n   奖励: ${b.reward_tai} TAI + ${b.reward_usdt} USDT`
    ).join("\n\n");

    await ctx.reply(`🎯 可用赏金任务:\n\n${list}\n\n宠物会自动接取匹配的任务执行。`);
  } catch {
    await ctx.reply("获取赏金列表失败。");
  }
});

bot.command("task", async (ctx: BotContext) => {
  const token = await authenticate(ctx);
  if (!token) return;

  // Simple inline task creation via command args
  // Format: /task <difficulty> <title>
  const args = ctx.message?.text?.split(" ").slice(1) || [];
  if (args.length < 2) {
    await ctx.reply(
      "📋 发布赏金任务:\n\n" +
      "格式: /task <难度> <任务描述>\n" +
      "难度: D(简单) C(普通) B(中等) A(困难) S(极难)\n\n" +
      "示例: /task B 帮我分析TON链上最近7天的DEX交易量趋势\n\n" +
      "奖励会根据难度自动计算。"
    );
    return;
  }

  const difficulty = args[0].toUpperCase();
  const title = args.slice(1).join(" ");

  if (!["D", "C", "B", "A", "S"].includes(difficulty)) {
    await ctx.reply("难度必须是 D/C/B/A/S 之一。");
    return;
  }

  // Reward auto-calculated by difficulty
  const rewards: Record<string, [number, number]> = {
    D: [5, 0.01], C: [15, 0.03], B: [50, 0.1], A: [150, 0.3], S: [500, 1.0],
  };
  const [rewardTAI, rewardUSDT] = rewards[difficulty];

  try {
    const resp = await fetch(`${BACKEND_URL}/api/v1/bounty/create`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`,
      },
      body: JSON.stringify({
        title,
        description: title, // use title as description for simple tasks
        difficulty,
        reward_tai: rewardTAI,
        reward_usdt: rewardUSDT,
      }),
    });

    if (resp.ok) {
      await ctx.reply(
        `✅ 赏金任务已发布！\n\n` +
        `📋 ${title}\n` +
        `难度: ${difficulty} | 奖励: ${rewardTAI} TAI + ${rewardUSDT} USDT\n\n` +
        `你的机甲宠物正在赶来接单...`
      );
    } else {
      const err = await resp.json() as any;
      await ctx.reply(`发布失败: ${err.error || "unknown"}`);
    }
  } catch {
    await ctx.reply("网络错误，请稍后重试。");
  }
});

// ─── Text handler (naming + chat) ─────────────────────────────────

bot.on("message:text", async (ctx: BotContext) => {
  const sess = ctx.session as SessionData;
  const text = ctx.message?.text?.trim();
  if (!text) return;

  if (sess.step === "naming" && sess.petId) {
    const token = await authenticate(ctx);
    if (token) {
      try {
        await fetch(`${BACKEND_URL}/api/v1/pet/${sess.petId}/name`, {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`,
          },
          body: JSON.stringify({ name: text }),
        });
      } catch { /* non-critical */ }
    }
    sess.step = "idle";
    await ctx.reply(
      `✅ 命名成功！你的机甲现在叫 "${text}"。\n\n` +
      `它已经开始自动搜索赏金任务了。\n` +
      `发送 /pet 随时查看状态。`
    );
    return;
  }

  // Default: personality response
  const responses = [
    `🤖 "${text}"? 收到, 老板。我去干活了。`,
    `🤖 了解。正在分析任务可行性...`,
    `🤖 老板放心，TAI 管够我就往死里干。`,
    `🤖 机甲已就绪，等待指令。`,
  ];
  await ctx.reply(responses[Math.floor(Math.random() * responses.length)]);
});

// ─── Start ─────────────────────────────────────────────────────────

bot.start({
  onStart: (botInfo) => {
    console.log(`Bot @${botInfo.username} running | backend: ${BACKEND_URL}`);
  },
});
