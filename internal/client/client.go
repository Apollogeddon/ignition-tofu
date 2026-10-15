package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

type IgnitionClient interface {
	GetResource(ctx context.Context, resourceType, name string, dest any) error
	CreateResource(ctx context.Context, resourceType string, item any, dest any) error
	UpdateResource(ctx context.Context, resourceType string, item any, dest any) error
	DeleteResource(ctx context.Context, resourceType, name, signature string) error
	GetResourceWithModule(ctx context.Context, module, resourceType, name string, dest any) error
	CreateResourceWithModule(ctx context.Context, module, resourceType string, item any, dest any) error
	UpdateResourceWithModule(ctx context.Context, module, resourceType string, item any, dest any) error
	DeleteResourceWithModule(ctx context.Context, module, resourceType, name, signature string) error
	EncryptSecret(ctx context.Context, plaintext string) (*IgnitionSecret, error)
	GetProject(ctx context.Context, name string) (*Project, error)
	CreateProject(ctx context.Context, p Project) (*Project, error)
	UpdateProject(ctx context.Context, p Project) (*Project, error)
	DeleteProject(ctx context.Context, name string) error
	GetDatabaseConnection(ctx context.Context, name string) (*ResourceResponse[DatabaseConfig], error)
	CreateDatabaseConnection(ctx context.Context, db ResourceResponse[DatabaseConfig]) (*ResourceResponse[DatabaseConfig], error)
	UpdateDatabaseConnection(ctx context.Context, db ResourceResponse[DatabaseConfig]) (*ResourceResponse[DatabaseConfig], error)
	DeleteDatabaseConnection(ctx context.Context, name, signature string) error
	GetUserSource(ctx context.Context, name string) (*ResourceResponse[UserSourceConfig], error)
	CreateUserSource(ctx context.Context, us ResourceResponse[UserSourceConfig]) (*ResourceResponse[UserSourceConfig], error)
	UpdateUserSource(ctx context.Context, us ResourceResponse[UserSourceConfig]) (*ResourceResponse[UserSourceConfig], error)
	DeleteUserSource(ctx context.Context, name, signature string) error
	GetTagProvider(ctx context.Context, name string) (*ResourceResponse[TagProviderConfig], error)
	CreateTagProvider(ctx context.Context, tp ResourceResponse[TagProviderConfig]) (*ResourceResponse[TagProviderConfig], error)
	UpdateTagProvider(ctx context.Context, tp ResourceResponse[TagProviderConfig]) (*ResourceResponse[TagProviderConfig], error)
	DeleteTagProvider(ctx context.Context, name, signature string) error
	GetAuditProfile(ctx context.Context, name string) (*ResourceResponse[AuditProfileConfig], error)
	CreateAuditProfile(ctx context.Context, ap ResourceResponse[AuditProfileConfig]) (*ResourceResponse[AuditProfileConfig], error)
	UpdateAuditProfile(ctx context.Context, ap ResourceResponse[AuditProfileConfig]) (*ResourceResponse[AuditProfileConfig], error)
	DeleteAuditProfile(ctx context.Context, name, signature string) error
	GetAlarmNotificationProfile(ctx context.Context, name string) (*ResourceResponse[AlarmNotificationProfileConfig], error)
	CreateAlarmNotificationProfile(ctx context.Context, anp ResourceResponse[AlarmNotificationProfileConfig]) (*ResourceResponse[AlarmNotificationProfileConfig], error)
	UpdateAlarmNotificationProfile(ctx context.Context, anp ResourceResponse[AlarmNotificationProfileConfig]) (*ResourceResponse[AlarmNotificationProfileConfig], error)
	DeleteAlarmNotificationProfile(ctx context.Context, name, signature string) error
	GetOpcUaConnection(ctx context.Context, name string) (*ResourceResponse[OpcUaConnectionConfig], error)
	CreateOpcUaConnection(ctx context.Context, item ResourceResponse[OpcUaConnectionConfig]) (*ResourceResponse[OpcUaConnectionConfig], error)
	UpdateOpcUaConnection(ctx context.Context, item ResourceResponse[OpcUaConnectionConfig]) (*ResourceResponse[OpcUaConnectionConfig], error)
	DeleteOpcUaConnection(ctx context.Context, name, signature string) error
	GetAlarmJournal(ctx context.Context, name string) (*ResourceResponse[AlarmJournalConfig], error)
	CreateAlarmJournal(ctx context.Context, item ResourceResponse[AlarmJournalConfig]) (*ResourceResponse[AlarmJournalConfig], error)
	UpdateAlarmJournal(ctx context.Context, item ResourceResponse[AlarmJournalConfig]) (*ResourceResponse[AlarmJournalConfig], error)
	DeleteAlarmJournal(ctx context.Context, name, signature string) error
	GetSMTPProfile(ctx context.Context, name string) (*ResourceResponse[SMTPProfileConfig], error)
	CreateSMTPProfile(ctx context.Context, item ResourceResponse[SMTPProfileConfig]) (*ResourceResponse[SMTPProfileConfig], error)
	UpdateSMTPProfile(ctx context.Context, item ResourceResponse[SMTPProfileConfig]) (*ResourceResponse[SMTPProfileConfig], error)
	DeleteSMTPProfile(ctx context.Context, name, signature string) error
	GetStoreAndForward(ctx context.Context, name string) (*ResourceResponse[StoreAndForwardConfig], error)
	CreateStoreAndForward(ctx context.Context, item ResourceResponse[StoreAndForwardConfig]) (*ResourceResponse[StoreAndForwardConfig], error)
	UpdateStoreAndForward(ctx context.Context, item ResourceResponse[StoreAndForwardConfig]) (*ResourceResponse[StoreAndForwardConfig], error)
	DeleteStoreAndForward(ctx context.Context, name, signature string) error
	GetIdentityProvider(ctx context.Context, name string) (*ResourceResponse[IdentityProviderConfig], error)
	CreateIdentityProvider(ctx context.Context, item ResourceResponse[IdentityProviderConfig]) (*ResourceResponse[IdentityProviderConfig], error)
	UpdateIdentityProvider(ctx context.Context, item ResourceResponse[IdentityProviderConfig]) (*ResourceResponse[IdentityProviderConfig], error)
	DeleteIdentityProvider(ctx context.Context, name, signature string) error
	GetGanOutgoing(ctx context.Context, name string) (*ResourceResponse[GanOutgoingConfig], error)
	CreateGanOutgoing(ctx context.Context, item ResourceResponse[GanOutgoingConfig]) (*ResourceResponse[GanOutgoingConfig], error)
	UpdateGanOutgoing(ctx context.Context, item ResourceResponse[GanOutgoingConfig]) (*ResourceResponse[GanOutgoingConfig], error)
	DeleteGanOutgoing(ctx context.Context, name, signature string) error
	GetRedundancyConfig(ctx context.Context) (*RedundancyConfig, error)
	UpdateRedundancyConfig(ctx context.Context, config RedundancyConfig) error
	GetGanGeneralSettings(ctx context.Context) (*ResourceResponse[GanGeneralSettingsConfig], error)
	UpdateGanGeneralSettings(ctx context.Context, item ResourceResponse[GanGeneralSettingsConfig]) (*ResourceResponse[GanGeneralSettingsConfig], error)
	GetDevice(ctx context.Context, name string) (*ResourceResponse[DeviceConfig], error)
	CreateDevice(ctx context.Context, item ResourceResponse[DeviceConfig]) (*ResourceResponse[DeviceConfig], error)
	UpdateDevice(ctx context.Context, item ResourceResponse[DeviceConfig]) (*ResourceResponse[DeviceConfig], error)
	DeleteDevice(ctx context.Context, name, signature string) error
}

