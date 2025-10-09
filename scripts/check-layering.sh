#!/usr/bin/env bash
set -euo pipefail

# 简单分层静态检查：
# 规则 1: handlers 层禁止直接 import 存储实现 (internal/storages)
# 规则 2: services 层禁止直接 import 存储实现 (internal/storages)

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
FAIL=0

echo "[layer-check] 执行分层规则校验..."

# 规则 1
if grep -R "internal/storages" -n "$ROOT_DIR/internal/handlers" >/dev/null 2>&1; then
  echo "[layer-check][FAIL] handlers 层发现对 storages 实现的直接引用"
  grep -R "internal/storages" -n "$ROOT_DIR/internal/handlers" || true
  FAIL=1
else
  echo "[layer-check][OK] handlers 未直接引用 storages 实现"
fi

# 规则 2
if grep -R "internal/storages" -n "$ROOT_DIR/internal/services" >/dev/null 2>&1; then
  echo "[layer-check][FAIL] services 层发现对 storages 实现的直接引用"
  grep -R "internal/storages" -n "$ROOT_DIR/internal/services" || true
  FAIL=1
else
  echo "[layer-check][OK] services 未直接引用 storages 实现"
fi

if [ $FAIL -ne 0 ]; then
  echo "[layer-check] ❌ 发现违规，请修正后再提交"
  exit 1
fi

echo "[layer-check] ✅ 通过"
