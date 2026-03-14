# snowctl

A `kubectl`-style CLI for managing ServiceNow resources. Single binary, zero dependencies.

```
snowctl get incidents --query "priority=1^active=true" --limit 10
snowctl describe incident INC0012345
snowctl create incident --set short_description="DB outage" --set priority=1
snowctl apply -f incident.yaml
snowctl logs incident INC0012345 --follow
```


## Install

### From source

```bash
git clone https://github.com/mjamal14/snowctl.git
cd snowctl
make build
# binary is at ./bin/snowctl
```

### Prerequisites

- Go 1.22+
- A ServiceNow instance (PDI works fine)

## Quick Start

```bash
# 1. Configure your ServiceNow instance
snowctl config set-context dev \
  --instance https://dev12345.service-now.com \
  --username admin

# 2. Set your password (env var is safest)
export SNOWCTL_PASSWORD='your-password'

# 3. Verify connectivity
snowctl doctor

# 4. Start using it
snowctl get incidents
snowctl get users --fields "user_name,email,name" -o json
```

## Commands

### Read

```bash
snowctl get <resource> [--query <q>] [--limit <n>] [--offset <n>] [--fields <f1,f2>]
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

| Flag | Description |
|------|-------------|
| `-o, --output` | Output format: `table` (default), `json`, `yaml` |
| `--agent` | Structured JSON envelope for AI agents |
| `--context` | Override the active context |
| `--config` | Override config file path |
| `--debug` | Verbose HTTP request logging |

## Resources

snowctl ships with 14 built-in resource types that map to ServiceNow tables:

| Resource | Alias | ServiceNow Table | Category |
|----------|-------|-------------------|----------|
| `incidents` | `inc` | `incident` | ITSM |
| `changes` | `chg` | `change_request` | ITSM |
| `problems` | `prb` | `problem` | ITSM |
| `requests` | `req` | `sc_request` | ITSM |
| `request-items` | `ritm` | `sc_req_item` | ITSM |
| `tasks` | — | `task` | ITSM |
| `users` | `usr` | `sys_user` | Platform |
| `groups` | `grp` | `sys_user_group` | Platform |
| `cis` | — | `cmdb_ci` | CMDB |
| `servers` | `srv` | `cmdb_ci_server` | CMDB |
| `applications` | `app` | `cmdb_ci_appl` | CMDB |
| `services` | `svc` | `cmdb_ci_service` | CMDB |
| `kb-articles` | `kb` | `kb_knowledge` | Platform |
| `catalog-items` | `catitem` | `sc_cat_item` | Platform |

**Any unrecognized name is treated as a raw ServiceNow table name**, so you can query any table:

```bash
snowctl get sys_audit --query "tablename=incident" --limit 5
snowctl get sys_dictionary --query "name=incident^element=state" --fields "element,column_label"
snowctl get sys_choice --query "name=incident^element=close_code" --fields "label,value"
```

## YAML Manifests

snowctl supports declarative resource management via YAML manifests, similar to `kubectl apply`.

### Manifest Format

```yaml
apiVersion: snowctl/v1
kind: <ResourceKind>
metadata:
  number: INC0012345      # optional: if present, updates existing record
  # or
  sys_id: abc123def456    # optional: alternative way to identify existing record
spec:
  field_name: value
  another_field: value
```

- `apiVersion`: Always `snowctl/v1`
- `kind`: The resource type — `Incident`, `Change`, `Problem`, `User`, `Server`, etc. (matches singular resource name, capitalized)
- `metadata`: Optional identifiers. If `number`, `sys_id`, or the resource's identifier field is present, snowctl updates the existing record. If absent, a new record is created.
- `spec`: The ServiceNow fields and values to set

### Create an Incident

```yaml
# new-incident.yaml
apiVersion: snowctl/v1
kind: Incident
spec:
  short_description: "Production database unresponsive"
  priority: 1
  urgency: 1
  impact: 1
  assignment_group: "Database Administration"
  description: |
    The production Oracle database cluster became unresponsive
    at approximately 09:15 UTC. All dependent services affected.
