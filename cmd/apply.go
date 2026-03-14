package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mjamalu/snowctl/internal/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var applyCmd = &cobra.Command{
	Use:   "apply -f <file>",
	Short: "Create or update resources from a YAML manifest",
	Long: `Apply a declarative configuration to ServiceNow.

If the manifest includes an identifier (sys_id or number), the record is updated.
Otherwise, a new record is created.

Supports multi-document YAML files (separated by ---) and directories.

Examples:
  snowctl apply -f incident.yaml
  snowctl apply -f manifests/
  snowctl apply -f incident.yaml --dry-run`,
	RunE: runApply,
}

var (
	applyFile   string
	applyDryRun bool
)

func init() {
	applyCmd.Flags().StringVarP(&applyFile, "file", "f", "", "path to YAML manifest file or directory (required)")
	applyCmd.Flags().BoolVar(&applyDryRun, "dry-run", false, "print actions without executing")
	applyCmd.MarkFlagRequired("file")

	rootCmd.AddCommand(applyCmd)
}

// Manifest represents a single resource definition in a YAML file.
type Manifest struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   map[string]interface{} `yaml:"metadata,omitempty"`
	Spec       map[string]interface{} `yaml:"spec"`
}

func runApply(cmd *cobra.Command, args []string) error {
	c, err := getClient()
	if err != nil {
		return err
	}

	info, err := os.Stat(applyFile)
	if err != nil {
		return fmt.Errorf("cannot access %s: %w", applyFile, err)
	}

	var files []string
	if info.IsDir() {
		entries, err := os.ReadDir(applyFile)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if !e.IsDir() && (strings.HasSuffix(e.Name(), ".yaml") || strings.HasSuffix(e.Name(), ".yml")) {
				files = append(files, filepath.Join(applyFile, e.Name()))
			}
		}
		if len(files) == 0 {
			return fmt.Errorf("no YAML files found in %s", applyFile)
		}
	} else {
		files = []string{applyFile}
	}

	for _, f := range files {
		manifests, err := parseManifestFile(f)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", f, err)
		}
		for _, m := range manifests {
			if err := applyManifest(c, m); err != nil {
				return fmt.Errorf("applying %s/%s: %w", f, m.Kind, err)
			}
		}
	}

	return nil
}

func parseManifestFile(path string) ([]Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var manifests []Manifest
	dec := yaml.NewDecoder(f)
	for {
		var m Manifest
		if err := dec.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if m.Kind == "" {
			return nil, fmt.Errorf("manifest missing 'kind' field")
		}
		manifests = append(manifests, m)
	}
	return manifests, nil
}

func applyManifest(c *client.SNClient, m Manifest) error {
	// Resolve kind to resource
	res := resolveResource(strings.ToLower(m.Kind))

	// Check if this is an update (has identifier in metadata)
	var sysID string
	var identifier string
	if m.Metadata != nil {
		if sid, ok := m.Metadata["sys_id"].(string); ok {
			sysID = sid
		}
		if res.IdentifierField != "" {
			if id, ok := m.Metadata[res.IdentifierField].(string); ok {
				identifier = id
			}
		}
	}

	// If we have an identifier but no sys_id, look it up
	if sysID == "" && identifier != "" {
		result, err := c.List(res.Table, client.ListOptions{
			Query:  fmt.Sprintf("%s=%s", res.IdentifierField, identifier),
			Fields: []string{"sys_id"},
			Limit:  1,
		})
		if err != nil {
			return err
		}
		if len(result.Records) > 0 {
			sysID, _ = result.Records[0]["sys_id"].(string)
		}
	}

	if applyDryRun {
		if sysID != "" {
			fmt.Printf("[dry-run] would update %s %s (sys_id: %s)\n", res.Singular, identifier, sysID)
		} else {
			fmt.Printf("[dry-run] would create %s\n", res.Singular)
		}
		return nil
	}

	if sysID != "" {
		// Update
		_, err := c.Patch(res.Table, sysID, m.Spec)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "%s %s updated.\n", res.Singular, identifier)
	} else {
		// Create
		record, err := c.Create(res.Table, m.Spec)
		if err != nil {
			return err
		}
		id := ""
		if res.IdentifierField != "" {
			if v, ok := record[res.IdentifierField]; ok {
				id = fmt.Sprintf(" %v", v)
			}
		}
		fmt.Fprintf(os.Stderr, "%s%s created.\n", res.Singular, id)
	}

	return nil
}
