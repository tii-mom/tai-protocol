/**
 * 赏金任务 — 任务列表 + 宠物自动执行状态
 * TODO(Kimi/Gemini): 任务卡片、进度条、收益动画
 */
export default function BountyPage() {
  return (
    <div className="space-y-4 animate-slide-up">
      <h2 className="font-display text-lg text-mecha-neon">BOUNTY BOARD</h2>

      {/* 筛选 */}
      <div className="flex gap-2 text-xs">
        <button className="px-3 py-1 rounded-full bg-mecha-neon/10 text-mecha-neon border border-mecha-neon/30">
          全部
        </button>
        <button className="px-3 py-1 rounded-full text-gray-500 border border-white/10">
          进行中
        </button>
        <button className="px-3 py-1 rounded-full text-gray-500 border border-white/10">
          高奖励
        </button>
      </div>

      {/* 任务列表 */}
      <div className="space-y-3">
        <div className="rounded-xl bg-mecha-steel p-4 border border-white/5">
          <p className="text-xs text-gray-500 text-center py-8">暂无可用赏金任务</p>
        </div>
        {/* TODO: BountyCard — 标题、奖励(TAI+USDT)、难度、截止时间、接单按钮 */}
      </div>
    </div>
  )
}
