import { Routes, Route } from 'react-router-dom'
import Layout from '@/components/Layout'
import HomePage from '@/pages/Home'
import MarketPage from '@/pages/Market'
import PetDetailPage from '@/pages/PetDetail'
import BreedPage from '@/pages/Breed'
import BountyPage from '@/pages/Bounty'
import GuildPage from '@/pages/Guild'
import ProfilePage from '@/pages/Profile'

export default function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/market" element={<MarketPage />} />
        <Route path="/pet/:id" element={<PetDetailPage />} />
        <Route path="/breed" element={<BreedPage />} />
        <Route path="/bounty" element={<BountyPage />} />
        <Route path="/guild" element={<GuildPage />} />
        <Route path="/profile" element={<ProfilePage />} />
      </Routes>
    </Layout>
  )
}
