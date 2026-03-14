package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mjamalu/snowctl/internal/registry"
)

var testResource = &registry.Resource{
	Plural:   "incidents",
	Singular: "incident",
	Table:    "incident",
	DefaultColumns: []registry.Column{
		{Header: "NUMBER", Field: "number"},
		{Header: "STATE", Field: "state"},
		{Header: "SHORT DESCRIPTION", Field: "short_description"},
	},
}

var testRecords = []map[string]interface{}{
	{"number": "INC0001", "state": "New", "short_description": "Test incident 1", "sys_id": "abc123"},
	{"number": "INC0002", "state": "In Progress", "short_description": "Test incident 2", "sys_id": "def456"},
}

func TestTableFormatter_FormatList(t *testing.T) {
	var buf bytes.Buffer
	f := &TableFormatter{}
	err := f.FormatList(&buf, testRecords, testResource)
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "NUMBER") {
		t.Error("expected header 'NUMBER' in output")
	}
	if !strings.Contains(out, "INC0001") {
		t.Error("expected 'INC0001' in output")
	}
	if !strings.Contains(out, "INC0002") {
		t.Error("expected 'INC0002' in output")
	}
}

func TestTableFormatter_EmptyList(t *testing.T) {
	var buf bytes.Buffer
	f := &TableFormatter{}
	err := f.FormatList(&buf, nil, testResource)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No resources found") {
		t.Error("expected 'No resources found' message")
	}
}

func TestTableFormatter_FormatSingle(t *testing.T) {
	var buf bytes.Buffer
	f := &TableFormatter{}
	err := f.FormatSingle(&buf, testRecords[0], testResource)
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "INC0001") {
		t.Error("expected 'INC0001' in output")
	}
}

func TestJSONFormatter_FormatList(t *testing.T) {
	var buf bytes.Buffer
	f := &JSONFormatter{}
	err := f.FormatList(&buf, testRecords, testResource)
	if err != nil {
		t.Fatal(err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 records, got %d", len(result))
	}
}

func TestJSONFormatter_FormatSingle(t *testing.T) {
	var buf bytes.Buffer
	f := &JSONFormatter{}
	err := f.FormatSingle(&buf, testRecords[0], testResource)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if result["number"] != "INC0001" {
		t.Errorf("expected number 'INC0001', got %v", result["number"])
	}
}

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		format string
		expect string
	}{
		{"json", "*output.JSONFormatter"},
		{"yaml", "*output.YAMLFormatter"},
		{"table", "*output.TableFormatter"},
		{"", "*output.TableFormatter"},
	}
	for _, tt := range tests {
		f := NewFormatter(tt.format)
		if f == nil {
			t.Errorf("NewFormatter(%q) returned nil", tt.format)
		}
	}
}

func TestTableFormatter_AdHocResource(t *testing.T) {
	adHoc := &registry.Resource{
		Plural: "sys_audit",
		Table:  "sys_audit",
	}
	records := []map[string]interface{}{
		{"sys_id": "abc", "tablename": "incident", "fieldname": "state"},
	}
	var buf bytes.Buffer
	f := &TableFormatter{}
	err := f.FormatList(&buf, records, adHoc)
	if err != nil {
		t.Fatal(err)
	}
	// Should infer columns from record keys
	if !strings.Contains(buf.String(), "TABLENAME") || !strings.Contains(buf.String(), "FIELDNAME") {
		t.Error("expected inferred column headers for ad-hoc resource")
	}
}
