package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestClient_GetResource(t *testing.T) {
	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected method GET, got %s", r.Method)
		}
		if r.Header.Get("X-Ignition-API-Token") != "test-token" {
			t.Errorf("Expected token test-token, got %s", r.Header.Get("X-Ignition-API-Token"))
		}
		if r.URL.Path != "/data/api/v1/resources/find/ignition/test-type/test-name" {
			t.Errorf("Expected path /data/api/v1/resources/find/ignition/test-type/test-name, got %s", r.URL.Path)
		}

		response := ResourceResponse[map[string]any]{
			Name:    "test-name",
			Enabled: boolPtr(true),
			Config:  map[string]any{"key": "value"},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Client
	c, err := NewClient(server.URL, "test-token", false)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test
	var dest ResourceResponse[map[string]any]
	err = c.GetResource(context.Background(), "test-type", "test-name", &dest)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if dest.Name != "test-name" {
		t.Errorf("Expected name test-name, got %s", dest.Name)
	}
	if dest.Config["key"] != "value" {
		t.Errorf("Expected config key=value, got %v", dest.Config["key"])
	}
}

func TestClient_CreateResource(t *testing.T) {
	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected method POST, got %s", r.Method)
		}
		if r.URL.Path != "/data/api/v1/resources/ignition/test-type" {
			t.Errorf("Expected path /data/api/v1/resources/ignition/test-type, got %s", r.URL.Path)
		}

		// API returns array of created resources
		response := []ResourceResponse[map[string]any]{
			{
				Name:      "test-name",
				Enabled:   boolPtr(true),
				Signature: "sig-123",
				Config:    map[string]any{"key": "value"},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Client
	c, err := NewClient(server.URL, "test-token", false)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test
	item := ResourceResponse[map[string]any]{
		Name:    "test-name",
		Enabled: boolPtr(true),
		Config:  map[string]any{"key": "value"},
	}
	var dest ResourceResponse[map[string]any]
	err = c.CreateResource(context.Background(), "test-type", item, &dest)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if dest.Signature != "sig-123" {
		t.Errorf("Expected signature sig-123, got %s", dest.Signature)
	}
}

func TestClient_UpdateResource(t *testing.T) {
	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected method PUT, got %s", r.Method)
		}
		if r.URL.Path != "/data/api/v1/resources/ignition/test-type" {
			t.Errorf("Expected path /data/api/v1/resources/ignition/test-type, got %s", r.URL.Path)
		}

		response := []ResourceResponse[map[string]any]{
			{
				Name:      "test-name",
				Enabled:   boolPtr(true),
				Signature: "sig-456",
				Config:    map[string]any{"key": "updated"},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Client
	c, err := NewClient(server.URL, "test-token", false)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test
	item := ResourceResponse[map[string]any]{
		Name:    "test-name",
		Enabled: boolPtr(true),
		Config:  map[string]any{"key": "updated"},
	}
	var dest ResourceResponse[map[string]any]
	err = c.UpdateResource(context.Background(), "test-type", item, &dest)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if dest.Signature != "sig-456" {
		t.Errorf("Expected signature sig-456, got %s", dest.Signature)
	}
}

func TestClient_DeleteResource(t *testing.T) {
	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected method DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/data/api/v1/resources/ignition/test-type/test-name/sig-123" {
			t.Errorf("Expected path /data/api/v1/resources/ignition/test-type/test-name/sig-123, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Client
	c, err := NewClient(server.URL, "test-token", false)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test
	err = c.DeleteResource(context.Background(), "test-type", "test-name", "sig-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestClient_ProjectOperations(t *testing.T) {
	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path != "/data/api/v1/projects/find/test-project" {
				t.Errorf("Expected path /data/api/v1/projects/find/test-project, got %s", r.URL.Path)
			}
			p := Project{Name: "test-project", Enabled: boolPtr(true)}
			_ = json.NewEncoder(w).Encode(p)
		case http.MethodPost:
			if r.URL.Path != "/data/api/v1/projects" {
				t.Errorf("Expected path /data/api/v1/projects, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusCreated)
		case http.MethodPut:
			if r.URL.Path != "/data/api/v1/projects/test-project" {
				t.Errorf("Expected path /data/api/v1/projects/test-project, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			if r.URL.Path != "/data/api/v1/projects/test-project" {
				t.Errorf("Expected path /data/api/v1/projects/test-project, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Client
	c, err := NewClient(server.URL, "test-token", false)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test Create
	p := Project{Name: "test-project", Enabled: boolPtr(true)}
	created, err := c.CreateProject(ctx, p)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	if created.Name != "test-project" {
		t.Errorf("Expected name test-project, got %s", created.Name)
	}

	// Test Update
	updated, err := c.UpdateProject(ctx, p)
	if err != nil {
		t.Fatalf("UpdateProject failed: %v", err)
	}
	if updated.Name != "test-project" {
		t.Errorf("Expected name test-project, got %s", updated.Name)
	}

	// Test Delete
	err = c.DeleteProject(ctx, "test-project")
	if err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}
}

func TestClient_APIError(t *testing.T) {
	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		response := APIErrorResponse{
			Success: false,
			Problem: &struct {
				Message    string   `json:"message"`
				StackTrace []string `json:"stacktrace,omitempty"`
			}{
				Message: "Resource already exists",
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	// Client
	c, err := NewClient(server.URL, "test-token", false)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test
	var dest ResourceResponse[map[string]any]
	err = c.GetResource(context.Background(), "test-type", "test-name", &dest)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedMsg := "API error: Resource already exists"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestClient_DatabaseConnectionOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := []ResourceResponse[DatabaseConfig]{
			{
				Name:      "test-db",
				Signature: "sig-db",
				Config:    DatabaseConfig{Driver: "MariaDB"},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	db := ResourceResponse[DatabaseConfig]{Name: "test-db", Config: DatabaseConfig{Driver: "MariaDB"}}

	_, err := c.CreateDatabaseConnection(context.Background(), db)
	if err != nil {
		t.Errorf("CreateDatabaseConnection failed: %v", err)
	}

	_, err = c.UpdateDatabaseConnection(context.Background(), db)
	if err != nil {
		t.Errorf("UpdateDatabaseConnection failed: %v", err)
	}
}

func TestClient_UserSourceOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := []ResourceResponse[UserSourceConfig]{
			{
				Name:      "test-us",
				Signature: "sig-us",
				Config:    UserSourceConfig{Profile: UserSourceProfile{Type: "internal"}},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	us := ResourceResponse[UserSourceConfig]{Name: "test-us", Config: UserSourceConfig{Profile: UserSourceProfile{Type: "internal"}}}

	_, err := c.CreateUserSource(context.Background(), us)
	if err != nil {
		t.Errorf("CreateUserSource failed: %v", err)
	}

	_, err = c.UpdateUserSource(context.Background(), us)
	if err != nil {
		t.Errorf("UpdateUserSource failed: %v", err)
	}
}

func TestClient_TagProviderOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := []ResourceResponse[TagProviderConfig]{
			{
				Name:      "test-tp",
				Signature: "sig-tp",
				Config:    TagProviderConfig{Profile: TagProviderProfile{Type: "STANDARD"}},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	tp := ResourceResponse[TagProviderConfig]{Name: "test-tp", Config: TagProviderConfig{Profile: TagProviderProfile{Type: "STANDARD"}}}

	_, err := c.CreateTagProvider(context.Background(), tp)
	if err != nil {
		t.Errorf("CreateTagProvider failed: %v", err)
	}

	_, err = c.UpdateTagProvider(context.Background(), tp)
	if err != nil {
		t.Errorf("UpdateTagProvider failed: %v", err)
	}
}

func TestClient_AuditProfileOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := []ResourceResponse[AuditProfileConfig]{
			{
				Name:      "test-ap",
				Signature: "sig-ap",
				Config:    AuditProfileConfig{Profile: AuditProfileProfile{Type: "database"}},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	ap := ResourceResponse[AuditProfileConfig]{Name: "test-ap", Config: AuditProfileConfig{Profile: AuditProfileProfile{Type: "database"}}}

	_, err := c.CreateAuditProfile(context.Background(), ap)
	if err != nil {
		t.Errorf("CreateAuditProfile failed: %v", err)
	}

	_, err = c.UpdateAuditProfile(context.Background(), ap)
	if err != nil {
		t.Errorf("UpdateAuditProfile failed: %v", err)
	}
}

func TestClient_ModuleResourceOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Just return a generic success for these module-based calls
		response := []ResourceResponse[map[string]any]{
			{
				Name:      "test-res",
				Signature: "sig-mod",
				Config:    map[string]any{},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)

	anp := ResourceResponse[AlarmNotificationProfileConfig]{Name: "anp"}
	_, err := c.CreateAlarmNotificationProfile(context.Background(), anp)
	if err != nil {
		t.Errorf("CreateAlarmNotificationProfile failed: %v", err)
	}

	opc := ResourceResponse[OpcUaConnectionConfig]{Name: "opc"}
	_, err = c.CreateOpcUaConnection(context.Background(), opc)
	if err != nil {
		t.Errorf("CreateOpcUaConnection failed: %v", err)
	}

	dev := ResourceResponse[DeviceConfig]{Name: "dev"}
	_, err = c.CreateDevice(context.Background(), dev)
	if err != nil {
		t.Errorf("CreateDevice failed: %v", err)
	}
}

func TestClient_RedundancyOperations(t *testing.T) {
	var capturedMethod string
	var capturedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// The real gateway uses "masterRecoveryMode", not "recoveryMode".
			_, _ = w.Write([]byte(`{"role":"Independent","masterRecoveryMode":"Automatic"}`))
			return
		}
		capturedMethod = r.Method
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)

	config, err := c.GetRedundancyConfig(context.Background())
	if err != nil {
		t.Fatalf("GetRedundancyConfig failed: %v", err)
	}
	if config.Role != "Independent" {
		t.Errorf("Expected Role Independent, got %s", config.Role)
	}
	if config.RecoveryMode != "Automatic" {
		t.Errorf("expected RecoveryMode to be unmarshaled from masterRecoveryMode, got %q", config.RecoveryMode)
	}

	config.RecoveryMode = "Manual"
	if err := c.UpdateRedundancyConfig(context.Background(), *config); err != nil {
		t.Fatalf("UpdateRedundancyConfig failed: %v", err)
	}
	if capturedMethod != http.MethodPut {
		t.Errorf("expected UpdateRedundancyConfig to use PUT (the endpoint doesn't support POST), got %s", capturedMethod)
	}
	if !bytes.Contains(capturedBody, []byte(`"masterRecoveryMode":"Manual"`)) {
		t.Errorf("expected request body to use the masterRecoveryMode key, got: %s", capturedBody)
	}
}

func TestClient_AlarmJournalOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			response := ResourceResponse[AlarmJournalConfig]{
				Name:      "test-journal",
				Signature: "sig-journal",
				Config:    AlarmJournalConfig{Profile: AlarmJournalProfile{Type: "database"}},
			}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		response := []ResourceResponse[AlarmJournalConfig]{
			{
				Name:      "test-journal",
				Signature: "sig-journal",
				Config:    AlarmJournalConfig{Profile: AlarmJournalProfile{Type: "database"}},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	item := ResourceResponse[AlarmJournalConfig]{Name: "test-journal"}

	if _, err := c.CreateAlarmJournal(context.Background(), item); err != nil {
		t.Errorf("CreateAlarmJournal failed: %v", err)
	}
	if _, err := c.UpdateAlarmJournal(context.Background(), item); err != nil {
		t.Errorf("UpdateAlarmJournal failed: %v", err)
	}
	if _, err := c.GetAlarmJournal(context.Background(), "test-journal"); err != nil {
		t.Errorf("GetAlarmJournal failed: %v", err)
	}
	if err := c.DeleteAlarmJournal(context.Background(), "test-journal", "sig"); err != nil {
		t.Errorf("DeleteAlarmJournal failed: %v", err)
	}
}

func TestClient_SMTPProfileOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			response := ResourceResponse[SMTPProfileConfig]{
				Name:      "test-smtp",
				Signature: "sig-smtp",
				Config:    SMTPProfileConfig{Profile: SMTPProfileProfile{Type: "classic"}},
			}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		response := []ResourceResponse[SMTPProfileConfig]{
			{
				Name:      "test-smtp",
				Signature: "sig-smtp",
				Config:    SMTPProfileConfig{Profile: SMTPProfileProfile{Type: "classic"}},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	item := ResourceResponse[SMTPProfileConfig]{Name: "test-smtp"}

	if _, err := c.CreateSMTPProfile(context.Background(), item); err != nil {
		t.Errorf("CreateSMTPProfile failed: %v", err)
	}
	if _, err := c.UpdateSMTPProfile(context.Background(), item); err != nil {
		t.Errorf("UpdateSMTPProfile failed: %v", err)
	}
	if _, err := c.GetSMTPProfile(context.Background(), "test-smtp"); err != nil {
		t.Errorf("GetSMTPProfile failed: %v", err)
	}
	if err := c.DeleteSMTPProfile(context.Background(), "test-smtp", "sig"); err != nil {
		t.Errorf("DeleteSMTPProfile failed: %v", err)
	}
}

func TestClient_StoreAndForwardOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			response := ResourceResponse[StoreAndForwardConfig]{
				Name:      "test-sf",
				Signature: "sig-sf",
				Config:    StoreAndForwardConfig{ForwardingPolicy: "ALL"},
			}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		response := []ResourceResponse[StoreAndForwardConfig]{
			{
				Name:      "test-sf",
				Signature: "sig-sf",
				Config:    StoreAndForwardConfig{ForwardingPolicy: "ALL"},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	item := ResourceResponse[StoreAndForwardConfig]{Name: "test-sf"}

	if _, err := c.CreateStoreAndForward(context.Background(), item); err != nil {
		t.Errorf("CreateStoreAndForward failed: %v", err)
	}
	if _, err := c.UpdateStoreAndForward(context.Background(), item); err != nil {
		t.Errorf("UpdateStoreAndForward failed: %v", err)
	}
	if _, err := c.GetStoreAndForward(context.Background(), "test-sf"); err != nil {
		t.Errorf("GetStoreAndForward failed: %v", err)
	}
	if err := c.DeleteStoreAndForward(context.Background(), "test-sf", "sig"); err != nil {
		t.Errorf("DeleteStoreAndForward failed: %v", err)
	}
}

func TestClient_IdentityProviderOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			response := ResourceResponse[IdentityProviderConfig]{
				Name:      "test-idp",
				Signature: "sig-idp",
				Config:    IdentityProviderConfig{Profile: IdentityProviderProfile{Type: "Ignition"}},
			}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		response := []ResourceResponse[IdentityProviderConfig]{
			{
				Name:      "test-idp",
				Signature: "sig-idp",
				Config:    IdentityProviderConfig{Profile: IdentityProviderProfile{Type: "Ignition"}},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	item := ResourceResponse[IdentityProviderConfig]{Name: "test-idp"}

	if _, err := c.CreateIdentityProvider(context.Background(), item); err != nil {
		t.Errorf("CreateIdentityProvider failed: %v", err)
	}
	if _, err := c.UpdateIdentityProvider(context.Background(), item); err != nil {
		t.Errorf("UpdateIdentityProvider failed: %v", err)
	}
	if _, err := c.GetIdentityProvider(context.Background(), "test-idp"); err != nil {
		t.Errorf("GetIdentityProvider failed: %v", err)
	}
	if err := c.DeleteIdentityProvider(context.Background(), "test-idp", "sig"); err != nil {
		t.Errorf("DeleteIdentityProvider failed: %v", err)
	}
}

func TestClient_GanOutgoingOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			response := ResourceResponse[GanOutgoingConfig]{
				Name:      "test-gan",
				Signature: "sig-gan",
				Config:    GanOutgoingConfig{Host: "localhost"},
			}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		response := []ResourceResponse[GanOutgoingConfig]{
			{
				Name:      "test-gan",
				Signature: "sig-gan",
				Config:    GanOutgoingConfig{Host: "localhost"},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	item := ResourceResponse[GanOutgoingConfig]{Name: "test-gan"}

	if _, err := c.CreateGanOutgoing(context.Background(), item); err != nil {
		t.Errorf("CreateGanOutgoing failed: %v", err)
	}
	if _, err := c.UpdateGanOutgoing(context.Background(), item); err != nil {
		t.Errorf("UpdateGanOutgoing failed: %v", err)
	}
	if _, err := c.GetGanOutgoing(context.Background(), "test-gan"); err != nil {
		t.Errorf("GetGanOutgoing failed: %v", err)
	}
	if err := c.DeleteGanOutgoing(context.Background(), "test-gan", "sig"); err != nil {
		t.Errorf("DeleteGanOutgoing failed: %v", err)
	}
}

func TestClient_GanGeneralSettingsOperations(t *testing.T) {
	type call struct{ method, path string }
	var calls []call

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, call{r.Method, r.URL.Path})
		if r.Method == http.MethodGet {
			// Singleton responses never include a "name" field.
			_, _ = w.Write([]byte(`{"signature":"sig-settings","config":{"allowIncoming":true}}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)

	got, err := c.GetGanGeneralSettings(context.Background())
	if err != nil {
		t.Fatalf("GetGanGeneralSettings failed: %v", err)
	}
	if len(calls) != 1 || calls[0].path != "/data/api/v1/resources/singleton/ignition/gateway-network-settings" {
		t.Errorf("expected GetGanGeneralSettings to use the singleton path, got %+v", calls)
	}
	if got.Name != "gateway-network-settings" {
		t.Errorf("expected Name to be filled in despite the response omitting it, got %q", got.Name)
	}

	calls = nil
	item := ResourceResponse[GanGeneralSettingsConfig]{Name: "gateway-network-settings", Signature: "sig-settings"}
	updated, err := c.UpdateGanGeneralSettings(context.Background(), item)
	if err != nil {
		t.Fatalf("UpdateGanGeneralSettings failed: %v", err)
	}
	if len(calls) != 2 {
		t.Fatalf("expected UpdateGanGeneralSettings to PUT then re-fetch via singleton GET, got %+v", calls)
	}
	if calls[0].method != http.MethodPut || calls[0].path != "/data/api/v1/resources/ignition/gateway-network-settings" {
		t.Errorf("expected first call to be PUT to the resource path, got %+v", calls[0])
	}
	if calls[1].method != http.MethodGet || calls[1].path != "/data/api/v1/resources/singleton/ignition/gateway-network-settings" {
		t.Errorf("expected follow-up call to be a singleton GET (find-by-name doesn't exist for singletons), got %+v", calls[1])
	}
	if updated.Name != "gateway-network-settings" {
		t.Errorf("expected updated Name to be filled in, got %q", updated.Name)
	}
}

func TestClient_DeviceOperations(t *testing.T) {
	var capturedCreateBody, capturedUpdateBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// The driver flavor lives in config.profile.type, distinct from
			// the outer generic resource-type string.
			_, _ = w.Write([]byte(`{
				"type": "com.inductiveautomation.opcua/device",
				"name": "test-device",
				"signature": "sig-device",
				"config": {
					"profile": {"type": "ProgrammableSimulatorDevice"},
					"settings": {"timeIntervalRate": 1000}
				}
			}`))
			return
		}

		body, _ := io.ReadAll(r.Body)
		if r.Method == http.MethodPost {
			capturedCreateBody = body
		} else {
			capturedUpdateBody = body
		}
		_, _ = w.Write([]byte(`{"success":true,"changes":[{"name":"test-device","newSignature":"sig-device"}]}`))
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	item := ResourceResponse[DeviceConfig]{
		Name:   "test-device",
		Type:   "ProgrammableSimulatorDevice",
		Config: DeviceConfig{"timeIntervalRate": float64(1000)},
	}

	if _, err := c.CreateDevice(context.Background(), item); err != nil {
		t.Fatalf("CreateDevice failed: %v", err)
	}
	if !bytes.Contains(capturedCreateBody, []byte(`"profile":{"type":"ProgrammableSimulatorDevice"}`)) {
		t.Errorf("expected CreateDevice to wrap the driver type under config.profile.type, got: %s", capturedCreateBody)
	}
	if !bytes.Contains(capturedCreateBody, []byte(`"settings":{"timeIntervalRate":1000}`)) {
		t.Errorf("expected CreateDevice to nest params under config.settings, got: %s", capturedCreateBody)
	}

	if _, err := c.UpdateDevice(context.Background(), item); err != nil {
		t.Fatalf("UpdateDevice failed: %v", err)
	}
	if !bytes.Contains(capturedUpdateBody, []byte(`"profile":{"type":"ProgrammableSimulatorDevice"}`)) {
		t.Errorf("expected UpdateDevice to wrap the driver type under config.profile.type, got: %s", capturedUpdateBody)
	}

	got, err := c.GetDevice(context.Background(), "test-device")
	if err != nil {
		t.Fatalf("GetDevice failed: %v", err)
	}
	if got.Type != "ProgrammableSimulatorDevice" {
		t.Errorf("expected Type to be extracted from config.profile.type (not the outer generic type), got %q", got.Type)
	}
	if _, ok := got.Config["timeIntervalRate"]; !ok {
		t.Errorf("expected Config to be extracted from config.settings, got %+v", got.Config)
	}

	if err := c.DeleteDevice(context.Background(), "test-device", "sig"); err != nil {
		t.Errorf("DeleteDevice failed: %v", err)
	}
}

func TestClient_EncryptSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("Expected Content-Type text/plain, got %s", r.Header.Get("Content-Type"))
		}

		response := map[string]interface{}{"jwe": "mock-jwe"}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)

	secret, err := c.EncryptSecret(context.Background(), "my-password")
	if err != nil {
		t.Errorf("EncryptSecret failed: %v", err)
	}
	if secret.Type != "Embedded" {
		t.Errorf("Expected Type Embedded, got %s", secret.Type)
	}
	dataMap := secret.Data.(map[string]interface{})
	if dataMap["jwe"] != "mock-jwe" {
		t.Errorf("Expected jwe mock-jwe, got %v", dataMap["jwe"])
	}
}

func TestClient_EncryptSecret_RetryOn503(t *testing.T) {
	var reqCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&reqCount, 1)
		if count <= 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("Service Restarting"))
			return
		}
		response := map[string]interface{}{"jwe": "mock-jwe"}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c, err := NewClient(server.URL, "token", false)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	c.HTTPClient.RetryWaitMin = 10 * time.Millisecond
	c.HTTPClient.RetryWaitMax = 50 * time.Millisecond

	secret, err := c.EncryptSecret(context.Background(), "my-password")
	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}
	if atomic.LoadInt32(&reqCount) != 4 {
		t.Errorf("Expected 4 requests (3 retries), got %d", reqCount)
	}
	dataMap := secret.Data.(map[string]interface{})
	if dataMap["jwe"] != "mock-jwe" {
		t.Errorf("Expected jwe mock-jwe, got %v", dataMap["jwe"])
	}
}

func TestClient_EncryptSecret_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"success": false,
			"messages": ["Encryption failed"]
		}`))
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)

	_, err := c.EncryptSecret(context.Background(), "my-password")
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "Encryption failed") {
		t.Errorf("Expected structured error message, got: %s", err.Error())
	}
}

func boolPtr(b bool) *bool {
	return &b
}
