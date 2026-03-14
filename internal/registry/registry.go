package registry

import "strings"

// Registry maps user-facing resource names to ServiceNow table definitions.
type Registry struct {
	resources []*Resource
	lookup    map[string]*Resource
}

// New creates a registry populated with all built-in resources.
func New() *Registry {
	r := &Registry{
		lookup: make(map[string]*Resource),
	}
	for i := range builtinResources {
		res := &builtinResources[i]
		r.resources = append(r.resources, res)
		r.lookup[res.Plural] = res
		r.lookup[res.Singular] = res
		r.lookup[res.Table] = res
		for _, alias := range res.Aliases {
			r.lookup[alias] = res
		}
	}
	return r
}

// Resolve looks up a resource by plural, singular, alias, or raw table name.
// Unrecognized names are treated as raw ServiceNow table names.
func (r *Registry) Resolve(name string) *Resource {
	name = strings.ToLower(name)
	if res, ok := r.lookup[name]; ok {
		return res
	}
	// Ad-hoc resource for unknown table names
	return &Resource{
		Plural:   name,
		Singular: name,
		Table:    name,
	}
}

// List returns all built-in resources.
func (r *Registry) List() []*Resource {
	return r.resources
}

// ListByCategory returns resources filtered by category.
func (r *Registry) ListByCategory(category string) []*Resource {
	var result []*Resource
	for _, res := range r.resources {
		if res.Category == category {
			result = append(result, res)
		}
	}
	return result
}
