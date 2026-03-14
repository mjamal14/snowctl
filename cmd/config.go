package cmd

import (
	"fmt"
	"os"

	"github.com/mjamalu/snowctl/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage snowctl configuration",
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Display the current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
		return nil
	},
}

var configGetContextsCmd = &cobra.Command{
	Use:   "get-contexts",
	Short: "List all configured contexts",
	Run: func(cmd *cobra.Command, args []string) {
		if len(cfg.Contexts) == 0 {
			fmt.Println("No contexts configured. Use 'snowctl config set-context' to add one.")
			return
		}
		for _, ctx := range cfg.Contexts {
			marker := "  "
			if ctx.Name == cfg.CurrentContext {
				marker = "* "
			}
			fmt.Printf("%s%s\t%s\t(%s)\n", marker, ctx.Name, ctx.Instance, ctx.Auth.Type)
		}
	},
}

var configCurrentContextCmd = &cobra.Command{
	Use:   "current-context",
	Short: "Show the current context name",
	Run: func(cmd *cobra.Command, args []string) {
		if cfg.CurrentContext == "" {
			fmt.Println("No current context set.")
			return
		}
		fmt.Println(cfg.CurrentContext)
	},
}

var configUseContextCmd = &cobra.Command{
	Use:   "use-context <name>",
	Short: "Switch the active context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		found := false
		for _, ctx := range cfg.Contexts {
			if ctx.Name == name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("context %q not found; use 'snowctl config get-contexts' to list available contexts", name)
		}

		cfg.CurrentContext = name
		path := cfgFile
		if path == "" {
			path = config.DefaultConfigPath()
		}
		if err := config.Save(cfg, path); err != nil {
			return err
		}
		fmt.Printf("Switched to context %q.\n", name)
		return nil
	},
}

var (
	setCtxInstance string
	setCtxAuthType string
	setCtxUsername string
)

var configSetContextCmd = &cobra.Command{
	Use:   "set-context <name>",
	Short: "Create or update a context",
	Long: `Create or update a named context with connection details.

Examples:
  snowctl config set-context dev --instance https://dev12345.service-now.com --username admin
  snowctl config set-context prod --instance https://prod.service-now.com --auth-type oauth`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Find existing or create new
		var ctx *config.ContextConfig
		for i := range cfg.Contexts {
			if cfg.Contexts[i].Name == name {
				ctx = &cfg.Contexts[i]
				break
			}
		}
		if ctx == nil {
			cfg.Contexts = append(cfg.Contexts, config.ContextConfig{Name: name})
			ctx = &cfg.Contexts[len(cfg.Contexts)-1]
		}

		if setCtxInstance != "" {
			ctx.Instance = setCtxInstance
		}
		if setCtxAuthType != "" {
			ctx.Auth.Type = setCtxAuthType
		}
		if setCtxUsername != "" {
			ctx.Auth.Username = setCtxUsername
		}

		// Set as current if it's the only context
		if len(cfg.Contexts) == 1 || cfg.CurrentContext == "" {
			cfg.CurrentContext = name
		}

		path := cfgFile
		if path == "" {
			path = config.DefaultConfigPath()
		}
		if err := config.Save(cfg, path); err != nil {
			return err
		}

		fmt.Printf("Context %q configured.\n", name)
		if cfg.CurrentContext == name {
			fmt.Printf("Active context set to %q.\n", name)
		}
		return nil
	},
}

var configDeleteContextCmd = &cobra.Command{
	Use:   "delete-context <name>",
	Short: "Remove a context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		found := false
		for i, ctx := range cfg.Contexts {
			if ctx.Name == name {
				cfg.Contexts = append(cfg.Contexts[:i], cfg.Contexts[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("context %q not found", name)
		}

		if cfg.CurrentContext == name {
			cfg.CurrentContext = ""
			if len(cfg.Contexts) > 0 {
				cfg.CurrentContext = cfg.Contexts[0].Name
			}
		}

		path := cfgFile
		if path == "" {
			path = config.DefaultConfigPath()
		}
		if err := config.Save(cfg, path); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Context %q deleted.\n", name)
		return nil
	},
}

func init() {
	configSetContextCmd.Flags().StringVar(&setCtxInstance, "instance", "", "ServiceNow instance URL")
	configSetContextCmd.Flags().StringVar(&setCtxAuthType, "auth-type", "basic", "authentication type: basic, oauth")
	configSetContextCmd.Flags().StringVar(&setCtxUsername, "username", "", "username for basic auth")

	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configGetContextsCmd)
	configCmd.AddCommand(configCurrentContextCmd)
	configCmd.AddCommand(configUseContextCmd)
	configCmd.AddCommand(configSetContextCmd)
	configCmd.AddCommand(configDeleteContextCmd)

	rootCmd.AddCommand(configCmd)
}
