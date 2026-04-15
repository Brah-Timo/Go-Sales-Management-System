#!/bin/bash
export PATH=$PATH:/usr/local/go/bin
export CGO_ENABLED=1
cd "$(dirname "$0")/.."
go build -ldflags="-s -w" -o gestion-commerciale ./cmd/app/
echo "Build terminé: gestion-commerciale"
