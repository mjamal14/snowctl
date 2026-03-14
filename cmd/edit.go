package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mjamalu/snowctl/internal/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var editCmd = &cobra.Command{
	Use:   "edit <resource> <name-or-sysid>",
	Short: "Edit a resource in your default editor",
	Long: `Fetch a record, open it in $EDITOR as YAML, and apply changes on save.

Examples:
  snowctl edit incident INC0012345
  snowctl edit change CHG0001234
  EDITOR=nano snowctl edit user john.smith`,
	Args: cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return resourceNames(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	res := resolveResource(args[0])
	identifier := args[1]

	c, err := getClient()
	if err != nil {
		return err
	}

	// Fetch the record
	record, err := resolveRecord(c, res, identifier, client.ListOptions{DisplayValue: "true"})
	if err != nil {
		return err
	}

	sysID, _ := record["sys_id"].(string)
	if sysID == "" {
		return fmt.Errorf("record has no sys_id")
	}

	// Remove read-only sys_ fields before editing
	editableRecord := filterEditable(record)

	// Write to temp file
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("snowctl-edit-*.yaml"))
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	enc := yaml.NewEncoder(tmpFile)
	enc.SetIndent(2)
	if err := enc.Encode(editableRecord); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing YAML: %w", err)
	}
	tmpFile.Close()

	// Get file info before edit
	infoBefore, _ := os.Stat(tmpPath)

	// Open in editor
	editor := getEditor()
	editorCmd := exec.Command(editor, tmpPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	// Check if file was modified
	infoAfter, _ := os.Stat(tmpPath)
	if infoBefore.ModTime().Equal(infoAfter.ModTime()) {
		fmt.Fprintln(os.Stderr, "Edit cancelled (no changes).")
		return nil
	}

	// Read back
	edited, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("reading edited file: %w", err)
	}

	var updatedData map[string]interface{}
	if err := yaml.Unmarshal(edited, &updatedData); err != nil {
		return fmt.Errorf("parsing edited YAML: %w", err)
	}

	// Apply changes
	result, err := c.Patch(res.Table, sysID, updatedData)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%s updated.\n", res.Singular)
	_ = result
	return nil
}

func getEditor() string {
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	if e := os.Getenv("VISUAL"); e != "" {
		return e
	}
	// Check config
	ctx, err := cfg.ActiveContext()
	if err == nil && ctx.Defaults.Editor != "" {
		return ctx.Defaults.Editor
	}
	if cfg.Defaults.Editor != "" {
		return cfg.Defaults.Editor
	}
	return "vi"
}

func filterEditable(record map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{})
	readOnly := map[string]bool{
		"sys_id":         true,
		"sys_created_on": true,
		"sys_created_by": true,
		"sys_updated_on": true,
		"sys_updated_by": true,
		"sys_mod_count":  true,
		"sys_tags":       true,
		"sys_class_name": true,
	}
	for k, v := range record {
		if !readOnly[k] {
			filtered[k] = v
		}
	}
	return filtered
}

// Ensure resolveRecord is accessible (it's defined in describe.go)
var _ = func() {} // prevents import cycle
