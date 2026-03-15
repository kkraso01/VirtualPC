#!/usr/bin/env bash
set -euo pipefail
mkdir -p bin
go build -o bin/virtualpcd ./cmd/virtualpcd
go build -o bin/vpc ./cmd/vpc
go build -o bin/vpc-agent ./cmd/vpc-agent