// ErrNotFound reports that the gateway answered 404 Not Found: the thing
// requested does not exist. Match it with errors.Is.
var ErrNotFound = errors.New("not found")

// notFoundError marks a 404 as ErrNotFound while keeping the gateway's own
// error, so its message and errors.As(*APIErrorResponse) are unchanged.
type notFoundError struct{ err error }

func (e notFoundError) Error() string        { return e.err.Error() }
func (e notFoundError) Unwrap() error        { return e.err }
func (e notFoundError) Is(target error) bool { return target == ErrNotFound }

type Client struct {
	HostURL    string
	HTTPClient *retryablehttp.Client
	Token      string
}

func NewClient(host, token string, allowInsecureTLS bool) (*Client, error) {
	rc := retryablehttp.NewClient()
	rc.RetryMax = 10
	rc.Logger = nil
	rc.ErrorHandler = retryablehttp.PassthroughErrorHandler
	rc.CheckRetry = retryPolicy
	rc.HTTPClient.Timeout = 10 * time.Second

	if allowInsecureTLS {
		// a clone keeps the default transport's proxy settings and dial and handshake timeouts
		t := http.DefaultTransport.(*http.Transport).Clone()
		// allow_insecure_tls is an explicit opt-in, for gateways with a self-signed certificate
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // see above
		rc.HTTPClient.Transport = t
	}

	return &Client{HTTPClient: rc, HostURL: strings.TrimRight(host, "/"), Token: token}, nil
}

