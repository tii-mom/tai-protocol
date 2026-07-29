#!/bin/bash
# TAI Protocol - Contract Development Environment Setup
# Prerequisites: Node.js 20+, TON SDK

set -e

echo "🔧 Setting up TON contract development environment..."

# 1. Install TON CLI tools
echo "→ Installing @ton/ton SDK..."
cd "$(dirname "$0")/.."
npm init -y --silent 2>/dev/null || true
npm install --save-dev @ton/ton @ton/crypto @ton/core typescript ts-node --silent

# 2. Create build output directory
mkdir -p contracts/build

# 3. Verify Tolk compiler availability
if command -v tolk &> /dev/null; then
    echo "✅ Tolk compiler found: $(tolk --version 2>/dev/null || echo 'installed')"
else
    echo "⚠️  Tolk compiler not found."
    echo "   Install via: https://docs.ton.org/develop/smart-contracts/tolk-overview"
    echo "   Or use TON Blueprint: npx @ton/blueprint"
fi

# 4. Install TON Blueprint (recommended dev tool)
echo "→ Installing TON Blueprint..."
npm install --save-dev @ton/blueprint --silent 2>/dev/null || echo "   (Blueprint optional, install manually if needed)"

echo ""
echo "✅ Contract environment ready."
echo ""
echo "Usage:"
echo "  Build:    ./scripts/build-contracts.sh"
echo "  Test:     ./scripts/test-contracts.sh"
echo "  Deploy:   ./scripts/deploy-contracts.sh [testnet|mainnet]"
echo ""
