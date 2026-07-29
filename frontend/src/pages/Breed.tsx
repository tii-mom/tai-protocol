/**
 * 繁殖中心 — 选择两只宠物 + 费用 + 基因预览
 * TODO(Kimi/Gemini): 宠物选择器、基因遗传概率可视化、繁殖动画
 */
export default function BreedPage() {
  return (
    <div className="space-y-4 animate-slide-up">
      <h2 className="font-display text-lg text-mecha-neon">BREEDING LAB</h2>

      {/* 选择父母 */}
      <div className="flex items-center justify-center gap-4">
        <div className="w-24 h-24 rounded-xl bg-mecha-steel border border-dashed border-mecha-neon/30 flex items-center justify-center">
          <span className="text-xs text-gray-500">选择父体</span>
        </div>
        <span className="text-mecha-gold text-xl">×</span>
        <div className="w-24 h-24 rounded-xl bg-mecha-steel border border-dashed border-mecha-gold/30 flex items-center justify-center">
          <span className="text-xs text-gray-500">选择母体</span>
        </div>
      </div>

      {/* 费用信息 */}
      <div className="rounded-xl bg-mecha-steel p-4 border border-white/5 text-sm space-y-2">
        <div className="flex justify-between">
          <span className="text-gray-400">繁殖费用</span>
          <span className="text-mecha-neon">-- TAI</span>
        </div>
        <div className="flex justify-between">
          <span className="text-gray-400">冷却时间</span>
          <span className="text-gray-300">24h</span>
        </div>
        <div className="flex justify-between">
          <span className="text-gray-400">后代品质概率</span>
          <span className="text-gray-300">基于双亲资质</span>
        </div>
      </div>

      <button className="w-full py-3 rounded-xl bg-gradient-to-r from-mecha-neon/20 to-mecha-gold/20 border border-mecha-neon/30 text-sm font-medium text-white">
        开始繁殖
      </button>
    </div>
  )
}