// maxResponseBytes caps what is read of one response; the gateway's are a
// few kilobytes, so anything near this is not a gateway answering.
var maxResponseBytes = 32 << 20

// noResendKey marks a request that must not be sent twice: a repeated create
// fails with "already exists", and a repeated signed update with "signature
// mismatch", when the first one reached the gateway.
type noResendKey struct{}

// sendOnce marks a resource update as not to be sent twice: it carries the signature the
// gateway checks, which the first update changes if it was applied, so a resent one fails
// with a signature mismatch. A create (POST) is marked in doRequest.
func sendOnce(ctx context.Context, method string) context.Context {
	if method == http.MethodPut {
		return context.WithValue(ctx, noResendKey{}, true)
	}
	return ctx
}

// retryPolicy is retryablehttp's default, except that a request marked with
// noResendKey is retried only when the gateway can't have acted on it: the
// connection was refused, or it answered 429 or 503.
func retryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if ctx.Err() == nil && ctx.Value(noResendKey{}) != nil {
		if err != nil {
			var opErr *net.OpError
			return errors.As(err, &opErr) && opErr.Op == "dial", nil
		}
		return resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable, nil
	}
	return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body []byte, contentType ...string) ([]byte, error) {
	if method == http.MethodPost {
		ctx = context.WithValue(ctx, noResendKey{}, true)
	}
	req, err := retryablehttp.NewRequestWithContext(ctx, method, c.HostURL+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	ct := "application/json"
	if len(contentType) > 0 {
		ct = contentType[0]
	}

	req.Header.Set("X-Ignition-API-Token", c.Token)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("Accept", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	// one byte past the cap tells a body at the cap from a longer one
	bodyBytes, err := io.ReadAll(io.LimitReader(res.Body, int64(maxResponseBytes)+1))
	if err != nil {
		return nil, err
	}
	if len(bodyBytes) > maxResponseBytes {
		return nil, fmt.Errorf("response from %s %s is over %d bytes", method, path, maxResponseBytes)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		err := statusError(res.StatusCode, bodyBytes)
		if res.StatusCode == http.StatusNotFound {
			return nil, notFoundError{err}
		}
		return nil, err
	}

	return bodyBytes, nil
}

// apiPath appends each segment to base, escaped, so a name holding ?, #, / or %
// stays one path segment rather than becoming a query, fragment or extra segment.
func apiPath(base string, segments ...string) string {
	var b strings.Builder
	b.WriteString(base)
	for _, s := range segments {
		b.WriteByte('/')
		b.WriteString(url.PathEscape(s))
	}
	return b.String()
}

func statusError(status int, body []byte) error {
	var apiErr APIErrorResponse
	if err := json.Unmarshal(body, &apiErr); err == nil &&
		(apiErr.Problem != nil || len(apiErr.Messages) > 0 || len(apiErr.FieldMessages) > 0) {
		return &apiErr
	}
	return fmt.Errorf("status: %d, body: %s", status, body)
}

func (c *Client) GetResource(ctx context.Context, resourceType, name string, dest any) error {
	return c.GetResourceWithModule(ctx, "ignition", resourceType, name, dest)
}

func (c *Client) GetResourceWithModule(ctx context.Context, module, resourceType, name string, dest any) error {
	path := apiPath("/data/api/v1/resources/find", module, resourceType, name)
	body, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dest)
}

// GetResourceSingleton reads a gateway-wide singleton resource (one that has
// no name of its own, e.g. gateway network settings), which the API exposes
// under a dedicated path rather than the name-based find endpoint.
func (c *Client) GetResourceSingleton(ctx context.Context, module, resourceType string, dest any) error {
	path := apiPath("/data/api/v1/resources/singleton", module, resourceType)
	body, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dest)
}

