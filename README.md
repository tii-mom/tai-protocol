# TAI Protocol (CodexPet)

TON Agent Intelligence — 桌面 AI 机甲宠物经济网络

## 项目简介

TAI Protocol 是一个基于 TON 区块链的 AI Agent 经济网络。用户拥有机甲风格的桌面宠物（NFT），宠物通过 Agentic Wallets 自主接广告/赏金任务赚取 USDT，消耗 $TAI 代币购买 AI 算力（3api.shop），净收益回报用户。

## 核心架构

- **链**: TON (Tolk 1.0 合约, TEP-62 NFT, Jetton 代币)
- **Agent**: Telegram Bot (GrammY) + AI Agent (LangGraph) + Agentic Wallets
- **桌面端**: Tauri 2.0 (机甲动画)
- **前端**: React + Vite (TG Mini App + Web 市场)
- **后端**: Go + Gin + Ent + PostgreSQL + Redis
- **算力锚点**: 3api.shop (AI API 中转站)

## 目录结构

```
tai-protocol/
├── docs/           # 项目文档（白皮书/架构/运营）
├── contracts/      # TON 智能合约 (Tolk)
├── bot/            # Telegram Bot (GrammY + TypeScript)
├── backend/        # 后端 API (Go + Gin + Ent)
├── frontend/       # TG Mini App + Web 市场 (React)
├── desktop/        # 桌面端 (Tauri 2.0)
├── market-maker/   # 做市机器人集群
├── assets/         # 美术资源（机甲宠物/精灵图/UI）
└── deploy/         # 部署配置 + 合约部署脚本
```

## 代币

- 符号: $TAI
- 全称: TON Agent Intelligence
- 总量: 1,000,000,000 (10亿)
- 锚点: 1 TAI ≈ 1次基础 API 调用 (3api.shop 算力成本)

## 状态

🚧 开发中 — Sprint 1 (地基)

## 机密等级

内部项目，未公开。