```

```bash
snowctl apply -f new-incident.yaml
# Output: incident INC0010003 created.
```

### Update an Existing Incident

```yaml
# resolve-incident.yaml
apiVersion: snowctl/v1
kind: Incident
metadata:
  number: INC0012345
spec:
  state: "6"
  close_code: "Resolved by caller"
  close_notes: "User confirmed the issue is no longer occurring after the patch."
```

```bash
snowctl apply -f resolve-incident.yaml
# Output: incident INC0012345 updated.
```

### Create a Change Request

```yaml
# deploy-change.yaml
apiVersion: snowctl/v1
kind: Change
spec:
  short_description: "Deploy database patch v2.3.1"
  type: Standard
  assignment_group: "Database Administration"
  description: |
    Deploy the latest database patch to address the
    connection pooling issue identified in PRB0001234.
  start_date: "2026-03-20 06:00:00"
  end_date: "2026-03-20 08:00:00"
```

### Multi-Document Manifests

Multiple resources can be defined in a single file, separated by `---`:

```yaml
# incident-and-change.yaml
apiVersion: snowctl/v1
kind: Incident
metadata:
  number: INC0012345
spec:
  state: "6"
  close_code: "Resolved by change"
  close_notes: "Fixed by CHG0001234"
---
apiVersion: snowctl/v1
kind: Change
spec:
  short_description: "Deploy hotfix for INC0012345"
  type: Emergency
  assignment_group: "Cloud Platform"
```

```bash
snowctl apply -f incident-and-change.yaml
# Output:
# incident INC0012345 updated.
# change CHG0005678 created.
```

### Applying a Directory

Point `apply` at a directory to process all `.yaml`/`.yml` files in it:

```bash
ls manifests/
# incident.yaml  change.yaml  user.yaml

snowctl apply -f manifests/
# Output:
# incident INC0010004 created.
# change CHG0005679 created.
# user jdoe created.
```

### Dry Run

Preview what would happen without making changes:

```bash
snowctl apply -f incident.yaml --dry-run
# Output: [dry-run] would create incident
```

### Bulk Operations Example

Close a batch of stale incidents:

```yaml
# close-stale.yaml
apiVersion: snowctl/v1
kind: Incident
metadata:
  number: INC0009009
spec:
  state: "6"
  close_code: "No resolution provided"
  close_notes: "Stale incident sweep — 7+ years with no activity."
---
apiVersion: snowctl/v1
kind: Incident
metadata:
  number: INC0007002
spec:
  state: "6"
  close_code: "No resolution provided"
  close_notes: "Stale incident sweep — 7+ years with no activity."
---
apiVersion: snowctl/v1
kind: Incident
metadata:
  number: INC0000048
spec:
  state: "6"
  close_code: "No resolution provided"
  close_notes: "Stale incident sweep — 10+ years with no activity."
```

```bash
snowctl apply -f close-stale.yaml
# incident INC0009009 updated.
# incident INC0007002 updated.
# incident INC0000048 updated.
```

## Audit Logs

The `logs` command shows field-level change history for any record by querying the `sys_audit` table:

```bash
# Show last 50 changes (default)
snowctl logs incident INC0012345

# Show only state changes
snowctl logs incident INC0012345 --field state

# Show last 10 changes
snowctl logs incident INC0012345 --tail 10

# Follow mode — poll for new changes every 5 seconds
snowctl logs incident INC0012345 --follow

# JSON output
snowctl logs incident INC0012345 -o json
```

Output:

```
TIMESTAMP            USER                  FIELD                 OLD VALUE                  NEW VALUE
2026-03-14 08:15:01  admin                 state                 New                        In Progress
2026-03-14 08:16:30  admin                 assignment_group                                 Database Administration
2026-03-14 09:45:12  admin                 priority              3 - Moderate               1 - Critical
2026-03-14 10:30:00  admin                 state                 In Progress                Resolved
```

## Query Syntax

snowctl passes queries directly to the ServiceNow Table API's `sysparm_query` parameter. The syntax uses `^` as AND and `^OR` as OR:

```bash
# Active P1 incidents
snowctl get incidents --query "priority=1^active=true"

