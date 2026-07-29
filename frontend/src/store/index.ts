/**
 * 全局状态管理 (Zustand)
 * 轻量级，适合 TG Mini App 场景
 */
import { create } from 'zustand'

interface UserState {
  userId: string | null
  walletAddress: string | null
  taiBalance: number
  usdtBalance: number
  setUser: (data: Partial<UserState>) => void
}

export const useUserStore = create<UserState>((set) => ({
  userId: null,
  walletAddress: null,
  taiBalance: 0,
  usdtBalance: 0,
  setUser: (data) => set(data),
}))

interface MarketState {
  currentPair: string
  lastPrice: number
  priceChange24h: number
  setMarket: (data: Partial<MarketState>) => void
}

export const useMarketStore = create<MarketState>((set) => ({
  currentPair: 'PET/TAI',
  lastPrice: 0,
  priceChange24h: 0,
  setMarket: (data) => set(data),
}))
