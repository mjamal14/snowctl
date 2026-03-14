package cmd

import (
	"fmt"
	"os"

	"github.com/mjamalu/snowctl/internal/client"
	"github.com/mjamalu/snowctl/internal/config"
	"github.com/mjamalu/snowctl/internal/output"
	"github.com/mjamalu/snowctl/internal/registry"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	contextName string
	outputFmt   string
	debug       bool
	agentMode   bool

	cfg *config.Config
	reg *registry.Registry
)

var rootCmd = &cobra.Command{
	Use:   "snowctl",
	Short: "kubectl-style CLI for ServiceNow",
	Long: `snowctl is a command-line tool for managing ServiceNow resources.
It uses familiar verb-noun syntax inspired by kubectl.

  snowctl get incidents --query "priority=1"
  snowctl describe incident INC0012345
  snowctl create incident --short-description "Outage"
  snowctl config use-context prod`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/snowctl/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&contextName, "context", "", "override the current-context")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "", "output format: table, json, yaml")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug output")
	rootCmd.PersistentFlags().BoolVar(&agentMode, "agent", false, "structured JSON output for AI agents")
}

func initConfig() {
	reg = registry.New()

	path := cfgFile
	if path == "" {
		path = config.DefaultConfigPath()
	}

	var err error
	cfg, err = config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load config: %s\n", err)
		cfg = &config.Config{APIVersion: "v1"}
	}

	if contextName != "" {
		cfg.CurrentContext = contextName
	}
}

// getClient creates an API client from the active context.
func getClient() (*client.SNClient, error) {
	ctx, err := cfg.ActiveContext()
	if err != nil {
		return nil, err
	}

	var auth client.Authenticator
	switch ctx.Auth.Type {
	case "basic", "":
		username := os.Getenv("SNOWCTL_USERNAME")
		if username == "" {
			username = ctx.Auth.Username
		}
		password := os.Getenv("SNOWCTL_PASSWORD")
		if password == "" {
			password = ctx.Auth.Password
		}
		if username == "" {
			return nil, fmt.Errorf("no username configured; set SNOWCTL_USERNAME or add username to config")
		}
		if password == "" {
			return nil, fmt.Errorf("no password configured; set SNOWCTL_PASSWORD or add password to config")
		}
		auth = &client.BasicAuth{Username: username, Password: password}
	default:
		return nil, fmt.Errorf("unsupported auth type: %s (OAuth support coming soon)", ctx.Auth.Type)
	}

	c := client.New(ctx.Instance, auth)
	c.SetDebug(debug)
	return c, nil
}

// getFormatter returns the output formatter based on flags and config.
func getFormatter() output.Formatter {
	format := outputFmt
	if format == "" {
		ctx, err := cfg.ActiveContext()
		if err == nil {
			format = cfg.EffectiveOutput(ctx)
		} else {
			format = "table"
		}
	}
	return output.NewFormatter(format)
}

// resolveResource resolves a resource name via the registry.
func resolveResource(name string) *registry.Resource {
	return reg.Resolve(name)
}

// resourceNames returns all valid resource names for shell completion.
func resourceNames() []string {
	var names []string
	for _, r := range reg.List() {
		names = append(names, r.Plural)
		names = append(names, r.Singular)
		for _, a := range r.Aliases {
			names = append(names, a)
		}
	}
	return names
}