func (c *Client) createOrUpdate(ctx context.Context, method, module, resourceType string, item, dest any) error {
	rb, err := json.Marshal([]any{item})
	if err != nil {
		return err
	}

	path := apiPath("/data/api/v1/resources", module, resourceType)
	body, err := c.doRequest(sendOnce(ctx, method), method, path, rb)
	if err != nil {
		return err
	}

	return c.unmarshalResourceResponse(ctx, module, resourceType, body, dest)
}

func (c *Client) CreateResource(ctx context.Context, resourceType string, item, dest any) error {
	return c.createOrUpdate(ctx, http.MethodPost, "ignition", resourceType, item, dest)
}

func (c *Client) UpdateResource(ctx context.Context, resourceType string, item, dest any) error {
	return c.createOrUpdate(ctx, http.MethodPut, "ignition", resourceType, item, dest)
}

// updateSingleton is like createOrUpdate but follows up with
// GetResourceSingleton rather than the name-based find endpoint, since
// singleton resources have no name to look up by.
func (c *Client) updateSingleton(ctx context.Context, module, resourceType string, item, dest any) error {
	rb, err := json.Marshal([]any{item})
	if err != nil {
		return err
	}

	path := apiPath("/data/api/v1/resources", module, resourceType)
	if _, err := c.doRequest(sendOnce(ctx, http.MethodPut), http.MethodPut, path, rb); err != nil {
		return err
	}

	return c.GetResourceSingleton(ctx, module, resourceType, dest)
}

func (c *Client) CreateResourceWithModule(ctx context.Context, module, resourceType string, item, dest any) error {
	return c.createOrUpdate(ctx, http.MethodPost, module, resourceType, item, dest)
}

func (c *Client) UpdateResourceWithModule(ctx context.Context, module, resourceType string, item, dest any) error {
	return c.createOrUpdate(ctx, http.MethodPut, module, resourceType, item, dest)
}

func (c *Client) DeleteResource(ctx context.Context, resourceType, name, signature string) error {
	return c.DeleteResourceWithModule(ctx, "ignition", resourceType, name, signature)
}

func (c *Client) DeleteResourceWithModule(ctx context.Context, module, resourceType, name, signature string) error {
	path := apiPath("/data/api/v1/resources", module, resourceType, name, signature)
	_, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	return err
}

func (c *Client) unmarshalResourceResponse(ctx context.Context, module, resourceType string, body []byte, dest any) error {
	if len(body) == 0 {
		return fmt.Errorf("empty response body")
	}

	var changes ResourceChangesResponse
	if err := json.Unmarshal(body, &changes); err == nil && len(changes.Changes) > 0 {
		return c.GetResourceWithModule(ctx, module, resourceType, changes.Changes[0].Name, dest)
	}

	if body[0] == '{' {
		return json.Unmarshal(body, dest)
	}

	var rawItems []json.RawMessage
	if err := json.Unmarshal(body, &rawItems); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if len(rawItems) == 0 {
		return errors.New("failed to unmarshal response: no resource in the response")
	}

	return json.Unmarshal(rawItems[0], dest)
}

func getR[T any](ctx context.Context, c *Client, m, t, n string) (*ResourceResponse[T], error) {
	var r ResourceResponse[T]
	err := c.GetResourceWithModule(ctx, m, t, n, &r)
	return &r, err
}

func (c *Client) GetDatabaseConnection(ctx context.Context, n string) (*ResourceResponse[DatabaseConfig], error) {
	return getR[DatabaseConfig](ctx, c, "ignition", "database-connection", n)
}

func (c *Client) CreateDatabaseConnection(ctx context.Context, i ResourceResponse[DatabaseConfig]) (*ResourceResponse[DatabaseConfig], error) {
	var r ResourceResponse[DatabaseConfig]
	err := c.CreateResource(ctx, "database-connection", i, &r)
	return &r, err
}

func (c *Client) UpdateDatabaseConnection(ctx context.Context, i ResourceResponse[DatabaseConfig]) (*ResourceResponse[DatabaseConfig], error) {
	var r ResourceResponse[DatabaseConfig]
	err := c.UpdateResource(ctx, "database-connection", i, &r)
	return &r, err
}

func (c *Client) DeleteDatabaseConnection(ctx context.Context, n, s string) error {
	return c.DeleteResource(ctx, "database-connection", n, s)
}

func (c *Client) GetUserSource(ctx context.Context, n string) (*ResourceResponse[UserSourceConfig], error) {
	return getR[UserSourceConfig](ctx, c, "ignition", "user-source", n)
}

