import { Bot, Context, session } from "grammy";
import dotenv from "dotenv";

dotenv.config();

const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";
const bot = new Bot(process.env.TG_BOT_TOKEN || "");

// Session: stores JWT token after auth
interface SessionData {
  step: "idle" | "naming";
  token: string | null;
  petId: string | null;
}

bot.use(session({ initial: (): SessionData => ({ step: "idle", token: null, petId: null }) }));

// ─── Auth helper ───────────────────────────────────────────────────

async function authenticate(ctx: Context): Promise<string | null> {
  const sess = ctx.session as SessionData;
  if (sess.token) return sess.token;

  // Build initData string from the update (for WebApp this comes from Telegram.WebApp)
  // For bot commands, we use a simplified auth: send tg_user_id directly
  // In production, Mini App sends real initData; bot commands use internal auth
  try {
    const resp = await fetch(`${BACKEND_URL}/api/v1/user/auth/tg`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        // For bot-to-backend auth, we pass user info directly
        // The backend verifies via bot token (server-to-server trust)
        init_data: buildBotInitData(ctx),
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

bot.command("start", async (ctx: Context) => {
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

bot.command("pet", async (ctx: Context) => {
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

bot.command("earn", async (ctx: Context) => {
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

// ─── Text handler (naming + chat) ─────────────────────────────────

bot.on("message:text", async (ctx: Context) => {
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
