/**
 * TAI Protocol API Client
 * 所有后端接口调用集中在此，Kimi/Gemini 补齐页面时直接调用
 */
import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10_000,
})

// Telegram WebApp auth token 注入
api.interceptors.request.use((config) => {
  const tg = (window as any).Telegram?.WebApp
  if (tg?.initData) {
    config.headers['X-TG-Init-Data'] = tg.initData
  }
  return config
})

// ─── User ───────────────────────────────────────────
export const userApi = {
  auth: () => api.post('/user/auth'),
  profile: () => api.get('/user/profile'),
  balance: () => api.get('/user/balance'),
}

// ─── Pet ────────────────────────────────────────────
export const petApi = {
  claim: () => api.post('/pet/claim'),
  list: (params?: { page?: number; sort?: string }) => api.get('/pet/list', { params }),
  detail: (id: string) => api.get(`/pet/${id}`),
  myPets: () => api.get('/pet/mine'),
}

// ─── Market ─────────────────────────────────────────
export const marketApi = {
  listings: (params?: { type?: string; sort?: string; page?: number }) =>
    api.get('/market/listings', { params }),
  orderBook: (pair: string) => api.get(`/market/book/${pair}`),
  createOrder: (data: { side: 'buy' | 'sell'; item_type: string; item_id: string; price: number }) =>
    api.post('/market/order', data),
  cancelOrder: (id: string) => api.delete(`/market/order/${id}`),
  priceHistory: (pair: string) => api.get(`/market/price/${pair}`),
}

// ─── Skill ──────────────────────────────────────────
export const skillApi = {
  list: () => api.get('/skill/list'),
  buy: (skillId: string) => api.post('/skill/buy', { skill_id: skillId }),
  use: (petId: string, skillId: string) => api.post('/skill/use', { pet_id: petId, skill_id: skillId }),
}

// ─── Bounty ─────────────────────────────────────────
export const bountyApi = {
  list: (params?: { status?: string }) => api.get('/bounty/list', { params }),
  accept: (id: string) => api.post(`/bounty/${id}/accept`),
  submit: (id: string, data: { result: string }) => api.post(`/bounty/${id}/submit`, data),
}

// ─── Breed ──────────────────────────────────────────
export const breedApi = {
  request: (data: { pet_a_id: string; pet_b_id: string }) => api.post('/breed/request', data),
  history: () => api.get('/breed/history'),
}

// ─── Guild ──────────────────────────────────────────
export const guildApi = {
  list: () => api.get('/guild/list'),
  create: (data: { name: string }) => api.post('/guild/create', data),
  join: (id: string) => api.post(`/guild/${id}/join`),
  detail: (id: string) => api.get(`/guild/${id}`),
}

export default api
