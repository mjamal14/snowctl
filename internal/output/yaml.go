package output

import (
	"fmt"
	"io"

	"github.com/mjamalu/snowctl/internal/registry"
	"gopkg.in/yaml.v3"
)

// YAMLFormatter outputs records as YAML.
type YAMLFormatter struct{}

func (f *YAMLFormatter) FormatList(w io.Writer, records []map[string]interface{}, res *registry.Resource) error {
	return writeYAML(w, records)
}

func (f *YAMLFormatter) FormatSingle(w io.Writer, record map[string]interface{}, res *registry.Resource) error {
	return writeYAML(w, record)
}

func writeYAML(w io.Writer, v interface{}) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encoding YAML: %w", err)
	}
	return enc.Close()
}
