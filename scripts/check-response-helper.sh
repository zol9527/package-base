#!/usr/bin/env bash
set -euo pipefail

# 扫描 handlers 是否绕过统一响应 helper（utils.Success/Error）直接返回 JSON
# 规则：禁止在 internal/handlers 中直接使用 c.JSON / ctx.JSON / fiber.Map 直接返回

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
TARGET_DIR="$ROOT_DIR/internal/handlers"

violations=0

if grep -R "\.JSON\(" -n "$TARGET_DIR" >/dev/null 2>&1; then
  echo "[response-check][FAIL] 发现直接使用 JSON 返回："
  grep -R "\.JSON\(" -n "$TARGET_DIR" || true
  violations=1
fi

if grep -R "fiber\.Map\{" -n "$TARGET_DIR" >/dev/null 2>&1; then
  echo "[response-check][FAIL] 发现直接使用 fiber.Map 返回："
  grep -R "fiber\.Map\{" -n "$TARGET_DIR" || true
  violations=1
fi

if [ $violations -ne 0 ]; then
  echo "[response-check] ❌ 未通过"
  exit 1
fi

echo "[response-check] ✅ 通过"
