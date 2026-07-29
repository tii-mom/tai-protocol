# TAI Protocol — 开发指南

## 环境要求

| 工具 | 版本 | 用途 |
|------|------|------|
| Go | 1.22+ | Backend |
| Node.js | 20+ | Bot / Market Maker / Frontend |
| pnpm | 9+ | Frontend 包管理 |
| Rust | 1.77+ | Desktop (Tauri) |
| Docker + Compose | latest | 本地一键启动 |
| TON Blueprint | latest | 合约编译部署 |

## 快速启动（Docker 全家桶）

```bash
cp .env.example .env
# 编辑 .env 填入 TG_BOT_TOKEN 等必要值
docker compose up -d
```

服务地址：Backend :8080 / Frontend :3000 / PostgreSQL :5432 / Redis :6379

## 各模块独立开发

### Backend (Go)

```bash
cd backend
go mod download
# 需要本地 PostgreSQL + Redis，或用 docker compose up postgres redis
go run ./cmd/server
```

配置读取优先级：环境变量 > .env > config.yaml > 默认值

### Bot (TypeScript / GrammY)

```bash
cd bot
npm install
npm run dev    # ts-node 热重载
```

必须配置 `TG_BOT_TOKEN`（从 @BotFather 获取）和 `BACKEND_URL`。

### Market Maker

```bash
cd market-maker
npm install
npm run dev
```

通过 `MM_ENABLED=false` 可关闭做市逻辑，仅观察市场状态。

### Frontend (React + Vite)

```bash
cd frontend
pnpm install
pnpm dev       # http://localhost:3000
```

API 请求自动代理到 localhost:8080。页面骨架已搭好，TODO 注释标记了需要补齐的交互。

### Desktop (Tauri 2.0)

```bash
cd desktop
pnpm install
pnpm tauri:dev   # 需要 Rust toolchain
```

双窗口架构：main（主面板 420×720）+ pet-overlay（透明桌面宠物 200×200）。

### Contracts (Tolk)

```bash
cd contracts
bash scripts/setup.sh    # 安装 TON SDK + Blueprint（首次）
bash scripts/build.sh    # 编译 .tolk → .boc
bash scripts/deploy.sh   # 部署到 testnet/mainnet
```

合约当前为接口桩，核心逻辑待补齐。部署顺序：tai_token → pet_nft → breeding → storage。

## 项目结构

```
tai-protocol/
├── backend/          # Go API 服务 (Gin + Ent + PostgreSQL)
├── bot/              # Telegram Bot (GrammY + TON SDK + AI Agent)
├── market-maker/     # 做市机器人集群
├── frontend/         # TG Mini App + Web 市场 (React + Vite)
├── desktop/          # 桌面宠物 (Tauri 2.0)
├── contracts/        # TON 智能合约 (Tolk)
├── docs/             # 白皮书、架构、经济模型等文档
├── docker-compose.yml
├── .env.example
├── Makefile
└── .github/workflows/ci.yml
```

## 开发分工

| 模块 | 负责 | 说明 |
|------|------|------|
| 架构/合约/总控 | Qwen (本体) | 骨架已完成，补齐合约逻辑 |
| Bot/前端/做市 | ChatGPT 5.6sol | 补齐 handler 实现、做市策略 |
| 前端视觉 | Kimi K3 / Gemini 3.6 | 页面 TODO 标记处 |
| 美术资源 | GPT Image 2 | 机甲立绘、UI 素材 |
| 文档分析 | Gemini 3.6 | TON 文档、竞品研究 |
| 文案 | Kimi K3 | 社群运营、公告 |

## 常用命令 (Makefile)

```bash
make dev-bot          # 启动 Bot
make dev-backend      # 启动 Backend
make dev-mm           # 启动 Market Maker
make dev-frontend     # 启动 Frontend
make build-contracts  # 编译合约
make test             # 全模块测试
make docker-up        # Docker 全家桶
make docker-down      # 停止
```

## 注意事项

- `.env` 文件绝不提交到 Git（已在 .gitignore 中）
- 合约部署前必须在 testnet 完整测试
- Market Maker 有 150% 日涨幅熔断，生产环境勿随意修改
- 前端使用 `@/` 路径别名，对应 `src/` 目录
- Ent schema 修改后需运行 `go generate ./ent` 重新生成代码
