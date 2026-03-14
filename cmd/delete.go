package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mjamalu/snowctl/internal/client"
	"github.com/mjamalu/snowctl/internal/registry"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <resource> <name-or-sysid>",
	Short: "Delete a resource from ServiceNow",
	Long: `Delete a record from a ServiceNow table.

Prompts for confirmation unless --yes is specified.

Examples:
  snowctl delete incident INC0099999
  snowctl delete incident INC0099999 --yes
  snowctl delete user john.smith`,
	Args: cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return resourceNames(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: runDelete,
}

var deleteYes bool

func init() {
	deleteCmd.Flags().BoolVarP(&deleteYes, "yes", "y", false, "skip confirmation prompt")
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	res := resolveResource(args[0])
	identifier := args[1]

	c, err := getClient()
	if err != nil {
		return err
	}

	// Resolve to sys_id
	sysID, displayName, err := resolveToSysID(c, res, identifier)
	if err != nil {
		return err
	}

	// Confirm
	if !deleteYes {
		fmt.Fprintf(os.Stderr, "Delete %s %s (%s)? [y/N] ", res.Singular, displayName, sysID)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Fprintln(os.Stderr, "Cancelled.")
			return nil
		}
	}

	if err := c.Delete(res.Table, sysID); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%s %s deleted.\n", res.Singular, displayName)
	return nil
}

func resolveToSysID(c *client.SNClient, res *registry.Resource, identifier string) (sysID string, displayName string, err error) {
	if isSysID(identifier) {
		return identifier, identifier, nil
	}

	idField := res.IdentifierField
	if idField == "" {
		idField = "name"
	}

	result, err := c.List(res.Table, client.ListOptions{
		Query:  fmt.Sprintf("%s=%s", idField, identifier),
		Fields: []string{"sys_id", idField},
		Limit:  1,
	})
	if err != nil {
		return "", "", err
	}
	if len(result.Records) == 0 {
		return "", "", fmt.Errorf("%s %q not found", res.Singular, identifier)
	}

	sid, _ := result.Records[0]["sys_id"].(string)
	name := identifier
	if v, ok := result.Records[0][idField].(string); ok {
		name = v
	}
	return sid, name, nil
}
