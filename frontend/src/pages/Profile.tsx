/**
 * 个人中心 — 资产、宠物列表、收益记录、设置
 * TODO(Kimi/Gemini): 资产饼图、收益折线图、邀请链接分享
 */
export default function ProfilePage() {
  return (
    <div className="space-y-4 animate-slide-up">
      {/* 用户信息 */}
      <div className="flex items-center gap-3">
        <div className="w-12 h-12 rounded-full bg-mecha-steel border border-mecha-neon/30 flex items-center justify-center">
          <span className="text-xl">🤖</span>
        </div>
        <div>
          <p className="font-medium text-sm">未连接</p>
          <p className="text-xs text-gray-500">连接 Telegram 钱包开始</p>
        </div>
      </div>

      {/* 资产 */}
      <section className="rounded-xl bg-mecha-steel p-4 border border-white/5">
        <h3 className="font-display text-sm text-gray-400 mb-3">ASSETS</h3>
        <div className="grid grid-cols-2 gap-4 text-center">
          <div>
            <p className="text-xl font-bold text-mecha-neon">0</p>
            <p className="text-xs text-gray-500">TAI</p>
          </div>
          <div>
            <p className="text-xl font-bold text-mecha-gold">0</p>
            <p className="text-xs text-gray-500">USDT</p>
          </div>
        </div>
      </section>

      {/* 功能入口 */}
      <section className="space-y-2">
        {['我的宠物', '交易记录', '收益明细', '邀请好友', '设置'].map((item) => (
          <button
            key={item}
            className="w-full py-3 px-4 rounded-xl bg-mecha-steel border border-white/5 text-left text-sm text-gray-300 flex justify-between items-center"
          >
            {item}
            <span className="text-gray-600">›</span>
          </button>
        ))}
      </section>
    </div>
  )
}
