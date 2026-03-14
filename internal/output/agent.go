package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/mjamalu/snowctl/internal/registry"
)

// AgentEnvelope is the structured JSON envelope for AI agent consumption.
type AgentEnvelope struct {
	OK      bool          `json:"ok"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *AgentError   `json:"error,omitempty"`
	Context *AgentContext `json:"context,omitempty"`
}

// AgentError is the error field in the agent envelope.
type AgentError struct {
	Code        string   `json:"code"`
	Message     string   `json:"message"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// AgentContext is metadata about the operation.
type AgentContext struct {
	Verb        string   `json:"verb"`
	Resource    string   `json:"resource"`
	Instance    string   `json:"instance,omitempty"`
	Total       int      `json:"total,omitempty"`
	HasMore     bool     `json:"has_more,omitempty"`
	Offset      int      `json:"offset,omitempty"`
	Limit       int      `json:"limit,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// AgentFormatter wraps output in the agent JSON envelope.
type AgentFormatter struct {
	Verb     string
	Resource string
	Instance string
	Total    int
	HasMore  bool
	Offset   int
	Limit    int
}

func (f *AgentFormatter) FormatList(w io.Writer, records []map[string]interface{}, res *registry.Resource) error {
	envelope := AgentEnvelope{
		OK:     true,
		Result: records,
		Context: &AgentContext{
			Verb:     f.Verb,
			Resource: res.Plural,
			Instance: f.Instance,
			Total:    f.Total,
			HasMore:  f.HasMore,
			Offset:   f.Offset,
			Limit:    f.Limit,
		},
	}

	if f.HasMore {
		envelope.Context.Suggestions = []string{
			fmt.Sprintf("Use '--offset %d' to see the next page", f.Offset+f.Limit),
		}
	}

	return writeJSONIndented(w, envelope)
}

func (f *AgentFormatter) FormatSingle(w io.Writer, record map[string]interface{}, res *registry.Resource) error {
	envelope := AgentEnvelope{
		OK:     true,
		Result: record,
		Context: &AgentContext{
			Verb:     f.Verb,
			Resource: res.Singular,
			Instance: f.Instance,
		},
	}
	return writeJSONIndented(w, envelope)
}

// FormatError creates an error envelope.
func FormatAgentError(w io.Writer, code, message string, suggestions []string, ctx *AgentContext) error {
	envelope := AgentEnvelope{
		OK: false,
		Error: &AgentError{
			Code:        code,
			Message:     message,
			Suggestions: suggestions,
		},
		Context: ctx,
	}
	return writeJSONIndented(w, envelope)
}

func writeJSONIndented(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
