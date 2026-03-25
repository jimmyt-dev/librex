#!/bin/sh
set -e

echo "[reliquary] Applying database schema..."
pnpm db:push

echo "[reliquary] Starting API..."
./api &

echo "[reliquary] Starting server..."
exec node build