# Incidents updated in the last 7 days
snowctl get incidents --query "sys_updated_on>javascript:gs.daysAgoStart(7)"

# Incidents assigned to a specific group
snowctl get incidents --query "assignment_group.name=Software^active=true"

# Multiple conditions with OR
snowctl get incidents --query "state=1^ORstate=2"

# Ordering
snowctl get incidents --query "active=true^ORDERBYDESCsys_updated_on"

# LIKE operator
snowctl get incidents --query "short_descriptionLIKEdatabase"
```

## Configuration

Config file location (respects `XDG_CONFIG_HOME`):
- Linux: `~/.config/snowctl/config.yaml`
- macOS: `~/Library/Application Support/snowctl/config.yaml`
- Windows: `%LOCALAPPDATA%\snowctl\config.yaml`

Override with `SNOWCTL_CONFIG_DIR` env var or `--config` flag.

### Config File Format

```yaml
apiVersion: v1
current-context: dev

contexts:
  - name: dev
    instance: https://dev12345.service-now.com
    auth:
      type: basic
      username: admin
    defaults:
      limit: 50

  - name: prod
    instance: https://prod.service-now.com
    auth:
      type: basic
      username: svc_snowctl
    defaults:
      limit: 25
      output: table

defaults:
  limit: 50
  display-value: "true"
  output: table
  editor: vim
  watch-interval: 5s
```

### Authentication

Credentials are resolved in this order:

1. **Environment variables** (highest priority): `SNOWCTL_USERNAME`, `SNOWCTL_PASSWORD`
2. **Config file**: `username` and `password` fields (not recommended for passwords)

```bash
# Recommended: set password via env var
export SNOWCTL_PASSWORD='your-password'

# Optional: override username too
export SNOWCTL_USERNAME='admin'
```

## AI Agent Integration

snowctl includes first-class support for AI agents (Claude Code, Cursor, Copilot, etc.).

### Agent Output Mode

The `--agent` flag wraps all output in a structured JSON envelope:

```bash
snowctl get incidents --query "priority=1" --agent
```

```json
{
  "ok": true,
  "result": [
    {
      "number": "INC0012345",
      "state": "In Progress",
      "priority": "1 - Critical",
      "short_description": "Production database unresponsive"
    }
  ],
  "context": {
    "verb": "get",
    "resource": "incidents",
    "instance": "https://dev12345.service-now.com",
    "total": 23,
    "has_more": true,
    "offset": 0,
    "limit": 50,
    "suggestions": [
      "Use '--offset 50' to see the next page"
    ]
  }
}
```

Error responses follow the same structure:

```json
{
  "ok": false,
  "error": {
    "code": "AUTH_FAILURE",
    "message": "Authentication failed (HTTP 401)",
    "suggestions": [
      "Check credentials with 'snowctl config view'",
      "Run 'snowctl doctor' to test connectivity"
    ]
  }
}
```

### Command Catalog

Get a machine-readable list of all commands and resources for LLM tool discovery:

```bash
snowctl commands -o json
```

## Project Structure

```
cmd/              Cobra commands (one file per verb)
internal/
  client/         ServiceNow Table API HTTP client
  config/         Config loading, context switching
  registry/       Resource name -> table mapping
  output/         Formatters (table, JSON, YAML, agent envelope)
pkg/
  commands/       Public command catalog (importable)
skills/
  snowctl/        AI agent skill definition
examples/
  manifests/      Example YAML manifests
```

## Building

```bash
make build        # build to ./bin/snowctl
make test         # run tests with race detector
make lint         # run golangci-lint
make clean        # remove build artifacts
```

## License

MIT
