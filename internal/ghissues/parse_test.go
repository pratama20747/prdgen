package ghissues

import "testing"

func TestParse_ValidJSON(t *testing.T) {
	raw := `[{"title":"Setup DB","body":"do migration","labels":["phase-1"],"phase":"Fase 1"}]`
	issues, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Title != "Setup DB" {
		t.Errorf("unexpected title: %s", issues[0].Title)
	}
}

func TestParse_StripsMarkdownCodeFence(t *testing.T) {
	raw := "```json\n[{\"title\":\"Setup DB\",\"body\":\"x\",\"labels\":[],\"phase\":\"Fase 1\"}]\n```"
	issues, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 || issues[0].Title != "Setup DB" {
		t.Errorf("unexpected result: %+v", issues)
	}
}

func TestParse_StripsPlainCodeFenceNoLangTag(t *testing.T) {
	raw := "```\n[{\"title\":\"X\",\"body\":\"y\",\"labels\":[],\"phase\":\"Fase 1\"}]\n```"
	issues, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("unexpected result: %+v", issues)
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := Parse("bukan json sama sekali")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParse_EmptyArray(t *testing.T) {
	_, err := Parse("[]")
	if err == nil {
		t.Fatal("expected error for empty issue list")
	}
}

func TestParse_MissingTitle(t *testing.T) {
	raw := `[{"title":"","body":"x","labels":[],"phase":"Fase 1"}]`
	_, err := Parse(raw)
	if err == nil {
		t.Fatal("expected error for missing title")
	}
}

func TestParseSingle_ValidObject(t *testing.T) {
	raw := `{"title":"Revised title","body":"x","labels":["phase-1"],"phase":"Fase 1"}`
	issue, err := ParseSingle(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue.Title != "Revised title" {
		t.Errorf("unexpected title: %s", issue.Title)
	}
}

func TestParseSingle_StripsCodeFence(t *testing.T) {
	raw := "```json\n{\"title\":\"X\",\"body\":\"y\",\"labels\":[],\"phase\":\"Fase 1\"}\n```"
	issue, err := ParseSingle(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue.Title != "X" {
		t.Errorf("unexpected title: %s", issue.Title)
	}
}

func TestParseSingle_RejectsArray(t *testing.T) {
	raw := `[{"title":"X","body":"y","labels":[],"phase":"Fase 1"}]`
	_, err := ParseSingle(raw)
	if err == nil {
		t.Fatal("expected error when given an array instead of single object")
	}
}

func TestParseSingle_MissingTitle(t *testing.T) {
	raw := `{"title":"","body":"y","labels":[],"phase":"Fase 1"}`
	_, err := ParseSingle(raw)
	if err == nil {
		t.Fatal("expected error for missing title")
	}
}
