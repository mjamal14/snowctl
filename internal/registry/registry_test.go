package registry

import "testing"

func TestResolve_Plural(t *testing.T) {
	reg := New()
	res := reg.Resolve("incidents")
	if res.Table != "incident" {
		t.Errorf("expected table 'incident', got %q", res.Table)
	}
}

func TestResolve_Singular(t *testing.T) {
	reg := New()
	res := reg.Resolve("incident")
	if res.Table != "incident" {
		t.Errorf("expected table 'incident', got %q", res.Table)
	}
}

func TestResolve_Alias(t *testing.T) {
	reg := New()
	res := reg.Resolve("inc")
	if res.Table != "incident" {
		t.Errorf("expected table 'incident', got %q", res.Table)
	}
}

func TestResolve_TableName(t *testing.T) {
	reg := New()
	res := reg.Resolve("change_request")
	if res.Table != "change_request" {
		t.Errorf("expected table 'change_request', got %q", res.Table)
	}
	if res.Plural != "changes" {
		t.Errorf("expected plural 'changes', got %q", res.Plural)
	}
}

func TestResolve_CaseInsensitive(t *testing.T) {
	reg := New()
	res := reg.Resolve("INCIDENTS")
	if res.Table != "incident" {
		t.Errorf("expected table 'incident', got %q", res.Table)
	}
}

func TestResolve_AdHoc(t *testing.T) {
	reg := New()
	res := reg.Resolve("sys_audit")
	if res.Table != "sys_audit" {
		t.Errorf("expected ad-hoc table 'sys_audit', got %q", res.Table)
	}
	if !res.IsAdHoc() {
		t.Error("expected ad-hoc resource")
	}
}

func TestResolve_AllAliases(t *testing.T) {
	tests := []struct {
		input string
		table string
	}{
		{"chg", "change_request"},
		{"prb", "problem"},
		{"req", "sc_request"},
		{"ritm", "sc_req_item"},
		{"usr", "sys_user"},
		{"grp", "sys_user_group"},
		{"srv", "cmdb_ci_server"},
		{"app", "cmdb_ci_appl"},
		{"svc", "cmdb_ci_service"},
		{"kb", "kb_knowledge"},
		{"catitem", "sc_cat_item"},
	}
	reg := New()
	for _, tt := range tests {
		res := reg.Resolve(tt.input)
		if res.Table != tt.table {
			t.Errorf("alias %q: expected table %q, got %q", tt.input, tt.table, res.Table)
		}
	}
}

func TestList(t *testing.T) {
	reg := New()
	resources := reg.List()
	if len(resources) != 14 {
		t.Errorf("expected 14 built-in resources, got %d", len(resources))
	}
}

func TestListByCategory(t *testing.T) {
	reg := New()

	itsm := reg.ListByCategory("ITSM")
	if len(itsm) != 6 {
		t.Errorf("expected 6 ITSM resources, got %d", len(itsm))
	}

	cmdb := reg.ListByCategory("CMDB")
	if len(cmdb) != 4 {
		t.Errorf("expected 4 CMDB resources, got %d", len(cmdb))
	}

	platform := reg.ListByCategory("Platform")
	if len(platform) != 4 {
		t.Errorf("expected 4 Platform resources, got %d", len(platform))
	}
}
