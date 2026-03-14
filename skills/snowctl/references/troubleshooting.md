# Troubleshooting snowctl

## Installation

### Building from source

```bash
# Requires Go 1.22+
git clone https://github.com/mjamal14/snowctl.git
cd snowctl
make build
# Binary at ./bin/snowctl
```

### Go not found

If `go` is not in your PATH:

```bash
# Install Go to home directory
curl -sL https://go.dev/dl/go1.22.5.linux-amd64.tar.gz -o /tmp/go.tar.gz
mkdir -p ~/go-sdk && tar -C ~/go-sdk -xzf /tmp/go.tar.gz
export PATH=$HOME/go-sdk/go/bin:$HOME/go/bin:$PATH
```

## Connection Issues

### Run doctor first

```bash
snowctl doctor
```

This checks:
1. DNS resolution for the instance hostname
2. HTTPS connectivity
3. API authentication (tries listing sys_properties with limit 1)

### DNS failure

- Verify the instance URL: `snowctl config view`
- Check if the instance is reachable: `curl -I https://devXXXXX.service-now.com`
- PDIs hibernate after inactivity — wake it at developer.servicenow.com

### HTTPS failure

- Check if a VPN is required
- Verify the URL doesn't have a trailing slash or path

### Authentication failure (HTTP 401)

- Check env vars: `echo $SNOWCTL_USERNAME $SNOWCTL_PASSWORD | wc -w` (should be 2)
- Verify credentials in config: `snowctl config view` (password shown if in config)
- PDI default user is `admin`

## API Errors

### HTTP 403: Access Denied

Common causes:
1. **Data Policy violation**: A mandatory field is missing. The error message will say which field.
   ```bash
   # Check required fields for a table
   snowctl get sys_dictionary --query "name=incident^mandatory=true" --fields "element,column_label"
   ```

2. **Invalid field values**: e.g., wrong `close_code`. Check valid values:
   ```bash
   snowctl get sys_choice --query "name=<table>^element=<field>" --fields "label,value" -o json
   ```

3. **State machine violation**: Trying to jump states (e.g., New -> Resolved directly).

4. **ACL restriction**: The user lacks the required role. Check in the ServiceNow UI under System Security > ACL.

### HTTP 404: Not Found

- Verify the table exists: `snowctl get sys_db_object --query "name=<table>" --limit 1`
- Verify the record exists: `snowctl get <resource> --query "number=<value>" --limit 1`
- Check for typos in the sys_id (must be exactly 32 hex characters)

### HTTP 429: Rate Limited

- Reduce `--limit` values
- Add `--query` filters to narrow results
- Wait and retry

## Debug Mode

Use `--debug` to see the actual HTTP requests:

```bash
snowctl get incidents --debug
# DEBUG: GET https://dev12345.service-now.com/api/now/table/incident?sysparm_limit=50&...
```

This helps diagnose:
- Incorrect URLs
- Wrong query parameters
- Auth header issues