func (c *Client) CreateUserSource(ctx context.Context, i ResourceResponse[UserSourceConfig]) (*ResourceResponse[UserSourceConfig], error) {
	var r ResourceResponse[UserSourceConfig]
	err := c.CreateResource(ctx, "user-source", i, &r)
	return &r, err
}

func (c *Client) UpdateUserSource(ctx context.Context, i ResourceResponse[UserSourceConfig]) (*ResourceResponse[UserSourceConfig], error) {
	var r ResourceResponse[UserSourceConfig]
	err := c.UpdateResource(ctx, "user-source", i, &r)
	return &r, err
}

func (c *Client) DeleteUserSource(ctx context.Context, n, s string) error {
	return c.DeleteResource(ctx, "user-source", n, s)
}

func (c *Client) GetTagProvider(ctx context.Context, n string) (*ResourceResponse[TagProviderConfig], error) {
	return getR[TagProviderConfig](ctx, c, "ignition", "tag-provider", n)
}

func (c *Client) CreateTagProvider(ctx context.Context, i ResourceResponse[TagProviderConfig]) (*ResourceResponse[TagProviderConfig], error) {
	var r ResourceResponse[TagProviderConfig]
	err := c.CreateResource(ctx, "tag-provider", i, &r)
	return &r, err
}

func (c *Client) UpdateTagProvider(ctx context.Context, i ResourceResponse[TagProviderConfig]) (*ResourceResponse[TagProviderConfig], error) {
	var r ResourceResponse[TagProviderConfig]
	err := c.UpdateResource(ctx, "tag-provider", i, &r)
	return &r, err
}

func (c *Client) DeleteTagProvider(ctx context.Context, n, s string) error {
	return c.DeleteResource(ctx, "tag-provider", n, s)
}

func (c *Client) GetAuditProfile(ctx context.Context, n string) (*ResourceResponse[AuditProfileConfig], error) {
	return getR[AuditProfileConfig](ctx, c, "ignition", "audit-profile", n)
}

func (c *Client) CreateAuditProfile(ctx context.Context, i ResourceResponse[AuditProfileConfig]) (*ResourceResponse[AuditProfileConfig], error) {
	var r ResourceResponse[AuditProfileConfig]
	err := c.CreateResource(ctx, "audit-profile", i, &r)
	return &r, err
}

func (c *Client) UpdateAuditProfile(ctx context.Context, i ResourceResponse[AuditProfileConfig]) (*ResourceResponse[AuditProfileConfig], error) {
	var r ResourceResponse[AuditProfileConfig]
	err := c.UpdateResource(ctx, "audit-profile", i, &r)
	return &r, err
}

func (c *Client) DeleteAuditProfile(ctx context.Context, n, s string) error {
	return c.DeleteResource(ctx, "audit-profile", n, s)
}

func (c *Client) GetAlarmNotificationProfile(ctx context.Context, n string) (*ResourceResponse[AlarmNotificationProfileConfig], error) {
	return getR[AlarmNotificationProfileConfig](ctx, c, "com.inductiveautomation.alarm-notification", "alarm-notification-profile", n)
}

func (c *Client) CreateAlarmNotificationProfile(ctx context.Context, i ResourceResponse[AlarmNotificationProfileConfig]) (*ResourceResponse[AlarmNotificationProfileConfig], error) {
	var r ResourceResponse[AlarmNotificationProfileConfig]
	err := c.CreateResourceWithModule(ctx, "com.inductiveautomation.alarm-notification", "alarm-notification-profile", i, &r)
	return &r, err
}

func (c *Client) UpdateAlarmNotificationProfile(ctx context.Context, i ResourceResponse[AlarmNotificationProfileConfig]) (*ResourceResponse[AlarmNotificationProfileConfig], error) {
	var r ResourceResponse[AlarmNotificationProfileConfig]
	err := c.UpdateResourceWithModule(ctx, "com.inductiveautomation.alarm-notification", "alarm-notification-profile", i, &r)
	return &r, err
}

func (c *Client) DeleteAlarmNotificationProfile(ctx context.Context, n, s string) error {
	return c.DeleteResourceWithModule(ctx, "com.inductiveautomation.alarm-notification", "alarm-notification-profile", n, s)
}

