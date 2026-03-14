package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/mjamalu/snowctl/internal/client"
	"github.com/mjamalu/snowctl/internal/output"
	"github.com/mjamalu/snowctl/internal/registry"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs <resource> <name-or-sysid>",
	Short: "Show the audit trail (field change history) for a record",
	Long: `Show a chronological audit trail of all field changes on a ServiceNow record.

Queries the sys_audit table filtered by the target record's sys_id.

Examples:
  snowctl logs incident INC0012345
  snowctl logs incident INC0012345 --tail 10
  snowctl logs incident INC0012345 --field state
  snowctl logs incident INC0012345 --follow
  snowctl logs inc INC0012345 -f --field priority -o json`,
	Args: cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return resourceNames(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: runLogs,
}

var (
	logsTail   int
	logsFollow bool
	logsField  string
)

func init() {
	logsCmd.Flags().IntVar(&logsTail, "tail", 50, "number of most recent audit entries to show")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "poll for new changes (like tail -f)")
	logsCmd.Flags().StringVar(&logsField, "field", "", "filter to changes on a specific field")

	rootCmd.AddCommand(logsCmd)
}

// auditResource is the resource definition for formatting audit log output.
var auditResource = &registry.Resource{
	Plural:   "audit-entries",
	Singular: "audit-entry",
	Table:    "sys_audit",
	DefaultColumns: []registry.Column{
		{Header: "TIMESTAMP", Field: "sys_created_on", Width: 19},
		{Header: "USER", Field: "user", Width: 20},
		{Header: "FIELD", Field: "fieldname", Width: 20},
		{Header: "OLD VALUE", Field: "oldvalue", Width: 25},
		{Header: "NEW VALUE", Field: "newvalue", Width: 25},
	},
}

func runLogs(cmd *cobra.Command, args []string) error {
	res := resolveResource(args[0])
	identifier := args[1]

	c, err := getClient()
	if err != nil {
		return err
	}

	// Resolve the identifier to a sys_id
	sysID, displayName, err := resolveToSysID(c, res, identifier)
	if err != nil {
		return err
	}

	query := buildAuditQuery(res.Table, sysID, logsField)

	if !logsFollow {
		// One-shot: fetch and display
		return fetchAndDisplayLogs(c, query, displayName)
	}

	// Follow mode: poll loop
	return followLogs(c, query, displayName)
}

func buildAuditQuery(table, sysID, fieldFilter string) string {
	parts := []string{
		fmt.Sprintf("tablename=%s", table),
		fmt.Sprintf("documentkey=%s", sysID),
	}
	if fieldFilter != "" {
		parts = append(parts, fmt.Sprintf("fieldname=%s", fieldFilter))
	}
	parts = append(parts, "ORDERBYDESCsys_created_on")
	return strings.Join(parts, "^")
}

func fetchAndDisplayLogs(c *client.SNClient, query, displayName string) error {
	opts := client.ListOptions{
		Query:        query,
		Fields:       []string{"sys_created_on", "user", "fieldname", "oldvalue", "newvalue"},
		Limit:        logsTail,
		DisplayValue: "true",
	}

	result, err := c.List("sys_audit", opts)
	if err != nil {
		if agentMode {
			ctx, _ := cfg.ActiveContext()
			instance := ""
			if ctx != nil {
				instance = ctx.Instance
			}
			return output.FormatAgentError(os.Stdout, "API_ERROR", err.Error(), nil,
				&output.AgentContext{Verb: "logs", Resource: "sys_audit", Instance: instance})
		}
		return err
	}

	if len(result.Records) == 0 {
		fmt.Fprintf(os.Stderr, "No audit entries found for %s.\n", displayName)
		return nil
	}

	// Reverse so oldest is first (chronological order)
	reverseRecords(result.Records)

	if agentMode {
		ctx, _ := cfg.ActiveContext()
		instance := ""
		if ctx != nil {
			instance = ctx.Instance
		}
		f := &output.AgentFormatter{
			Verb:     "logs",
			Resource: "audit-entries",
			Instance: instance,
			Total:    result.TotalCount,
			HasMore:  result.TotalCount > logsTail,
			Limit:    logsTail,
		}
		return f.FormatList(os.Stdout, result.Records, auditResource)
	}

	formatter := getFormatter()
	if err := formatter.FormatList(os.Stdout, result.Records, auditResource); err != nil {
		return err
	}

	if result.TotalCount > logsTail {
		fmt.Fprintf(os.Stderr, "\nShowing %d of %d audit entries. Use --tail to see more.\n",
			len(result.Records), result.TotalCount)
	}

	return nil
}

func followLogs(c *client.SNClient, query, displayName string) error {
	fmt.Fprintf(os.Stderr, "Following audit trail for %s (Ctrl+C to stop)...\n\n", displayName)

	// Fetch initial batch
	if err := fetchAndDisplayLogs(c, query, displayName); err != nil {
		return err
	}

	// Track the latest timestamp we've seen
	lastTimestamp := ""

	// Set up signal handler for graceful exit
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigCh:
			fmt.Fprintln(os.Stderr, "\nStopped.")
			return nil
		case <-ticker.C:
			// Build query for entries newer than lastTimestamp
			pollQuery := query
			if lastTimestamp != "" {
				// Insert a timestamp filter before the ORDERBY
				pollQuery = strings.Replace(pollQuery,
					"^ORDERBYDESCsys_created_on",
					fmt.Sprintf("^sys_created_on>%s^ORDERBYDESCsys_created_on", lastTimestamp),
					1)
			}

			opts := client.ListOptions{
				Query:        pollQuery,
				Fields:       []string{"sys_created_on", "user", "fieldname", "oldvalue", "newvalue"},
				Limit:        100,
				DisplayValue: "true",
			}

			result, err := c.List("sys_audit", opts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Poll error: %s\n", err)
				continue
			}

			if len(result.Records) == 0 {
				continue
			}

			// Reverse to chronological order
			reverseRecords(result.Records)

			// Update the latest timestamp
			newest := result.Records[len(result.Records)-1]
			if ts, ok := newest["sys_created_on"].(string); ok {
				lastTimestamp = ts
			}

			// Print new entries (no header on subsequent polls)
			formatter := getFormatter()
			formatter.FormatList(os.Stdout, result.Records, auditResource)
		}
	}
}

func reverseRecords(records []map[string]interface{}) {
	for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}
}
