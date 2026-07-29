# TAI Protocol — 工作交接文档

> 最后更新: 2026-07-29
> 当前进度: ~45%

## 已完成

- 全部文档（白皮书、经济模型、操盘方案、架构、DB设计、合约规范、3api集成方案）
- 6个模块骨架（backend/bot/market-maker/frontend/desktop/contracts）
- TG initData 签名验证 + JWT 认证（完整实现）
- UserService / PetService 真实 Ent DB 查询
- /pet/execute 代理端点（TAI扣费 → 3api平台key转发）
- 3api Phase 0 对接（平台池账户模式，tai-pets组6折）
- Bounty 系统框架（service + handler + bot命令）
- Bot 完整命令流（/start → 领宠物 → /task发赏金 → /bounty查看）
- Deploy 全套（docker-compose, Dockerfile, CI, .env.example）

## 未完成（按优先级）

### P0 — 阻塞项（必须先做）

1. **go generate ./ent**
   - 本机无 Go 环境，执行 `make generate`（需 Docker）或 `make generate-local`（需 Go 1.22+）
   - 生成后 backend 才能编译

2. **Ent bounty schema**
   - 新建 `backend/ent/schema/bounty.go`（字段参考 service/bounty_service.go 的 Bounty struct）
   - 新建 `backend/ent/schema/usage_log.go`（记录每次API调用）
   - 然后重新 `make generate`

3. **本地联调**
   - `docker compose up postgres redis` → `make generate` → `make dev-backend` → `make dev-bot`
   - 需要真实 TG_BOT_TOKEN（从 @BotFather 获取）
   - 需要 3api 平台 key（在 3api 后台创建 tai-pets 组 + key）

### P1 — 核心功能补齐

4. BountyService DB 查询实现（等 bounty schema generate 后）
5. handler.go CreateBounty 接入 BountyService（去掉硬编码）
6. Bot buildBotInitData 改为真正的 TG WebApp 签名验证
7. 3api router.go 注册 `RegisterTAIInternalRoutes`（一行代码）
8. PetExecute 中的 TODO 步骤（余额检查、扣费、记录）— 逻辑已写好，等 Ent 生成后对接

### P2 — 做市 + 经济循环

9. Market Maker 策略细化（挂单算法、persona行为、价格曲线、熔断触发）
10. 赏金任务自动匹配 + Agent Loop 真实运行
11. 宠物经验/升级系统
12. 邀请裂变 + 排行榜

### P3 — 交给其他模型

13. **Frontend**（Kimi K3 / Gemini 3.6）：frontend/src/pages/ 里的 TODO
14. **美术**（GPT Image 2）：机甲立绘、技能图标、UI素材
15. **合约逻辑**（Qwen 后续）：繁殖基因算法、Jetton mint/burn、testnet部署

## 关键文件索引

| 文件 | 说明 |
|------|------|
| backend/internal/handler/handler.go | 所有API handler（核心） |
| backend/internal/service/*.go | 业务逻辑层 |
| backend/internal/threeapi/client.go | 3api代理客户端 |
| backend/internal/auth/telegram.go | TG签名验证 |
| backend/ent/schema/*.go | 数据库schema |
| bot/src/index.ts | Bot主入口+命令 |
| bot/src/agent/executor.ts | AI Agent执行器 |
| bot/src/services/bounty.ts | 赏金客户端 |
| market-maker/src/index.ts | 做市机器人 |
| docs/07_3api集成方案.md | 3api对接设计 |

## 环境要求

- Go 1.22+（或 Docker 跑 generate）
- Node.js 20+
- PostgreSQL 16 + Redis 7（docker compose 提供）
- TG Bot Token
- 3api 平台 API Key

## 注意事项

- 3api 项目改动未 push（本地 commit），需 review 后手动推
- Ent schema 修改后必须重新 generate
- 前端用 pnpm，bot/market-maker 用 npm
- 合约部署顺序：tai_token → pet_nft → breeding → storage
