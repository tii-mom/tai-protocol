import { ReactNode } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import clsx from 'clsx'

const NAV_ITEMS = [
  { path: '/', label: '首页', icon: '⚡' },
  { path: '/market', label: '市场', icon: '📊' },
  { path: '/bounty', label: '赏金', icon: '🎯' },
  { path: '/breed', label: '繁殖', icon: '🧬' },
  { path: '/profile', label: '我的', icon: '🤖' },
]

export default function Layout({ children }: { children: ReactNode }) {
  const navigate = useNavigate()
  const { pathname } = useLocation()

  return (
    <div className="min-h-screen flex flex-col max-w-md mx-auto bg-mecha-carbon">
      {/* Header */}
      <header className="tg-safe-top sticky top-0 z-50 bg-mecha-steel/90 backdrop-blur border-b border-white/5 px-4 py-3">
        <div className="flex items-center justify-between">
          <h1 className="font-display text-lg text-mecha-neon tracking-wider">TAI</h1>
          <div className="text-xs text-gray-400">
            {/* TODO: 显示 TAI 余额，从 useUserStore 读取 */}
            0 TAI
          </div>
        </div>
      </header>

      {/* Main content */}
      <main className="flex-1 overflow-y-auto px-4 py-4">
        {children}
      </main>

      {/* Bottom navigation */}
      <nav className="tg-safe-bottom sticky bottom-0 bg-mecha-steel/95 backdrop-blur border-t border-white/5">
        <div className="flex justify-around py-2">
          {NAV_ITEMS.map((item) => (
            <button
              key={item.path}
              onClick={() => navigate(item.path)}
              className={clsx(
                'flex flex-col items-center gap-0.5 px-3 py-1 rounded-lg transition-colors',
                pathname === item.path
                  ? 'text-mecha-neon'
                  : 'text-gray-500 hover:text-gray-300',
              )}
            >
              <span className="text-lg">{item.icon}</span>
              <span className="text-[10px]">{item.label}</span>
            </button>
          ))}
        </div>
      </nav>
    </div>
  )
}
