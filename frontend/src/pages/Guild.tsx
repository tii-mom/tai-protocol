/**
 * 公会 — 列表 + 创建 + 排行
 * TODO(Kimi/Gemini): 公会卡片、成员列表、贡献排行
 */
export default function GuildPage() {
  return (
    <div className="space-y-4 animate-slide-up">
      <h2 className="font-display text-lg text-mecha-neon">GUILDS</h2>

      <button className="w-full py-2 rounded-xl bg-mecha-steel border border-mecha-neon/20 text-sm text-mecha-neon">
        + 创建公会
      </button>

      <div className="space-y-3">
        <div className="rounded-xl bg-mecha-steel p-4 border border-white/5">
          <p className="text-xs text-gray-500 text-center py-8">暂无公会</p>
        </div>
        {/* TODO: GuildCard — 名称、人数、总战力、加入按钮 */}
      </div>
    </div>
  )
}
