package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/mjamalu/snowctl/internal/registry"
)

// JSONFormatter outputs records as JSON.
type JSONFormatter struct{}

func (f *JSONFormatter) FormatList(w io.Writer, records []map[string]interface{}, res *registry.Resource) error {
	return writeJSON(w, records)
}

func (f *JSONFormatter) FormatSingle(w io.Writer, record map[string]interface{}, res *registry.Resource) error {
	return writeJSON(w, record)
}

func writeJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
}
