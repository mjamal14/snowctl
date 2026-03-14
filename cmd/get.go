package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mjamalu/snowctl/internal/client"
	"github.com/mjamalu/snowctl/internal/output"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <resource> [flags]",
	Short: "List resources from ServiceNow",
	Long: `List resources from a ServiceNow table.

Examples:
  snowctl get incidents
  snowctl get incidents --query "priority=1^active=true" --limit 10
  snowctl get inc -q "state=1" -o json
  snowctl get users --fields "user_name,email,name"
  snowctl get sys_audit --limit 5`,
	Args: cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return resourceNames(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: runGet,
}

var (
	getQuery  string
	getFields string
	getLimit  int
	getOffset int
)

func init() {
	getCmd.Flags().StringVarP(&getQuery, "query", "q", "", "ServiceNow encoded query (sysparm_query)")
	getCmd.Flags().StringVar(&getFields, "fields", "", "comma-separated list of fields to return")
	getCmd.Flags().IntVarP(&getLimit, "limit", "l", 0, "maximum number of records (default from config)")
	getCmd.Flags().IntVar(&getOffset, "offset", 0, "number of records to skip")

	rootCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	res := resolveResource(args[0])

	c, err := getClient()
	if err != nil {
		return err
	}

	limit := getLimit
	if limit == 0 {
		ctx, _ := cfg.ActiveContext()
		limit = cfg.EffectiveLimit(ctx)
	}

	opts := client.ListOptions{
		Query:        getQuery,
		Limit:        limit,
		Offset:       getOffset,
		DisplayValue: "true",
	}

	if getFields != "" {
		opts.Fields = strings.Split(getFields, ",")
	}

	result, err := c.List(res.Table, opts)
	if err != nil {
		if agentMode {
			ctx, _ := cfg.ActiveContext()
			instance := ""
			if ctx != nil {
				instance = ctx.Instance
			}
			return output.FormatAgentError(os.Stdout, "API_ERROR", err.Error(), nil,
				&output.AgentContext{Verb: "get", Resource: res.Plural, Instance: instance})
		}
		return err
	}

	if agentMode {
		ctx, _ := cfg.ActiveContext()
		instance := ""
		if ctx != nil {
			instance = ctx.Instance
		}
		f := &output.AgentFormatter{
			Verb:     "get",
			Resource: res.Plural,
			Instance: instance,
			Total:    result.TotalCount,
			HasMore:  result.HasMore,
			Offset:   getOffset,
			Limit:    limit,
		}
		return f.FormatList(os.Stdout, result.Records, res)
	}

	formatter := getFormatter()
	if err := formatter.FormatList(os.Stdout, result.Records, res); err != nil {
		return err
	}

	// Show record count hint for table output
	if outputFmt == "" || outputFmt == "table" {
		if result.HasMore {
			fmt.Fprintf(os.Stderr, "\nShowing %d of %d records. Use --limit or --offset to see more.\n",
				len(result.Records), result.TotalCount)
		}
	}

	return nil
}
