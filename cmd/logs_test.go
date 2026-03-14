package cmd

import (
	"strings"
	"testing"
)

func TestBuildAuditQuery_Basic(t *testing.T) {
	query := buildAuditQuery("incident", "abc123def456", "")
	expected := "tablename=incident^documentkey=abc123def456^ORDERBYDESCsys_created_on"
	if query != expected {
		t.Errorf("expected %q, got %q", expected, query)
	}
}

func TestBuildAuditQuery_WithFieldFilter(t *testing.T) {
	query := buildAuditQuery("incident", "abc123def456", "state")
	if !strings.Contains(query, "fieldname=state") {
		t.Errorf("expected query to contain fieldname=state, got %q", query)
	}
	if !strings.Contains(query, "tablename=incident") {
		t.Errorf("expected query to contain tablename=incident, got %q", query)
	}
	if !strings.Contains(query, "documentkey=abc123def456") {
		t.Errorf("expected query to contain documentkey=abc123def456, got %q", query)
	}
	if !strings.HasSuffix(query, "ORDERBYDESCsys_created_on") {
		t.Errorf("expected query to end with ORDERBYDESCsys_created_on, got %q", query)
	}
}

func TestBuildAuditQuery_ChangeRequest(t *testing.T) {
	query := buildAuditQuery("change_request", "sys123", "approval")
	if !strings.Contains(query, "tablename=change_request") {
		t.Errorf("expected change_request table in query, got %q", query)
	}
}

func TestReverseRecords(t *testing.T) {
	records := []map[string]interface{}{
		{"sys_created_on": "2026-03-14 10:00:00"},
		{"sys_created_on": "2026-03-14 09:00:00"},
		{"sys_created_on": "2026-03-14 08:00:00"},
	}
	reverseRecords(records)

	// After reverse, oldest should be first
	if records[0]["sys_created_on"] != "2026-03-14 08:00:00" {
		t.Errorf("expected oldest first after reverse, got %v", records[0]["sys_created_on"])
	}
	if records[2]["sys_created_on"] != "2026-03-14 10:00:00" {
		t.Errorf("expected newest last after reverse, got %v", records[2]["sys_created_on"])
	}
}

func TestReverseRecords_Empty(t *testing.T) {
	var records []map[string]interface{}
	reverseRecords(records) // should not panic
}

func TestReverseRecords_Single(t *testing.T) {
	records := []map[string]interface{}{{"a": "1"}}
	reverseRecords(records)
	if records[0]["a"] != "1" {
		t.Error("single element should remain unchanged")
	}
}

func TestAuditResource_Columns(t *testing.T) {
	if len(auditResource.DefaultColumns) != 5 {
		t.Errorf("expected 5 audit columns, got %d", len(auditResource.DefaultColumns))
	}
	expectedFields := []string{"sys_created_on", "user", "fieldname", "oldvalue", "newvalue"}
	for i, col := range auditResource.DefaultColumns {
		if col.Field != expectedFields[i] {
			t.Errorf("column %d: expected field %q, got %q", i, expectedFields[i], col.Field)
		}
	}
}
