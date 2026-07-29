#!/bin/bash
# Deploy TAI Protocol contracts to TON network
# Usage: ./deploy-contracts.sh [testnet|mainnet]
set -e

NETWORK="${1:-testnet}"

echo "🚀 Deploying TAI Protocol contracts to $NETWORK..."
echo ""

if [ "$NETWORK" = "mainnet" ]; then
    echo "⚠️  MAINNET DEPLOYMENT"
    read -p "   Are you sure? (yes/no): " confirm
    if [ "$confirm" != "yes" ]; then
        echo "   Aborted."
        exit 0
    fi
fi

# TODO: Deploy order matters (dependencies):
# 1. tai_token.tolk (Jetton Master) - no deps
# 2. pet_nft.tolk (NFT Collection) - no deps
# 3. skill_book.tolk - no deps
# 4. breeding.tolk - depends on pet_nft collection address
# 5. marketplace.tolk - depends on pet_nft + tai_token
# 6. bounty_vault.tolk - depends on tai_token
# 7. ad_vault.tolk - depends on tai_token
# 8. agentic_wallet.tolk - standalone (per-pet instances)

echo "Deployment order:"
echo "  1. TAI Token (Jetton Master)"
echo "  2. Pet NFT Collection"
echo "  3. Skill Book"
echo "  4. Breeding Contract"
echo "  5. Marketplace"
echo "  6. Bounty Vault"
echo "  7. Ad Vault"
echo "  8. Agentic Wallet (template)"
echo ""
echo "TODO: Implement deployment via @ton/blueprint or ton-cli"
echo "      Update .env with deployed addresses after each step."