func (c *Client) GetOpcUaConnection(ctx context.Context, n string) (*ResourceResponse[OpcUaConnectionConfig], error) {
	return getR[OpcUaConnectionConfig](ctx, c, "ignition", "opc-connection", n)
}

func (c *Client) CreateOpcUaConnection(ctx context.Context, i ResourceResponse[OpcUaConnectionConfig]) (*ResourceResponse[OpcUaConnectionConfig], error) {
	var r ResourceResponse[OpcUaConnectionConfig]
	err := c.CreateResourceWithModule(ctx, "ignition", "opc-connection", i, &r)
	return &r, err
}

func (c *Client) UpdateOpcUaConnection(ctx context.Context, i ResourceResponse[OpcUaConnectionConfig]) (*ResourceResponse[OpcUaConnectionConfig], error) {
	var r ResourceResponse[OpcUaConnectionConfig]
	err := c.UpdateResourceWithModule(ctx, "ignition", "opc-connection", i, &r)
	return &r, err
}

func (c *Client) DeleteOpcUaConnection(ctx context.Context, n, s string) error {
	return c.DeleteResourceWithModule(ctx, "ignition", "opc-connection", n, s)
}

func (c *Client) GetAlarmJournal(ctx context.Context, n string) (*ResourceResponse[AlarmJournalConfig], error) {
	return getR[AlarmJournalConfig](ctx, c, "ignition", "alarm-journal", n)
}

func (c *Client) CreateAlarmJournal(ctx context.Context, i ResourceResponse[AlarmJournalConfig]) (*ResourceResponse[AlarmJournalConfig], error) {
	var r ResourceResponse[AlarmJournalConfig]
	err := c.CreateResourceWithModule(ctx, "ignition", "alarm-journal", i, &r)
	return &r, err
}

func (c *Client) UpdateAlarmJournal(ctx context.Context, i ResourceResponse[AlarmJournalConfig]) (*ResourceResponse[AlarmJournalConfig], error) {
	var r ResourceResponse[AlarmJournalConfig]
	err := c.UpdateResourceWithModule(ctx, "ignition", "alarm-journal", i, &r)
	return &r, err
}

func (c *Client) DeleteAlarmJournal(ctx context.Context, n, s string) error {
	return c.DeleteResourceWithModule(ctx, "ignition", "alarm-journal", n, s)
}

func (c *Client) GetSMTPProfile(ctx context.Context, n string) (*ResourceResponse[SMTPProfileConfig], error) {
	return getR[SMTPProfileConfig](ctx, c, "ignition", "email-profile", n)
}

func (c *Client) CreateSMTPProfile(ctx context.Context, i ResourceResponse[SMTPProfileConfig]) (*ResourceResponse[SMTPProfileConfig], error) {
	var r ResourceResponse[SMTPProfileConfig]
	err := c.CreateResource(ctx, "email-profile", i, &r)
	return &r, err
}

func (c *Client) UpdateSMTPProfile(ctx context.Context, i ResourceResponse[SMTPProfileConfig]) (*ResourceResponse[SMTPProfileConfig], error) {
	var r ResourceResponse[SMTPProfileConfig]
	err := c.UpdateResource(ctx, "email-profile", i, &r)
	return &r, err
}

func (c *Client) DeleteSMTPProfile(ctx context.Context, n, s string) error {
	return c.DeleteResource(ctx, "email-profile", n, s)
}

func (c *Client) GetStoreAndForward(ctx context.Context, n string) (*ResourceResponse[StoreAndForwardConfig], error) {
	return getR[StoreAndForwardConfig](ctx, c, "ignition", "store-and-forward-engine", n)
}

func (c *Client) CreateStoreAndForward(ctx context.Context, i ResourceResponse[StoreAndForwardConfig]) (*ResourceResponse[StoreAndForwardConfig], error) {
	var r ResourceResponse[StoreAndForwardConfig]
	err := c.CreateResource(ctx, "store-and-forward-engine", i, &r)
	return &r, err
}

func (c *Client) UpdateStoreAndForward(ctx context.Context, i ResourceResponse[StoreAndForwardConfig]) (*ResourceResponse[StoreAndForwardConfig], error) {
	var r ResourceResponse[StoreAndForwardConfig]
	err := c.UpdateResource(ctx, "store-and-forward-engine", i, &r)
	return &r, err
}

func (c *Client) DeleteStoreAndForward(ctx context.Context, n, s string) error {
	return c.DeleteResource(ctx, "store-and-forward-engine", n, s)
}

