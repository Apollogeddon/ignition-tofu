package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_UnparsableError(t *testing.T) {
	// Mock Server returning 500 with non-JSON body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Critical System Failure"))
	}))
	defer server.Close()

	c, err := NewClient(server.URL, "test-token", false)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	// Disable retries for this test to avoid timeout and check 500 handling
	c.HTTPClient.RetryMax = 0

	var dest ResourceResponse[map[string]any]
	err = c.GetResource(context.Background(), "test-type", "test-name", &dest)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Critical System Failure") {
		t.Errorf("Expected error containing 'Critical System Failure', got '%v'", err)
	}
	if !strings.Contains(err.Error(), "status: 500") {
		t.Errorf("Expected error containing 'status: 500', got '%v'", err)
	}
}

func TestClient_EmptyBodyError(t *testing.T) {
	// Mock Server returning 200 OK but empty body (should be an error for get/create)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// No body written
	}))
	defer server.Close()

	c, err := NewClient(server.URL, "test-token", false)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	var dest ResourceResponse[map[string]any]
	err = c.GetResource(context.Background(), "test-type", "test-name", &dest)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "empty response body") && !strings.Contains(err.Error(), "unexpected end of JSON input") {
		// Note: io.ReadAll on empty body returns empty slice, unmarshalResourceResponse checks this
		t.Errorf("Expected 'empty response body' or 'unexpected end of JSON input', got '%v'", err)
	}
}

func TestClient_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"incomplete": `))
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	var dest map[string]any
	err := c.GetResource(context.Background(), "t", "n", &dest)
	if err == nil {
		t.Fatal("Expected error for malformed JSON")
	}
}

func TestClient_CreateResource_EmptyResponse(t *testing.T) {
	// CreateResource calls unmarshalResourceResponse which explicitly checks for empty body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	var dest map[string]any
	err := c.CreateResource(context.Background(), "t", map[string]any{}, &dest)

	if err == nil {
		t.Fatal("Expected error for empty response on Create")
	}
	if err.Error() != "empty response body" {
		t.Errorf("Expected 'empty response body', got '%v'", err)
	}
}

func TestClient_CreateResource_EmptyArray(t *testing.T) {
	// API returns [] which is valid JSON but invalid for resource creation response (expects [item])
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	var dest map[string]any
	err := c.CreateResource(context.Background(), "t", map[string]any{}, &dest)

	if err == nil {
		t.Fatal("Expected error for empty array response")
	}
	// json.Unmarshal will succeed into rawItems, but len(rawItems) == 0
	if !strings.Contains(err.Error(), "failed to unmarshal response") || strings.Contains(err.Error(), "%!") {
		t.Errorf("Expected unmarshal error, got '%v'", err)
	}
}

func TestClient_CreateResource_ResourceChanges(t *testing.T) {
	// Test the path where API returns { "success": true, "changes": [...] }
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Return changes object
			_, _ = w.Write([]byte(`{
				"success": true,
				"changes": [
					{
						"name": "new-resource",
						"type": "test",
						"collection": "tests",
						"newSignature": "sig-789"
					}
				]
			}`))
			return
		}

		// Then it calls GetResourceWithModule to fetch the full object
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(ResourceResponse[map[string]any]{
				Name:      "new-resource",
				Signature: "sig-789",
			})
		}
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	var dest ResourceResponse[map[string]any]
	err := c.CreateResource(context.Background(), "test", map[string]any{}, &dest)
	if err != nil {
		t.Fatalf("Unexpected error handling ResourceChangesResponse: %v", err)
	}
	if dest.Name != "new-resource" {
		t.Errorf("Expected fetched name 'new-resource', got '%s'", dest.Name)
	}
}

func TestClient_APIError_FieldMessages(t *testing.T) {
	// Test complex error with field messages
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"success": false,
			"messages": ["Top level error"],
			"fieldMessages": [
				{
					"fieldName": "config.port",
					"messages": ["Must be integer"]
				}
			]
		}`))
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	var dest map[string]any
	err := c.GetResource(context.Background(), "t", "n", &dest)

	if err == nil {
		t.Fatal("Expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "Top level error") {
		t.Errorf("Error missing top level message: %s", msg)
	}
	if !strings.Contains(msg, "config.port") {
		t.Errorf("Error missing field name: %s", msg)
	}
	if !strings.Contains(msg, "Must be integer") {
		t.Errorf("Error missing field message: %s", msg)
	}
}

func TestClient_NotFoundIsErrNotFound(t *testing.T) {
	for name, body := range map[string]string{
		"plain":   "no such resource",
		"apiJSON": `{"success":false,"problem":{"message":"Resource not found"}}`,
	} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(body))
			}))
			defer server.Close()

			c, _ := NewClient(server.URL, "token", false)
			_, err := c.GetSMTPProfile(context.Background(), "missing")
			if !errors.Is(err, ErrNotFound) {
				t.Fatalf("expected ErrNotFound, got %v", err)
			}
			if name == "apiJSON" {
				var apiErr *APIErrorResponse
				if !errors.As(err, &apiErr) || err.Error() != "API error: Resource not found" {
					t.Errorf("expected the gateway's APIErrorResponse to be kept, got %v", err)
				}
			} else if !strings.Contains(err.Error(), "status: 404") {
				t.Errorf("expected the status in the message, got %v", err)
			}
		})
	}
}

func TestClient_OtherStatusIsNotErrNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	_, err := c.GetSMTPProfile(context.Background(), "x")
	if err == nil || errors.Is(err, ErrNotFound) {
		t.Fatalf("expected a non-ErrNotFound error, got %v", err)
	}
}

func TestClient_OversizedResponseRefused(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"name": "` + strings.Repeat("x", 100) + `"}`))
	}))
	defer server.Close()

	defer func(n int) { maxResponseBytes = n }(maxResponseBytes)
	maxResponseBytes = 64

	c, _ := NewClient(server.URL, "token", false)
	_, err := c.GetProject(context.Background(), "p")
	if err == nil || !strings.Contains(err.Error(), "over 64 bytes") {
		t.Errorf("Expected the size cap, got %v", err)
	}
}
