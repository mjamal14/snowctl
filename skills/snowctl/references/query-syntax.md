# ServiceNow Encoded Query Syntax

snowctl passes `--query` values directly to the ServiceNow Table API's `sysparm_query` parameter.

## Operators

| Operator | Meaning | Example |
|----------|---------|---------|
| `=` | Equals | `priority=1` |
| `!=` | Not equals | `state!=7` |
| `<` | Less than | `priority<3` |
| `>` | Greater than | `sys_updated_on>2026-01-01` |
| `<=` | Less than or equal | `impact<=2` |
| `>=` | Greater than or equal | `priority>=3` |
| `LIKE` | Contains | `short_descriptionLIKEdatabase` |
| `STARTSWITH` | Starts with | `numberSTARTSWITHINC` |
| `ENDSWITH` | Ends with | `emailENDSWITH@acme.com` |
| `IN` | In list | `priorityIN1,2,3` |
| `NOT IN` | Not in list | `stateNOT IN6,7` |
| `ISEMPTY` | Is empty | `assigned_toISEMPTY` |
| `ISNOTEMPTY` | Is not empty | `assignment_groupISNOTEMPTY` |
| `BETWEEN` | Between values | `priorityBETWEEN1@3` |

## Logical Operators

| Operator | Meaning | Example |
|----------|---------|---------|
| `^` | AND | `priority=1^active=true` |
| `^OR` | OR | `state=1^ORstate=2` |
| `^NQ` | New query (union) | `priority=1^NQpriority=2` |

## Ordering

| Operator | Example |
|----------|---------|
| `ORDERBY` | `^ORDERBYsys_created_on` |
| `ORDERBYDESC` | `^ORDERBYDESCsys_updated_on` |

## JavaScript Functions

ServiceNow supports server-side JavaScript in queries:

```bash
# Records updated in the last 7 days
--query "sys_updated_on>javascript:gs.daysAgoStart(7)"

# Records created today
--query "sys_created_on>=javascript:gs.beginningOfToday()"

# Records assigned to the current user
--query "assigned_to=javascript:gs.getUserID()"
```

## Dot-Walking (Reference Fields)

Query through reference fields using dot notation:

```bash
# Incidents assigned to the "Software" group
--query "assignment_group.name=Software"

# Incidents where the caller is in a specific company
--query "caller_id.company.name=ACME North America"

# CIs managed by a specific user
--query "managed_by.user_name=john.smith"
```

## Common Patterns

```bash
# Active P1 incidents, newest first
snowctl get incidents --query "priority=1^active=true^ORDERBYDESCsys_updated_on"

# Unassigned incidents
snowctl get incidents --query "active=true^assigned_toISEMPTY"

# Incidents in specific groups
snowctl get incidents --query "assignment_group.nameINSoftware,Hardware,Network"

# Stale incidents (not updated in 30 days)
snowctl get incidents --query "active=true^sys_updated_on<javascript:gs.daysAgoStart(30)"

# Changes scheduled for this week
snowctl get changes --query "start_date>=javascript:gs.beginningOfThisWeek()^start_dateENDSTARTSWITH"

# Users in a specific department
snowctl get users --query "department.name=Engineering^active=true"

# CMDB servers by OS
snowctl get servers --query "osLIKELinux^operational_status=1"
```
