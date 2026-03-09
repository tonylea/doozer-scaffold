#!/bin/bash
set -e

# === Base tooling ===
npm install -g markdownlint-cli2 @commitlint/cli @commitlint/config-conventional

# === Claude Code ===
npm install -g @anthropic-ai/claude-code

# === Overstory + os-eco ecosystem (installed via bun) ===
# overstory requires bun runtime and has undeclared deps on seeds, mulch, and canopy CLIs
bun install -g @os-eco/overstory-cli @os-eco/mulch-cli @os-eco/seeds-cli @os-eco/canopy-cli @beads/bd

# === Go ===
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
