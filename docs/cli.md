# CLI

## Daemon
- `vpc daemon status`
- `vpc config inspect`

## Profiles
- `vpc profile list`

## Machines
- `vpc machine create --profile <name>`
- `vpc machine list`
- `vpc machine inspect <id>`
- `vpc machine start <id>`
- `vpc machine stop <id>`
- `vpc machine destroy <id>`
- `vpc machine exec <id> -- <command...>`
- `vpc machine shell <id>`
- `vpc machine logs <id>`
- `vpc machine ps <id>`
- `vpc machine assign <id> --project <project-id>`
- `vpc machine fork <snapshot-id>`

## Projects/services
- `vpc project create <name>`
- `vpc project list`
- `vpc service create --machine <id> --name <name> --image <image>`
- `vpc service list --machine <id>`

## Snapshots/tasks
- `vpc snapshot create <machine-id>`
- `vpc snapshot list`
- `vpc snapshot inspect <snapshot-id>`
- `vpc task create --machine <id> --goal "..."`
- `vpc task run <task-id>`
- `vpc task inspect <task-id>`
- `vpc task logs <task-id>`
