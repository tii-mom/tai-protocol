/** 全局类型定义 — 与后端 Ent schema 对齐 */

export type PetQuality = 'common' | 'rare' | 'epic' | 'legendary' | 'mythic'
export type PetStatus = 'idle' | 'working' | 'breeding' | 'trading' | 'resting'
export type OrderSide = 'buy' | 'sell'
export type ItemType = 'pet' | 'skill'

export interface Pet {
  id: string
  owner_id: string
  species: string
  name: string
  quality: PetQuality
  generation: number
  growth_rate: number
  aptitude_hp: number
  aptitude_atk: number
  aptitude_def: number
  aptitude_spd: number
  aptitude_int: number
  skill_slots: number
  personality: string
  level: number
  exp: number
  mood: number
  energy: number
  status: PetStatus
  image_url: string
  on_chain: boolean
  nft_address: string | null
  created_at: string
}

export interface Skill {
  id: string
  name: string
  type: string
  rarity: PetQuality
  description: string
  effect_value: number
  durability_max: number
  price_tai: number
  image_url: string
}

export interface Order {
  id: string
  seller_id: string
  item_type: ItemType
  item_id: string
  price_tai: number
  status: 'open' | 'filled' | 'cancelled'
  created_at: string
}

export interface Bounty {
  id: string
  title: string
  description: string
  reward_tai: number
  reward_usdt: number
  status: 'open' | 'accepted' | 'submitted' | 'completed' | 'expired'
  publisher_id: string
  acceptor_id: string | null
  deadline: string
  created_at: string
}

export interface Guild {
  id: string
  name: string
  leader_id: string
  member_count: number
  total_power: number
  created_at: string
}

export interface PricePoint {
  timestamp: string
  price: number
  volume: number
}