func (c *Client) GetIdentityProvider(ctx context.Context, n string) (*ResourceResponse[IdentityProviderConfig], error) {
	return getR[IdentityProviderConfig](ctx, c, "ignition", "identity-provider", n)
}

func (c *Client) CreateIdentityProvider(ctx context.Context, i ResourceResponse[IdentityProviderConfig]) (*ResourceResponse[IdentityProviderConfig], error) {
	var r ResourceResponse[IdentityProviderConfig]
	err := c.CreateResource(ctx, "identity-provider", i, &r)
	return &r, err
}

func (c *Client) UpdateIdentityProvider(ctx context.Context, i ResourceResponse[IdentityProviderConfig]) (*ResourceResponse[IdentityProviderConfig], error) {
	var r ResourceResponse[IdentityProviderConfig]
	err := c.UpdateResource(ctx, "identity-provider", i, &r)
	return &r, err
}

func (c *Client) DeleteIdentityProvider(ctx context.Context, n, s string) error {
	return c.DeleteResource(ctx, "identity-provider", n, s)
}

func (c *Client) GetGanOutgoing(ctx context.Context, n string) (*ResourceResponse[GanOutgoingConfig], error) {
	return getR[GanOutgoingConfig](ctx, c, "ignition", "gateway-network-outgoing", n)
}

func (c *Client) CreateGanOutgoing(ctx context.Context, i ResourceResponse[GanOutgoingConfig]) (*ResourceResponse[GanOutgoingConfig], error) {
	var r ResourceResponse[GanOutgoingConfig]
	err := c.CreateResource(ctx, "gateway-network-outgoing", i, &r)
	return &r, err
}

func (c *Client) UpdateGanOutgoing(ctx context.Context, i ResourceResponse[GanOutgoingConfig]) (*ResourceResponse[GanOutgoingConfig], error) {
	var r ResourceResponse[GanOutgoingConfig]
	err := c.UpdateResource(ctx, "gateway-network-outgoing", i, &r)
	return &r, err
}

func (c *Client) DeleteGanOutgoing(ctx context.Context, n, s string) error {
	return c.DeleteResource(ctx, "gateway-network-outgoing", n, s)
}

func (c *Client) GetRedundancyConfig(ctx context.Context) (*RedundancyConfig, error) {
	body, err := c.doRequest(ctx, http.MethodGet, "/data/api/v1/redundancy/config", nil)
	if err != nil {
		return nil, err
	}
	var config RedundancyConfig
	return &config, json.Unmarshal(body, &config)
}

func (c *Client) UpdateRedundancyConfig(ctx context.Context, config RedundancyConfig) error {
	rb, err := json.Marshal(config)
	if err != nil {
		return err
	}
	_, err = c.doRequest(ctx, http.MethodPut, "/data/api/v1/redundancy/config", rb)
	return err
}

// Singleton resources have no name of their own, so the gateway never
// includes one in singleton GET responses; fill in the fixed logical name
// the rest of the provider expects to see.
const ganGeneralSettingsName = "gateway-network-settings"

func (c *Client) GetGanGeneralSettings(ctx context.Context) (*ResourceResponse[GanGeneralSettingsConfig], error) {
	var r ResourceResponse[GanGeneralSettingsConfig]
	err := c.GetResourceSingleton(ctx, "ignition", ganGeneralSettingsName, &r)
	if err != nil {
		return nil, err
	}
	r.Name = ganGeneralSettingsName
	return &r, nil
}

func (c *Client) UpdateGanGeneralSettings(ctx context.Context, i ResourceResponse[GanGeneralSettingsConfig]) (*ResourceResponse[GanGeneralSettingsConfig], error) {
	var r ResourceResponse[GanGeneralSettingsConfig]
	err := c.updateSingleton(ctx, "ignition", ganGeneralSettingsName, i, &r)
	if err != nil {
		return nil, err
	}
	r.Name = ganGeneralSettingsName
	return &r, nil
}

