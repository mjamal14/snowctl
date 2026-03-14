# snowctl — kubectl-style CLI for ServiceNow

## What is snowctl?

A Go CLI tool for managing ServiceNow resources using verb-noun syntax. Single binary at `./bin/snowctl`.

## Build

```bash
make build    # outputs to bin/snowctl
make test     # run tests
```

## Command Reference

### Read

```bash
snowctl get <resource> [--query <q>] [--limit <n>] [--offset <n>] [--fields <f1,f2>] [-o json|yaml|table]
snowctl describe <resource> <name-or-sysid>
snowctl logs <resource> <name-or-sysid> [--tail <n>] [--follow] [--field <name>]
```

### Write

```bash
snowctl create <resource> --set key=value [--set key=value ...]
snowctl create <resource> --from-json '{"key":"value"}'
snowctl edit <resource> <name-or-sysid>          # opens in $EDITOR
snowctl apply -f <file.yaml|dir/> [--dry-run]
snowctl delete <resource> <name-or-sysid> [--yes]
```

### Config

```bash
snowctl config set-context <name> --instance <url> --username <user>
snowctl config use-context <name>
snowctl config get-contexts
snowctl config current-context
snowctl config view
snowctl config delete-context <name>
```

### Utility

```bash
snowctl doctor       # check connectivity + auth
snowctl commands     # machine-readable command catalog (JSON)
snowctl version
```

### Global Flags

- `-o, --output`: table (default), json, yaml
- `--agent`: structured JSON envelope for AI agents
- `--context`: override active context
- `--config`: override config file path
- `--debug`: verbose HTTP logging

## Resources

| Name | Alias | Table |
|------|-------|-------|
| incidents | inc | incident |
| changes | chg | change_request |
| problems | prb | problem |
| requests | req | sc_request |
| request-items | ritm | sc_req_item |
| tasks | -- | task |
| users | usr | sys_user |
| groups | grp | sys_user_group |
| cis | -- | cmdb_ci |
| servers | srv | cmdb_ci_server |
| applications | app | cmdb_ci_appl |
| services | svc | cmdb_ci_service |
| kb-articles | kb | kb_knowledge |
| catalog-items | catitem | sc_cat_item |

Any unrecognized name is treated as a raw ServiceNow table name (e.g. `snowctl get sys_audit`).

## Auth

Set `SNOWCTL_PASSWORD` env var (preferred) or put password in config (not recommended). Username comes from config or `SNOWCTL_USERNAME`.

## Project Structure

- `cmd/` — Cobra commands (one file per verb)
- `internal/client/` — ServiceNow Table API HTTP client
- `internal/config/` — Config loading, context switching
- `internal/registry/` — Resource name → table mapping
- `internal/output/` — Formatters (table, JSON, YAML, agent envelope)

## Manifest Format (for `apply`)

```yaml
apiVersion: snowctl/v1
kind: Incident
metadata:
  number: INC0012345   # if present, updates existing record
spec:
  short_description: "Description here"
  priority: 1
```
