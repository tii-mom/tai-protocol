import { Bot, Context, session } from "grammy";
import dotenv from "dotenv";

dotenv.config();

const bot = new Bot(process.env.TG_BOT_TOKEN || "");

// Session middleware
bot.use(session({ initial: () => ({ step: "idle" }) }));

// === Commands ===

bot.command("start", async (ctx: Context) => {
  const tgUser = ctx.from;
  if (!tgUser) return;

  // TODO: Check if user exists in DB, create if not
  // TODO: Issue starter pet (N quality, 1 skill slot)

  await ctx.reply(
    `⚙️ 欢迎, ${tgUser.first_name}!\n\n` +
    `我是 TAI Protocol 的机甲宠物终端。\n` +
    `你已获得一只初始机甲宠物。\n\n` +
    `给它取个名字吧 (直接发送名字):`
  );
});

bot.command("pet", async (ctx: Context) => {
  // TODO: Show pet status card
  await ctx.reply("🤖 [宠物状态卡片 - TODO]");
});

bot.command("market", async (ctx: Context) => {
  // TODO: Open Mini App (marketplace)
  await ctx.reply("📈 [交易市场 - TODO: 打开 Mini App]");
});

bot.command("earn", async (ctx: Context) => {
  // TODO: Show earnings dashboard
  await ctx.reply("💰 [收益面板 - TODO]");
});

bot.command("breed", async (ctx: Context) => {
  // TODO: Initiate breeding flow
  await ctx.reply("🧬 [繁殖系统 - TODO]");
});

// === Text handler (for naming pet) ===

bot.on("message:text", async (ctx: Context) => {
  const text = ctx.message?.text;
  if (!text) return;

  // TODO: If in "naming" step, save pet name
  // TODO: Otherwise, treat as chat with pet (personality responses)

  await ctx.reply(`🤖 "${text}"? 收到, 老板。我去干活了。`);
});

// === Start ===

bot.start({
  onStart: (botInfo) => {
    console.log(`✅ Bot @${botInfo.username} is running`);
  },
});