// deviceFromWire and deviceToWire translate between the flat DeviceConfig the
// rest of the provider works with and the profile/settings envelope the
// gateway actually requires on the wire.
func deviceFromWire(r ResourceResponse[DeviceWireConfig]) *ResourceResponse[DeviceConfig] {
	return &ResourceResponse[DeviceConfig]{
		Module: r.Module,
		// The outer "type" is the generic resource-type identifier
		// (e.g. "com.inductiveautomation.opcua/device"); the actual
		// driver flavor lives in config.profile.type.
		Type:        r.Config.Profile.Type,
		Name:        r.Name,
		Enabled:     r.Enabled,
		Description: r.Description,
		Signature:   r.Signature,
		Config:      r.Config.Settings,
	}
}

func deviceToWire(i ResourceResponse[DeviceConfig]) ResourceResponse[DeviceWireConfig] {
	return ResourceResponse[DeviceWireConfig]{
		Module:      i.Module,
		Type:        i.Type,
		Name:        i.Name,
		Enabled:     i.Enabled,
		Description: i.Description,
		Signature:   i.Signature,
		Config: DeviceWireConfig{
			Profile:  DeviceWireProfile{Type: i.Type},
			Settings: i.Config,
		},
	}
}

func (c *Client) GetDevice(ctx context.Context, n string) (*ResourceResponse[DeviceConfig], error) {
	r, err := getR[DeviceWireConfig](ctx, c, "com.inductiveautomation.opcua", "device", n)
	if err != nil {
		return nil, err
	}
	return deviceFromWire(*r), nil
}

func (c *Client) CreateDevice(ctx context.Context, i ResourceResponse[DeviceConfig]) (*ResourceResponse[DeviceConfig], error) {
	var r ResourceResponse[DeviceWireConfig]
	err := c.CreateResourceWithModule(ctx, "com.inductiveautomation.opcua", "device", deviceToWire(i), &r)
	return deviceFromWire(r), err
}

func (c *Client) UpdateDevice(ctx context.Context, i ResourceResponse[DeviceConfig]) (*ResourceResponse[DeviceConfig], error) {
	var r ResourceResponse[DeviceWireConfig]
	err := c.UpdateResourceWithModule(ctx, "com.inductiveautomation.opcua", "device", deviceToWire(i), &r)
	return deviceFromWire(r), err
}

func (c *Client) DeleteDevice(ctx context.Context, n, s string) error {
	return c.DeleteResourceWithModule(ctx, "com.inductiveautomation.opcua", "device", n, s)
}

func (c *Client) GetProject(ctx context.Context, name string) (*Project, error) {
	body, err := c.doRequest(ctx, http.MethodGet, apiPath("/data/api/v1/projects/find", name), nil)
	if err != nil {
		return nil, err
	}
	var p Project
	return &p, json.Unmarshal(body, &p)
}

func (c *Client) CreateProject(ctx context.Context, p Project) (*Project, error) {
	rb, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	if _, err := c.doRequest(ctx, http.MethodPost, "/data/api/v1/projects", rb); err != nil {
		return nil, err
	}
	return c.waitForProject(ctx, p)
}

func (c *Client) UpdateProject(ctx context.Context, p Project) (*Project, error) {
	rb, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	if _, err := c.doRequest(ctx, http.MethodPut, apiPath("/data/api/v1/projects", p.Name), rb); err != nil {
		return nil, err
	}
	return c.waitForProject(ctx, p)
}

func (c *Client) DeleteProject(ctx context.Context, name string) error {
	_, err := c.doRequest(ctx, http.MethodDelete, apiPath("/data/api/v1/projects", name), nil)
	return err
}

// waitForProject polls until the gateway shows the project with every field
// that was sent, so an update doesn't return the project as it was before.
func (c *Client) waitForProject(ctx context.Context, want Project) (*Project, error) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.After(10 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return c.GetProject(ctx, want.Name)
		case <-ticker.C:
			if p, err := c.GetProject(ctx, want.Name); err == nil && projectApplied(*p, want) {
				return p, nil
			}
		}
	}
}

// projectApplied reports whether got holds every field sent in want: the
// request omits empty strings and false, so only those that were set count.
func projectApplied(got, want Project) bool {
	same := func(g, w string) bool { return w == "" || g == w }
	return got.Name == want.Name &&
		same(got.Description, want.Description) && same(got.Title, want.Title) &&
		same(got.Parent, want.Parent) && same(got.DefaultDB, want.DefaultDB) &&
		same(got.TagProvider, want.TagProvider) && same(got.UserSource, want.UserSource) &&
		same(got.IdentityProvider, want.IdentityProvider) &&
		(!want.Enabled || got.Enabled) && (!want.Inheritable || got.Inheritable)
}
