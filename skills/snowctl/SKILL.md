---
name: snowctl
description: Manage ServiceNow environments using snowctl - configure instance contexts with basic or OAuth auth, and run kubectl-style operations (get/describe/create/edit/apply/delete/logs/commands) for incidents, changes, problems, requests, CMDB CIs, users, groups, knowledge articles, catalog items, and any arbitrary ServiceNow table. Use when the user wants to control ServiceNow resources via snowctl.
---

# ServiceNow Control with snowctl

Operate `snowctl`, the kubectl-style CLI for ServiceNow. This skill teaches core snowctl command patterns and operations.

## Recommended Initialization

At the start of a task, run these checks to establish context and credentials:

```bash
# Show current context
snowctl config current-context

# Show full config (instance URL, auth type, defaults)
snowctl config view

# Verify connectivity and auth
snowctl doctor

# Discover all available commands and resources
snowctl commands -o json
```

This displays:
- Current context name and ServiceNow instance URL
- Authentication type (basic or OAuth)
- All available commands, resources, and aliases

## ServiceNow Query Reference

Before writing queries, consult [references/query-syntax.md](references/query-syntax.md) for the full encoded query syntax.

If there is any conflict between assumptions and the reference, prefer the reference.

## Prerequisites

snowctl requires Go 1.22+ to build from source:

```bash
git clone https://github.com/mjamal14/snowctl.git
cd snowctl
make build
# binary at ./bin/snowctl
```

Configure an instance:

```bash
snowctl config set-context dev --instance https://devXXXXX.service-now.com --username admin
export SNOWCTL_PASSWORD='your-password'
snowctl doctor
```

## Resources & Commands

### Available Resources

snowctl uses a uniform pattern for all resource types. Any unrecognized resource name is treated as a raw ServiceNow table name.

| Resource | Aliases | Table | Category |
|----------|---------|-------|----------|
| incidents | inc | incident | ITSM |
| changes | chg | change_request | ITSM |
| problems | prb | problem | ITSM |
| requests | req | sc_request | ITSM |
| request-items | ritm | sc_req_item | ITSM |
| tasks | — | task | ITSM |
| users | usr | sys_user | Platform |
| groups | grp | sys_user_group | Platform |
| cis | — | cmdb_ci | CMDB |
| servers | srv | cmdb_ci_server | CMDB |
| applications | app | cmdb_ci_appl | CMDB |
| services | svc | cmdb_ci_service | CMDB |
| kb-articles | kb | kb_knowledge | Platform |
| catalog-items | catitem | sc_cat_item | Platform |

**Ad-hoc table access** — any unrecognized name queries the raw table:

```bash
snowctl get sys_audit --query "tablename=incident" --limit 5
snowctl get sys_dictionary --query "name=incident^element=state"
snowctl get sys_choice --query "name=incident^element=close_code" --fields "label,value"
```

### Command Verbs

| Verb | Description | Example |
|------|-------------|---------|
| **get** | List resources | `snowctl get incidents --query "priority=1"` |
| **describe** | Show resource details | `snowctl describe incident INC0012345` |
| **create** | Create new resource | `snowctl create incident --set short_description="Outage"` |
| **edit** | Edit interactively in $EDITOR | `snowctl edit incident INC0012345` |
| **apply** | Create/update from YAML manifest | `snowctl apply -f incident.yaml` |
| **delete** | Delete resource | `snowctl delete incident INC0012345 --yes` |
| **logs** | Show audit trail (field changes) | `snowctl logs incident INC0012345` |
| **config** | Manage contexts/settings | `snowctl config use-context prod` |
| **doctor** | Check connectivity + auth | `snowctl doctor` |
| **commands** | Machine-readable command catalog | `snowctl commands -o json` |

## Key Concepts for AI Agents

### Output Modes

```bash
# Agent envelope mode (structured JSON with ok/result/error/context)
--agent          # Structured JSON envelope for AI consumption

# Machine-readable formats
-o json          # JSON output
-o yaml          # YAML output

# Human-readable (default)
-o table         # Table format (default)
```

**For AI agents, prefer:** `snowctl <command> --agent` or `snowctl <command> -o json`

The `--agent` envelope provides structured metadata alongside results:

