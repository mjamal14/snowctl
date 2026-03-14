package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "List all available commands (machine-readable catalog for AI agents)",
	RunE:  runCommands,
}

func init() {
	rootCmd.AddCommand(commandsCmd)
}

type commandEntry struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Usage       string   `json:"usage"`
	Examples    []string `json:"examples,omitempty"`
	Category    string   `json:"category"`
}

func runCommands(cmd *cobra.Command, args []string) error {
	catalog := []commandEntry{
		{
			Name: "get", Description: "List resources from a ServiceNow table",
			Usage: "snowctl get <resource> [--query <query>] [--limit <n>] [--fields <f1,f2>] [-o json|yaml|table]",
			Examples: []string{
				`snowctl get incidents --query "priority=1^active=true" --limit 10`,
				`snowctl get users --fields "user_name,email" -o json`,
				`snowctl get sys_audit --limit 5`,
			},
			Category: "read",
		},
		{
			Name: "describe", Description: "Show detailed information about a single resource",
			Usage: "snowctl describe <resource> <name-or-sysid>",
			Examples: []string{
				`snowctl describe incident INC0012345`,
				`snowctl describe user john.smith`,
			},
			Category: "read",
		},
		{
			Name: "logs", Description: "Show the audit trail (field change history) for a record",
			Usage: "snowctl logs <resource> <name-or-sysid> [--tail <n>] [--follow] [--field <name>]",
			Examples: []string{
				`snowctl logs incident INC0012345`,
				`snowctl logs incident INC0012345 --tail 10 --field state`,
				`snowctl logs inc INC0012345 -f`,
			},
			Category: "read",
		},
		{
			Name: "create", Description: "Create a new resource in ServiceNow",
			Usage: "snowctl create <resource> --set key=value [--set key=value ...]",
			Examples: []string{
				`snowctl create incident --set short_description="Outage" --set priority=1`,
				`snowctl create user --from-json '{"user_name":"jdoe","first_name":"John"}'`,
			},
			Category: "write",
		},
		{
			Name: "edit", Description: "Edit a resource in your default editor (fetches YAML, opens editor, patches on save)",
			Usage: "snowctl edit <resource> <name-or-sysid>",
			Examples: []string{
				`snowctl edit incident INC0012345`,
			},
			Category: "write",
		},
		{
			Name: "apply", Description: "Create or update resources from YAML manifests (declarative)",
			Usage: "snowctl apply -f <file-or-dir> [--dry-run]",
			Examples: []string{
				`snowctl apply -f incident.yaml`,
				`snowctl apply -f manifests/`,
				`snowctl apply -f change.yaml --dry-run`,
			},
			Category: "write",
		},
		{
			Name: "delete", Description: "Delete a resource from ServiceNow",
			Usage: "snowctl delete <resource> <name-or-sysid> [--yes]",
			Examples: []string{
				`snowctl delete incident INC0099999 --yes`,
			},
			Category: "write",
		},
		{
			Name: "config", Description: "Manage snowctl configuration (contexts, settings)",
			Usage: "snowctl config <subcommand>",
			Examples: []string{
				`snowctl config set-context dev --instance https://dev.service-now.com --username admin`,
				`snowctl config use-context prod`,
				`snowctl config get-contexts`,
				`snowctl config view`,
			},
			Category: "config",
		},
		{
			Name: "doctor", Description: "Check connectivity and authentication with ServiceNow",
			Usage: "snowctl doctor",
			Category: "utility",
		},
	}

	// Add resource list
	type resourceEntry struct {
		Name    string   `json:"name"`
		Aliases []string `json:"aliases,omitempty"`
		Table   string   `json:"table"`
		Category string  `json:"category"`
	}

	resources := make([]resourceEntry, 0)
	for _, r := range reg.List() {
		aliases := r.Aliases
		if aliases == nil {
			aliases = []string{}
		}
		resources = append(resources, resourceEntry{
			Name:     r.Plural,
			Aliases:  aliases,
			Table:    r.Table,
			Category: r.Category,
		})
	}

	format := outputFmt
	if format == "json" || format == "" {
		out := map[string]interface{}{
			"commands":  catalog,
			"resources": resources,
			"notes":     "Any unrecognized resource name is treated as a raw ServiceNow table name.",
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	// Plain text fallback
	fmt.Println("Commands:")
	for _, c := range catalog {
		fmt.Printf("  %-12s %s\n", c.Name, c.Description)
	}
	fmt.Println("\nResources:")
	for _, r := range resources {
		fmt.Printf("  %-18s -> %s\n", r.Name, r.Table)
	}
	fmt.Println("\nAny unrecognized resource name is treated as a raw ServiceNow table name.")
	return nil
}
