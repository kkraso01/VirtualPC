#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "[agent-check] loading provider profiles"
go test ./agent/capabilities -run TestLoaderAndPolicyBindings -count=1

echo "[agent-check] validating dispatcher mocked flows"
go test ./agent/capabilities -run 'TestDispatchBuiltin|TestDispatchHTTPWithAllowlist|TestDispatchMCPPrompt|TestApprovalRequiredFlow|TestHTTPToolAllowlistBlocked|TestLocalToolTimeout' -count=1

echo "[agent-check] validating approval persistence and session recovery"
go test ./agent/controller -run 'TestApprovalPendingApproveDenyFlow|TestSessionRecoveryWithPendingApproval|TestBuildContextSkillOverlay' -count=1

echo "[agent-check] validating mcp mocked failures"
go test ./agent/mcp -run 'TestInvokeUnavailableServer|TestInvokeTimeout|TestInvokeFailureWrapped' -count=1

echo "[agent-check] complete"
