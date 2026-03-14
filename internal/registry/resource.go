package registry

// Resource defines a ServiceNow resource type mapping.
type Resource struct {
	Plural          string
	Singular        string
	Aliases         []string
	Table           string
	Description     string
	Category        string
	DefaultColumns  []Column
	FieldAliases    map[string]string
	IdentifierField string
}

// Column defines a display column for table output.
type Column struct {
	Header string
	Field  string
	Width  int
}

// IsAdHoc returns true if this resource was created as a fallback for an unknown table name.
func (r *Resource) IsAdHoc() bool {
	return r.Category == ""
}
