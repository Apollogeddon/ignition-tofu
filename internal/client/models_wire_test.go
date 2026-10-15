package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// A false or 0 the configuration asks for reaches the gateway; omitempty on a plain bool or
// int dropped it, so a project couldn't be disabled or audit pruning turned off.
func TestModels_SendFalseAndZero(t *testing.T) {
	f, zero := false, 0
	cases := []struct {
		name string
		v    any
		want []string
		not  []string
	}{
		{"project false", Project{Name: "p", Enabled: &f, Inheritable: &f}, []string{`"enabled":false`, `"inheritable":false`}, nil},
		{"project unset", Project{Name: "p"}, nil, []string{`enabled`, `inheritable`}},
		{"audit profile", AuditProfileConfig{
			Profile:  AuditProfileProfile{Type: "database", RetentionDays: &zero},
			Settings: AuditProfileSettings{PruneEnabled: &f, AutoCreate: &f, EnableStoreAndForward: &f},
		}, []string{`"retentionDays":0`, `"pruneEnabled":false`, `"autoCreate":false`, `"enableStoreAndForward":false`}, nil},
		{"audit profile unset", AuditProfileConfig{Profile: AuditProfileProfile{Type: "database"}}, nil, []string{`retentionDays`, `pruneEnabled`, `autoCreate`}},
		{"user source", UserSourceConfig{Profile: UserSourceProfile{Type: "internal"}}, []string{`"scheduleRestricted":false`}, nil},
	}
	for _, tc := range cases {
		b, err := json.Marshal(tc.v)
		if err != nil {
			t.Fatal(err)
		}
		for _, w := range tc.want {
			if !strings.Contains(string(b), w) {
				t.Errorf("%s: %s lacks %s", tc.name, b, w)
			}
		}
		for _, n := range tc.not {
			if strings.Contains(string(b), n) {
				t.Errorf("%s: %s sends %s", tc.name, b, n)
			}
		}
	}

	var redundancy RedundancyConfig
	redundancy.GatewayNetworkSetup = &struct {
		Host               string  `json:"host,omitempty"`
		Port               int     `json:"port,omitempty"`
		EnableSsl          bool    `json:"enableSsl"`
		PingRate           float64 `json:"pingRate,omitempty"`
		PingTimeout        float64 `json:"pingTimeout,omitempty"`
		PingMaxMissed      float64 `json:"pingMaxMissed,omitempty"`
		WebsocketTimeout   float64 `json:"websocketTimeout,omitempty"`
		HttpConnectTimeout float64 `json:"httpConnectTimeout,omitempty"`
		HttpReadTimeout    float64 `json:"httpReadTimeout,omitempty"`
		SendThreads        float64 `json:"sendThreads,omitempty"`
		ReceiveThreads     float64 `json:"receiveThreads,omitempty"`
	}{}
	if b, _ := json.Marshal(redundancy); !strings.Contains(string(b), `"enableSsl":false`) {
		t.Errorf("redundancy: %s lacks enableSsl false", b)
	}
}

// An update to enabled = false sends it, and waits until the gateway shows it.
func TestClient_UpdateProjectDisables(t *testing.T) {
	var body string
	var gets int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			b, _ := io.ReadAll(r.Body)
			body = string(b)
			return
		}
		gets++
		// the first reads still show it enabled
		_ = json.NewEncoder(w).Encode(map[string]any{"name": "p", "enabled": gets <= 2})
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	c.HTTPClient.RetryWaitMin = 10 * time.Millisecond
	f := false
	p, err := c.UpdateProject(context.Background(), Project{Name: "p", Enabled: &f})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, `"enabled":false`) {
		t.Errorf("PUT body %s lacks enabled false", body)
	}
	if p.Enabled == nil || *p.Enabled {
		t.Errorf("Expected the disabled project, got %+v", p)
	}
}
