package output

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/mjamalu/snowctl/internal/registry"
)

// TableFormatter outputs records as an ASCII table.
type TableFormatter struct{}

func (f *TableFormatter) FormatList(w io.Writer, records []map[string]interface{}, res *registry.Resource) error {
	if len(records) == 0 {
		fmt.Fprintln(w, "No resources found.")
		return nil
	}

	columns := res.DefaultColumns
	if len(columns) == 0 {
		columns = inferColumns(records)
	}

	// Calculate widths
	widths := make([]int, len(columns))
	for i, col := range columns {
		widths[i] = len(col.Header)
	}
	for _, rec := range records {
		for i, col := range columns {
			val := fieldToString(rec, col.Field)
			if len(val) > widths[i] {
				widths[i] = len(val)
			}
		}
	}

	// Cap widths
	for i := range widths {
		if widths[i] > 60 {
			widths[i] = 60
		}
	}

	// Print header
	parts := make([]string, len(columns))
	for i, col := range columns {
		parts[i] = padRight(col.Header, widths[i])
	}
	fmt.Fprintln(w, strings.Join(parts, "  "))

	// Print rows
	for _, rec := range records {
		for i, col := range columns {
			val := fieldToString(rec, col.Field)
			if len(val) > widths[i] {
				val = val[:widths[i]-3] + "..."
			}
			parts[i] = padRight(val, widths[i])
		}
		fmt.Fprintln(w, strings.Join(parts, "  "))
	}

	return nil
}

func (f *TableFormatter) FormatSingle(w io.Writer, record map[string]interface{}, res *registry.Resource) error {
	keys := sortedKeys(record)
	maxKeyLen := 0
	for _, k := range keys {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}

	for _, k := range keys {
		val := fieldToString(record, k)
		label := padRight(strings.ToUpper(k)+":", maxKeyLen+1)
		fmt.Fprintf(w, "%s  %s\n", label, val)
	}
	return nil
}

func inferColumns(records []map[string]interface{}) []registry.Column {
	if len(records) == 0 {
		return nil
	}
	keys := sortedKeys(records[0])
	cols := make([]registry.Column, 0, len(keys))
	for _, k := range keys {
		if k == "sys_id" || strings.HasPrefix(k, "sys_") && k != "sys_created_on" && k != "sys_updated_on" {
			continue
		}
		cols = append(cols, registry.Column{
			Header: strings.ToUpper(strings.ReplaceAll(k, "_", " ")),
			Field:  k,
		})
		if len(cols) >= 8 {
			break
		}
	}
	return cols
}

func fieldToString(record map[string]interface{}, field string) string {
	val, ok := record[field]
	if !ok {
		return "--"
	}
	switch v := val.(type) {
	case nil:
		return "--"
	case string:
		if v == "" {
			return "--"
		}
		return v
	case map[string]interface{}:
		// ServiceNow reference fields can be objects with display_value
		if dv, ok := v["display_value"]; ok {
			return fmt.Sprintf("%v", dv)
		}
		if dv, ok := v["value"]; ok {
			return fmt.Sprintf("%v", dv)
		}
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
