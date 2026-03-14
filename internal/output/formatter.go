package output

import (
	"io"

	"github.com/mjamalu/snowctl/internal/registry"
)

// Formatter writes records in a specific format.
type Formatter interface {
	FormatList(w io.Writer, records []map[string]interface{}, res *registry.Resource) error
	FormatSingle(w io.Writer, record map[string]interface{}, res *registry.Resource) error
}

// NewFormatter creates a formatter for the given output format.
func NewFormatter(format string) Formatter {
	switch format {
	case "json":
		return &JSONFormatter{}
	case "yaml":
		return &YAMLFormatter{}
	default:
		return &TableFormatter{}
	}
}
