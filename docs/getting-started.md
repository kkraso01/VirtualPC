# Getting Started (release/v1)

This guide is the shortest path for first-time users.

## 1) Install and verify

```bash
make build
./scripts/install.sh
vpc doctor
```

## 2) Start daemon

```bash
virtualpcd
vpc daemon status
```

## 3) Create and start a machine

```bash
MID=$(vpc machine create --profile minimal-shell | jq -r '.id')
vpc machine start "$MID"
```

## 4) Choose provider profile and skill, then start agent

```bash
vpc provider list
vpc skill list
vpc agent start --machine "$MID" --goal "fix failing tests" --provider-profile ollama-local --skill coding --approval
```

## 5) Inspect approvals/logs

```bash
vpc agent approvals
vpc agent approve <session-id> <approval-id>
vpc agent deny <session-id> <approval-id>
vpc agent logs <session-id>
```

## 6) Snapshot/fork workflow

```bash
SID=$(vpc snapshot create "$MID" | jq -r '.id')
vpc machine fork "$SID"
```
