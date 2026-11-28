package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnnotateAddsDefaultsToTopLevelAndNestedAttributes(t *testing.T) {
	page := filepath.Join(t.TempDir(), "example.md")
	content := strings.Join([]string{
		"### Optional",
		"",
		"- `enabled` (Boolean) Whether it is enabled.",
		"- `timeout` (Number) The timeout; defaults to 60.",
		"- `name` (String) The name.",
		"",
		"<a id=\"nestedatt--policy\"></a>",
		"### Nested Schema for `policy`",
		"",
		"- `enabled` (Boolean) Whether the policy is on.",
		"",
	}, "\n")
	if err := os.WriteFile(page, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	found := map[string]string{"enabled": "true", "timeout": "60", "policy.enabled": "false"}

	added, missing, err := annotate(page, found)
	if err != nil {
		t.Fatal(err)
	}
	if added != 2 || len(missing) != 0 {
		t.Fatalf("added %d, missing %v; want 2 added and none missing", added, missing)
	}
	got, _ := os.ReadFile(page)
	for _, want := range []string{
		"- `enabled` (Boolean) Whether it is enabled. Defaults to `true`.",
		"- `timeout` (Number) The timeout; defaults to 60.",
		"- `name` (String) The name.",
		"- `enabled` (Boolean) Whether the policy is on. Defaults to `false`.",
	} {
		if !strings.Contains(string(got), want+"\n") {
			t.Errorf("page lacks %q:\n%s", want, got)
		}
	}

	// a second run adds nothing
	if added, _, _ := annotate(page, found); added != 0 {
		t.Errorf("second run added %d", added)
	}
}

func TestAnnotateReportsADefaultWithNoLine(t *testing.T) {
	page := filepath.Join(t.TempDir(), "example.md")
	if err := os.WriteFile(page, []byte("- `name` (String) The name.\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, missing, err := annotate(page, map[string]string{"gone": "1"})
	if err != nil || len(missing) != 1 || missing[0] != "gone" {
		t.Fatalf("missing %v, err %v; want [gone]", missing, err)
	}
}
