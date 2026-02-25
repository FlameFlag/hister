#!/usr/bin/env bash
# shellcheck shell=bash
set -e

cd "$(dirname -- "$0")/.." || exit

npm install --workspaces
npm run build -w @hister/app

rm -rf server/static/app
mkdir -p server/static
cp -r webui/app/build server/static/app
