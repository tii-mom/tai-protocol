/**
 * 首页 — 宠物总览 + 快捷入口
 * TODO(Kimi/Gemini): 补齐宠物卡片动画、收益实时刷新、FOMO排行榜
 */
export default function HomePage() {
  return (
    <div className="space-y-6 animate-slide-up">
      {/* 我的机甲宠物 */}
      <section>
        <h2 className="font-display text-sm text-gray-400 mb-3">MY MECHA</h2>
        <div className="grid grid-cols-2 gap-3">
          {/* TODO: PetCard 组件 — 机甲立绘 + 等级 + 状态指示灯 */}
          <div className="aspect-square rounded-xl bg-mecha-steel border border-white/5 flex items-center justify-center">
            <span className="text-4xl opacity-30">🤖</span>
          </div>
        </div>
      </section>

      {/* 今日收益 */}
      <section className="rounded-xl bg-mecha-steel p-4 border border-white/5">
        <h2 className="font-display text-sm text-gray-400 mb-2">TODAY EARNINGS</h2>
        <p className="text-2xl font-bold text-mecha-gold">+0.00 TAI</p>
        <p className="text-xs text-gray-500 mt-1">宠物自动执行赏金任务中...</p>
      </section>

      {/* 市场快讯 */}
      <section className="rounded-xl bg-mecha-steel p-4 border border-white/5">
        <h2 className="font-display text-sm text-gray-400 mb-2">MARKET</h2>
        <div className="flex justify-between text-sm">
          <span className="text-gray-300">PET/TAI</span>
          <span className="text-mecha-neon">--</span>
        </div>
      </section>
    </div>
  )
}
