#!/bin/sh
set -e

echo "[Librex] Applying database schema..."
pnpm db:push

echo "[Librex] Starting API..."
./api &

echo "[Librex] Starting server..."
exec node build
