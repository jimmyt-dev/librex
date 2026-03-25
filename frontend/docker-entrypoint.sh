#!/bin/sh
set -e

echo "[reliquary] Applying database schema..."
pnpm db:push --force
echo "[reliquary] Starting server..."
exec node build
