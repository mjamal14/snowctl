package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mjamalu/snowctl/internal/client"
	"github.com/mjamalu/snowctl/internal/registry"
	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe <resource> <name-or-sysid>",
	Short: "Show detailed information about a resource",
	Long: `Show all fields for a single ServiceNow record.

The identifier can be a sys_id or a human-readable identifier (e.g., INC0012345).

Examples:
  snowctl describe incident INC0012345
  snowctl describe user john.smith
  snowctl describe server 6816f79cc0a8016401c5a33be04be441`,
	Args: cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return resourceNames(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: runDescribe,
}

func init() {
	rootCmd.AddCommand(describeCmd)
}

func runDescribe(cmd *cobra.Command, args []string) error {
	res := resolveResource(args[0])
	identifier := args[1]

	c, err := getClient()
	if err != nil {
		return err
	}

	opts := client.ListOptions{
		DisplayValue: "true",
	}

	// Check if identifier looks like a sys_id (32 hex chars) or a human-readable name
	record, err := resolveRecord(c, res, identifier, opts)
	if err != nil {
		return err
	}

	formatter := getFormatter()
	return formatter.FormatSingle(os.Stdout, record, res)
}

// resolveRecord finds a record by sys_id or by its identifier field (e.g., number, user_name).
func resolveRecord(c *client.SNClient, res *registry.Resource, identifier string, opts client.ListOptions) (map[string]interface{}, error) {
	// If it looks like a sys_id (32 hex chars), try direct GET
	if isSysID(identifier) {
		record, err := c.Get(res.Table, identifier, opts)
		if err == nil {
			return record, nil
		}
		// Fall through to query-based lookup
	}

	// Query by the resource's identifier field
	idField := res.IdentifierField
	if idField == "" {
		idField = "name"
	}

	listOpts := client.ListOptions{
		Query:        fmt.Sprintf("%s=%s", idField, identifier),
		Limit:        1,
		DisplayValue: opts.DisplayValue,
		Fields:       opts.Fields,
	}

	result, err := c.List(res.Table, listOpts)
	if err != nil {
		return nil, err
	}

	if len(result.Records) == 0 {
		return nil, fmt.Errorf("%s %q not found", res.Singular, identifier)
	}

	// Re-fetch the full record by sys_id to get all fields
	sysID, ok := result.Records[0]["sys_id"].(string)
	if !ok {
		return result.Records[0], nil
	}

	return c.Get(res.Table, sysID, opts)
}

func isSysID(s string) bool {
	if len(s) != 32 {
		return false
	}
	for _, c := range strings.ToLower(s) {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}
