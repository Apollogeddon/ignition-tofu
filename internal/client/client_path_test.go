package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// awkwardNames are names a gateway accepts that are not safe in a URL path as they are.
var awkwardNames = map[string]string{
	"x?y":   "x%3Fy",
	"a#b":   "a%23b",
	"a/b":   "a%2Fb",
	"50%":   "50%25",
	"a b":   "a%20b",
	"plain": "plain",
}

// pathServer answers every request with an empty resource and records the
// escaped path and query it received.
func pathServer(t *testing.T, gotPath, gotQuery *string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*gotPath = r.URL.EscapedPath()
		*gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]any{"name": "n"})
	}))
	t.Cleanup(server.Close)
	return server
}

func TestClient_EscapesNamesInPaths(t *testing.T) {
	var gotPath, gotQuery string
	server := pathServer(t, &gotPath, &gotQuery)
	c, err := NewClient(server.URL, "test-token", false)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	ctx := context.Background()

	for name, escaped := range awkwardNames {
		calls := []struct {
			desc string
			call func() error
			want string
		}{
			{"GetResource", func() error {
				var dest ResourceResponse[map[string]any]
				return c.GetResource(ctx, "test-type", name, &dest)
			}, "/data/api/v1/resources/find/ignition/test-type/" + escaped},
			{"DeleteResource", func() error {
				return c.DeleteResource(ctx, "test-type", name, "sig/1+")
			}, "/data/api/v1/resources/ignition/test-type/" + escaped + "/sig%2F1+"},
			{"GetProject", func() error {
				_, err := c.GetProject(ctx, name)
				return err
			}, "/data/api/v1/projects/find/" + escaped},
			{"DeleteProject", func() error {
				return c.DeleteProject(ctx, name)
			}, "/data/api/v1/projects/" + escaped},
		}
		for _, tc := range calls {
			gotPath, gotQuery = "", ""
			if err := tc.call(); err != nil {
				t.Fatalf("%s(%q): unexpected error: %v", tc.desc, name, err)
			}
			if gotPath != tc.want {
				t.Errorf("%s(%q): expected path %s, got %s", tc.desc, name, tc.want, gotPath)
			}
			if gotQuery != "" {
				t.Errorf("%s(%q): expected no query, got %q", tc.desc, name, gotQuery)
			}
		}
	}
}

func TestClient_UpdateProjectEscapesName(t *testing.T) {
	var gotPath, gotQuery string
	var putPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.EscapedPath(), r.URL.RawQuery
		if r.Method == http.MethodPut {
			putPath = gotPath
		}
		_ = json.NewEncoder(w).Encode(Project{Name: "x?y"})
	}))
	defer server.Close()
	c, err := NewClient(server.URL, "test-token", false)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if _, err := c.UpdateProject(context.Background(), Project{Name: "x?y"}); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if putPath != "/data/api/v1/projects/x%3Fy" {
		t.Errorf("Expected PUT path /data/api/v1/projects/x%%3Fy, got %s", putPath)
	}
	if gotQuery != "" {
		t.Errorf("Expected no query, got %q", gotQuery)
	}
}
