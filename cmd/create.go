package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <resource> [flags]",
	Short: "Create a new resource in ServiceNow",
	Long: `Create a new record in a ServiceNow table.

Fields are passed as --set key=value pairs or as JSON via --from-json.

Examples:
  snowctl create incident --set short_description="DB outage" --set priority=1
  snowctl create incident --set short_description="Outage" --set assignment_group="DBA Team"
  snowctl create change --set short_description="Deploy v2" --set type=Standard
  snowctl create user --from-json '{"user_name":"jdoe","first_name":"John","last_name":"Doe"}'`,
	Args: cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return resourceNames(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: runCreate,
}

var (
	createSetFlags []string
	createFromJSON string
)

func init() {
	createCmd.Flags().StringArrayVar(&createSetFlags, "set", nil, "set a field value (key=value), can be repeated")
	createCmd.Flags().StringVar(&createFromJSON, "from-json", "", "JSON string with field values")

	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	res := resolveResource(args[0])

	data := make(map[string]interface{})

	// Parse --from-json
	if createFromJSON != "" {
		if err := json.Unmarshal([]byte(createFromJSON), &data); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
	}

	// Parse --set flags (override JSON values)
	for _, kv := range createSetFlags {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid --set format %q, expected key=value", kv)
		}
		data[parts[0]] = parts[1]
	}

	if len(data) == 0 {
		return fmt.Errorf("no fields specified; use --set key=value or --from-json")
	}

	c, err := getClient()
	if err != nil {
		return err
	}

	record, err := c.Create(res.Table, data)
	if err != nil {
		return err
	}

	// Show the created record
	formatter := getFormatter()
	if err := formatter.FormatSingle(os.Stdout, record, res); err != nil {
		return err
	}

	// Print the identifier
	if id := res.IdentifierField; id != "" {
		if val, ok := record[id]; ok {
			fmt.Fprintf(os.Stderr, "\n%s %v created.\n", res.Singular, val)
		}
	}

	return nil
}
