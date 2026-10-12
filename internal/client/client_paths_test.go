package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Every typed read and delete goes to its own resource type, a mistyped one
// being a resource the gateway can't find.
func TestClient_TypedReadsAndDeletes(t *testing.T) {
	var method, path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.EscapedPath()
		_, _ = w.Write([]byte(`{"name": "n", "config": {}}`))
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "token", false)
	ctx := context.Background()
	find := func(module, kind string) string { return "/data/api/v1/resources/find/" + module + "/" + kind + "/n" }
	del := func(module, kind string) string { return "/data/api/v1/resources/" + module + "/" + kind + "/n/s" }
	get := func(f func() (any, error)) func() error { return func() error { _, err := f(); return err } }

	cases := []struct {
		name, method, path string
		call               func() error
	}{
		{"GetDatabaseConnection", http.MethodGet, find("ignition", "database-connection"), get(func() (any, error) { return c.GetDatabaseConnection(ctx, "n") })},
		{"DeleteDatabaseConnection", http.MethodDelete, del("ignition", "database-connection"), func() error { return c.DeleteDatabaseConnection(ctx, "n", "s") }},
		{"GetUserSource", http.MethodGet, find("ignition", "user-source"), get(func() (any, error) { return c.GetUserSource(ctx, "n") })},
		{"DeleteUserSource", http.MethodDelete, del("ignition", "user-source"), func() error { return c.DeleteUserSource(ctx, "n", "s") }},
		{"GetTagProvider", http.MethodGet, find("ignition", "tag-provider"), get(func() (any, error) { return c.GetTagProvider(ctx, "n") })},
		{"DeleteTagProvider", http.MethodDelete, del("ignition", "tag-provider"), func() error { return c.DeleteTagProvider(ctx, "n", "s") }},
		{"GetAuditProfile", http.MethodGet, find("ignition", "audit-profile"), get(func() (any, error) { return c.GetAuditProfile(ctx, "n") })},
		{"DeleteAuditProfile", http.MethodDelete, del("ignition", "audit-profile"), func() error { return c.DeleteAuditProfile(ctx, "n", "s") }},
		{"GetAlarmNotificationProfile", http.MethodGet, find("com.inductiveautomation.alarm-notification", "alarm-notification-profile"), get(func() (any, error) { return c.GetAlarmNotificationProfile(ctx, "n") })},
		{"DeleteAlarmNotificationProfile", http.MethodDelete, del("com.inductiveautomation.alarm-notification", "alarm-notification-profile"), func() error { return c.DeleteAlarmNotificationProfile(ctx, "n", "s") }},
		{"GetOpcUaConnection", http.MethodGet, find("ignition", "opc-connection"), get(func() (any, error) { return c.GetOpcUaConnection(ctx, "n") })},
		{"DeleteOpcUaConnection", http.MethodDelete, del("ignition", "opc-connection"), func() error { return c.DeleteOpcUaConnection(ctx, "n", "s") }},
		{"GetAlarmJournal", http.MethodGet, find("ignition", "alarm-journal"), get(func() (any, error) { return c.GetAlarmJournal(ctx, "n") })},
		{"DeleteAlarmJournal", http.MethodDelete, del("ignition", "alarm-journal"), func() error { return c.DeleteAlarmJournal(ctx, "n", "s") }},
		{"GetSMTPProfile", http.MethodGet, find("ignition", "email-profile"), get(func() (any, error) { return c.GetSMTPProfile(ctx, "n") })},
		{"DeleteSMTPProfile", http.MethodDelete, del("ignition", "email-profile"), func() error { return c.DeleteSMTPProfile(ctx, "n", "s") }},
		{"GetStoreAndForward", http.MethodGet, find("ignition", "store-and-forward-engine"), get(func() (any, error) { return c.GetStoreAndForward(ctx, "n") })},
		{"DeleteStoreAndForward", http.MethodDelete, del("ignition", "store-and-forward-engine"), func() error { return c.DeleteStoreAndForward(ctx, "n", "s") }},
		{"GetIdentityProvider", http.MethodGet, find("ignition", "identity-provider"), get(func() (any, error) { return c.GetIdentityProvider(ctx, "n") })},
		{"DeleteIdentityProvider", http.MethodDelete, del("ignition", "identity-provider"), func() error { return c.DeleteIdentityProvider(ctx, "n", "s") }},
		{"GetGanOutgoing", http.MethodGet, find("ignition", "gateway-network-outgoing"), get(func() (any, error) { return c.GetGanOutgoing(ctx, "n") })},
		{"DeleteGanOutgoing", http.MethodDelete, del("ignition", "gateway-network-outgoing"), func() error { return c.DeleteGanOutgoing(ctx, "n", "s") }},
		{"GetDevice", http.MethodGet, find("com.inductiveautomation.opcua", "device"), get(func() (any, error) { return c.GetDevice(ctx, "n") })},
		{"DeleteDevice", http.MethodDelete, del("com.inductiveautomation.opcua", "device"), func() error { return c.DeleteDevice(ctx, "n", "s") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); err != nil {
				t.Fatalf("%s failed: %v", tc.name, err)
			}
			if method != tc.method || path != tc.path {
				t.Errorf("got %s %s, want %s %s", method, path, tc.method, tc.path)
			}
		})
	}
}
