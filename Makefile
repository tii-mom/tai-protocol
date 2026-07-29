.PHONY: help dev-bot dev-backend build test clean

help: ## 显示帮助
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# === Bot ===
dev-bot: ## 启动 TG Bot (开发模式)
	cd bot && npx tsx watch src/index.ts

install-bot: ## 安装 Bot 依赖
	cd bot && pnpm install

# === Backend ===
dev-backend: ## 启动后端 API (开发模式)
	cd backend && go run cmd/server/main.go

generate: ## 生成 Ent ORM 代码 (Docker, 无需本地Go)
	docker run --rm -v $(PWD)/backend:/app -w /app golang:1.22-alpine sh -c "go mod download && go generate ./ent"

generate-local: ## 生成 Ent ORM 代码 (需要本地Go 1.22+)
	cd backend && go generate ./ent

test-backend: ## 运行后端测试
	cd backend && go test ./...

# === Frontend ===
dev-frontend: ## 启动前端 (开发模式)
	cd frontend && pnpm dev

# === Market Maker ===
dev-mm: ## 启动做市机器人
	cd market-maker && npx tsx watch src/index.ts

# === Contracts ===
build-contracts: ## 编译 Tolk 合约 (需要 TON SDK)
	@echo "TODO: tolk compile contracts/*.tolk"

# === Desktop ===
dev-desktop: ## 启动桌面端 (开发模式)
	cd desktop && pnpm tauri dev

# === Utils ===
docker-up: ## Docker 全家桶启动
	docker compose up -d

docker-down: ## Docker 全家桶停止
	docker compose down

clean: ## 清理构建产物
	rm -rf bot/dist frontend/dist desktop/src-tauri/target market-maker/dist
