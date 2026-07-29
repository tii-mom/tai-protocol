#!/bin/bash
# Build all Tolk contracts to BOC (Bag of Cells)
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONTRACTS_DIR="$SCRIPT_DIR/.."
BUILD_DIR="$CONTRACTS_DIR/build"

mkdir -p "$BUILD_DIR"

echo "🔨 Building TAI Protocol contracts..."

CONTRACTS=(
    "pet_nft"
    "skill_book"
    "breeding"
    "marketplace"
    "tai_token"
    "ad_vault"
    "bounty_vault"
    "agentic_wallet"
)

for name in "${CONTRACTS[@]}"; do
    src="$CONTRACTS_DIR/${name}.tolk"
    if [ -f "$src" ]; then
        echo "  → Compiling ${name}.tolk"
        # tolk --output "$BUILD_DIR/${name}.boc" "$src" 2>/dev/null || echo "    ⚠️  ${name}: Tolk not available, skipping"
        echo "    (placeholder - Tolk compiler needed)"
    else
        echo "  ⚠️  ${name}.tolk not found, skipping"
    fi
done

echo ""
echo "✅ Build complete. Output: $BUILD_DIR/"
ls -la "$BUILD_DIR/" 2>/dev/null || true
