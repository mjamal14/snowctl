# YAML Manifest Reference

snowctl supports declarative resource management via YAML manifests, similar to `kubectl apply`.

## Manifest Structure

```yaml
apiVersion: snowctl/v1
kind: <ResourceKind>
metadata:
  <identifier_field>: <value>   # optional: presence triggers update mode
spec:
  <field>: <value>
  <field>: <value>
```

### Fields

| Field | Required | Description |
|-------|----------|-------------|
| `apiVersion` | Yes | Always `snowctl/v1` |
| `kind` | Yes | Resource type, capitalized singular (e.g., `Incident`, `Change`, `User`) |
| `metadata` | No | Record identifiers. If present, snowctl updates; if absent, creates |
| `spec` | Yes | ServiceNow field names and values to set |

### Kind Mapping

The `kind` field maps to the resource's singular name (capitalized):

| Kind | Resource | Table |
|------|----------|-------|
| `Incident` | incidents | incident |
| `Change` | changes | change_request |
| `Problem` | problems | problem |
| `Request` | requests | sc_request |
| `Request-Item` | request-items | sc_req_item |
| `Task` | tasks | task |
| `User` | users | sys_user |
| `Group` | groups | sys_user_group |
| `Ci` | cis | cmdb_ci |
| `Server` | servers | cmdb_ci_server |
| `Application` | applications | cmdb_ci_appl |
| `Service` | services | cmdb_ci_service |
| `Kb-Article` | kb-articles | kb_knowledge |
| `Catalog-Item` | catalog-items | sc_cat_item |

Any other kind is treated as a raw table name (lowercased).

### Metadata Identifiers

snowctl uses metadata to decide whether to create or update:

- **No metadata** or **empty metadata**: creates a new record
- **`sys_id` present**: updates by sys_id (direct)
- **Identifier field present** (e.g., `number` for incidents, `user_name` for users): looks up the sys_id first, then updates

```yaml
# CREATE — no metadata
apiVersion: snowctl/v1
kind: Incident
spec:
  short_description: "New incident"

# UPDATE — by number
apiVersion: snowctl/v1
kind: Incident
metadata:
  number: INC0012345
spec:
  state: "6"

# UPDATE — by sys_id
apiVersion: snowctl/v1
kind: Incident
metadata:
  sys_id: abc123def456abc123def456abc12345
spec:
  priority: "1"
```

## Examples

### Create an Incident

```yaml
apiVersion: snowctl/v1
kind: Incident
spec:
  short_description: "Production database unresponsive"
  priority: 1
  urgency: 1
  impact: 1
  assignment_group: "Database Administration"
  category: "Software"
  description: |
    The production Oracle database cluster became unresponsive
    at approximately 09:15 UTC. All dependent services affected.
```

### Resolve an Incident

Important: check valid `close_code` values first:

```bash
snowctl get sys_choice --query "name=incident^element=close_code" --fields "label,value" -o json
```

```yaml
apiVersion: snowctl/v1
kind: Incident
metadata:
  number: INC0012345
spec:
  state: "6"
  close_code: "Resolved by caller"
  close_notes: "User confirmed issue resolved after applying patch KB0012345."
```

### Create a Change Request

```yaml
apiVersion: snowctl/v1
kind: Change
spec:
  short_description: "Deploy database patch v2.3.1"
  type: Standard
  assignment_group: "Database Administration"
  start_date: "2026-03-20 06:00:00"
  end_date: "2026-03-20 08:00:00"
  description: |
    Deploy the latest database patch to address the
    connection pooling issue identified in PRB0001234.
```

### Create a User

```yaml
apiVersion: snowctl/v1
kind: User
spec:
  user_name: jdoe
  first_name: John
  last_name: Doe
  email: john.doe@acme.com
  department: Engineering
  active: true
```

### Reassign + Add Work Note

```yaml
apiVersion: snowctl/v1
kind: Incident
metadata:
  number: INC0000040
spec:
  assignment_group: "8a4dde73c6112278017a6a4baf547aa7"
  work_notes: "Reassigned to Software group — JavaScript error requires developer investigation."
```

Note: `assignment_group` requires a sys_id when setting via API. Look it up first:

```bash
snowctl get groups --query "name=Software" --fields "name,sys_id" -o json
```

## Multi-Document Manifests

Separate multiple resources with `---`:

```yaml
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

## Directory Apply

Apply all `.yaml` / `.yml` files in a directory:

```bash
snowctl apply -f manifests/
```

Files are processed alphabetically. Each file can contain multiple documents.

## Dry Run

Preview actions without making changes:

```bash
snowctl apply -f incident.yaml --dry-run
# [dry-run] would update incident INC0012345 (sys_id: abc123...)
# [dry-run] would create change
```

## Common Gotchas

1. **State values are strings**: Use `state: "6"` not `state: 6` — YAML may interpret bare integers differently
2. **Reference fields need sys_ids**: Fields like `assignment_group`, `assigned_to`, `caller_id` require sys_id values when writing, not display names
3. **Data Policies**: ServiceNow may enforce mandatory fields (e.g., `close_code` when resolving). Check with: `snowctl get sys_choice --query "name=<table>^element=<field>"`
4. **State machine**: Some instances enforce state transitions (New -> In Progress -> Resolved). Direct jumps may return 403
5. **Display values vs internal values**: The API returns display values when `sysparm_display_value=true`, but writes expect internal values