```json
{
  "ok": true,
  "result": [ ... ],
  "context": {
    "verb": "get", "resource": "incidents",
    "instance": "https://dev12345.service-now.com",
    "total": 23, "has_more": true,
    "offset": 0, "limit": 50,
    "suggestions": ["Use '--offset 50' to see the next page"]
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

### Creating Records

Two approaches:

```bash
# Inline with --set flags
snowctl create incident --set short_description="DB outage" --set priority=1 --set urgency=1

# From JSON
snowctl create user --from-json '{"user_name":"jdoe","first_name":"John","last_name":"Doe"}'
```

### Declarative Apply (YAML Manifests)

See [references/resources/manifests.md](references/resources/manifests.md) for the full manifest specification.

```yaml
apiVersion: snowctl/v1
kind: Incident
metadata:
  number: INC0012345    # present = update, absent = create
spec:
  state: "6"
  close_code: "Resolved by caller"
  close_notes: "Issue resolved after patch."
```

```bash
# Apply single file
snowctl apply -f incident.yaml

# Apply directory of manifests
snowctl apply -f manifests/

# Preview without executing
snowctl apply -f incident.yaml --dry-run

# Multi-document YAML (separated by ---) supported
snowctl apply -f batch-operations.yaml
```

### Audit Trail (logs)

Query the `sys_audit` table for field-level change history:

```bash
# Last 50 changes
snowctl logs incident INC0012345

# Filter to state changes only
snowctl logs incident INC0012345 --field state

# Last 10 changes
snowctl logs incident INC0012345 --tail 10

# Follow mode (poll every 5s, like tail -f)
snowctl logs incident INC0012345 --follow
```

### Authentication & Configuration

```bash
# Set up a context
snowctl config set-context dev --instance https://dev12345.service-now.com --username admin

# Switch contexts
snowctl config use-context prod

# List all contexts (* = active)
snowctl config get-contexts

# Credentials (env vars take priority over config)
export SNOWCTL_PASSWORD='password'
export SNOWCTL_USERNAME='admin'
```

## Quick Reference: ServiceNow Queries

**Required workflow for querying:**
1. First consult [references/query-syntax.md](references/query-syntax.md)
2. Build the encoded query using documented operators
3. Execute with `snowctl get <resource> --query "..." -o json`

```bash
# Active P1 incidents
snowctl get incidents --query "priority=1^active=true" -o json

# Updated in last 7 days
snowctl get incidents --query "sys_updated_on>javascript:gs.daysAgoStart(7)"

# Filter by assignment group
snowctl get incidents --query "assignment_group.name=Software^active=true"

# LIKE operator
snowctl get incidents --query "short_descriptionLIKEdatabase"

# Pagination
snowctl get incidents --query "active=true" --limit 20 --offset 40

# Select specific fields
snowctl get users --fields "user_name,email,name,department" -o json
```

## Common Issues

**State transitions blocked (HTTP 403):**
- ServiceNow Data Policies may block direct state jumps
- Check valid values: `snowctl get sys_choice --query "name=<table>^element=<field>" --fields "label,value"`
- Some transitions require intermediate states or mandatory fields (e.g., close_code for resolved)

**Record not found:**
- Verify the number/name: `snowctl get <resource> --query "number=<value>" --limit 1`
- Try with sys_id directly: `snowctl describe <resource> <32-char-hex-id>`

**Authentication failures:**
- Run `snowctl doctor` to diagnose
- Check env vars: `echo $SNOWCTL_PASSWORD | wc -c` (non-zero = set)
- Verify instance URL: `snowctl config view`

**Unknown table:**
- Any resource name not in the built-in list is treated as a raw table name
- Check if the table exists: `snowctl get sys_db_object --query "name=<table_name>" --limit 1`

## Additional Resources

- **Query syntax**: [references/query-syntax.md](references/query-syntax.md)
- **Manifest format**: [references/resources/manifests.md](references/resources/manifests.md)
- **Troubleshooting**: [references/troubleshooting.md](references/troubleshooting.md)
- **CLI help**: `snowctl --help`, `snowctl <command> --help`
- **dtctl (inspiration)**: [github.com/dynatrace-oss/dtctl](https://github.com/dynatrace-oss/dtctl)

## Safety Reminders

- Use `--agent` or `-o json` for machine/AI consumption
- Confirm context with `snowctl config current-context` before destructive ops
- Use `snowctl apply --dry-run` before applying manifests
- Use `--yes` flag only when you're certain about deletions
- Run `snowctl doctor` after context switches to verify connectivity
