/**
 * 交易市场 — 挂单列表 + 下单
 * TODO(Kimi/Gemini): K线图(recharts)、深度图、实时价格跳动动画
 */
export default function MarketPage() {
  return (
    <div className="space-y-4 animate-slide-up">
      <h2 className="font-display text-lg text-mecha-neon">MARKET</h2>

      {/* 交易对切换 */}
      <div className="flex gap-2">
        <button className="px-3 py-1 rounded-full bg-mecha-neon/10 text-mecha-neon text-xs border border-mecha-neon/30">
          PET/TAI
        </button>
        <button className="px-3 py-1 rounded-full text-gray-500 text-xs border border-white/10">
          SKILL/TAI
        </button>
      </div>

      {/* 价格面板 */}
      <div className="rounded-xl bg-mecha-steel p-4 border border-white/5">
        <p className="text-3xl font-bold font-display">--</p>
        <p className="text-xs text-gray-500 mt-1">24h Vol: --</p>
      </div>

      {/* 挂单列表 */}
      <div className="space-y-2">
        <p className="text-xs text-gray-500">暂无挂单</p>
        {/* TODO: OrderRow 组件 — 买卖方向色(绿/红)、价格、数量 */}
      </div>
    </div>
  )
}
