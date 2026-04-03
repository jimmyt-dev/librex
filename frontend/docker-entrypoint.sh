#!/bin/sh
set -e

echo "[Librex] Applying database schema..."
pnpm db:push --force
echo "[Librex] Starting server..."
exec node build
