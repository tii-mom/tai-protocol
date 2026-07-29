/**
 * 宠物详情 — 机甲立绘 + 属性雷达图 + 技能槽 + 操作按钮
 * TODO(Kimi/Gemini): 雷达图(recharts)、技能装备拖拽、繁殖/上架按钮
 */
import { useParams } from 'react-router-dom'

export default function PetDetailPage() {
  const { id } = useParams<{ id: string }>()

  return (
    <div className="space-y-4 animate-slide-up">
      {/* 机甲立绘 */}
      <div className="aspect-square rounded-xl bg-mecha-steel border border-white/5 flex items-center justify-center relative overflow-hidden">
        <span className="text-6xl opacity-20">🤖</span>
        <div className="absolute bottom-2 left-2 text-xs text-mecha-neon font-display">
          GEN-0 #{id}
        </div>
      </div>

      {/* 属性面板 */}
      <section className="rounded-xl bg-mecha-steel p-4 border border-white/5">
        <h3 className="font-display text-sm text-gray-400 mb-3">ATTRIBUTES</h3>
        <div className="grid grid-cols-5 gap-2 text-center text-xs">
          {['HP', 'ATK', 'DEF', 'SPD', 'INT'].map((stat) => (
            <div key={stat}>
              <div className="w-8 h-8 mx-auto rounded-full bg-mecha-carbon flex items-center justify-center text-mecha-neon">
                --
              </div>
              <span className="text-gray-500 mt-1 block">{stat}</span>
            </div>
          ))}
        </div>
      </section>

      {/* 技能槽 */}
      <section className="rounded-xl bg-mecha-steel p-4 border border-white/5">
        <h3 className="font-display text-sm text-gray-400 mb-3">SKILL SLOTS</h3>
        <div className="flex gap-2">
          {[1, 2, 3, 4].map((slot) => (
            <div
              key={slot}
              className="w-12 h-12 rounded-lg bg-mecha-carbon border border-dashed border-white/10 flex items-center justify-center text-gray-600 text-xs"
            >
              {slot}
            </div>
          ))}
        </div>
      </section>

      {/* 操作 */}
      <div className="flex gap-3">
        <button className="flex-1 py-3 rounded-xl bg-mecha-neon/10 text-mecha-neon border border-mecha-neon/30 text-sm font-medium">
          上架交易
        </button>
        <button className="flex-1 py-3 rounded-xl bg-mecha-gold/10 text-mecha-gold border border-mecha-gold/30 text-sm font-medium">
          繁殖
        </button>
      </div>
    </div>
  )
}
