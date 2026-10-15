package client

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestClient_RetryOn503(t *testing.T) {
	var reqCount int32

	// Mock Server that fails 3 times with 503 then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&reqCount, 1)
		if count <= 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("Service Restarting"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name": "success", "enabled": true, "config": {}}`))
	}))
	defer server.Close()

	// Client with short retry wait for testing speed
	c, err := NewClient(server.URL, "test-token", false)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	// Override backoff to speed up test
	c.HTTPClient.RetryWaitMin = 10 * time.Millisecond
	c.HTTPClient.RetryWaitMax = 50 * time.Millisecond

	// Test
	var dest ResourceResponse[map[string]any]
	err = c.GetResource(context.Background(), "test-type", "test-name", &dest)
	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}

	if atomic.LoadInt32(&reqCount) != 4 {
		t.Errorf("Expected 4 requests (3 retries), got %d", reqCount)
	}
}

func TestClient_RetryOn429(t *testing.T) {
	var reqCount int32

	// Mock Server that fails 2 times with 429 then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&reqCount, 1)
		if count <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name": "success", "enabled": true, "config": {}}`))
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "test-token", false)
	c.HTTPClient.RetryWaitMin = 10 * time.Millisecond
	c.HTTPClient.RetryWaitMax = 50 * time.Millisecond

	var dest ResourceResponse[map[string]any]
	err := c.GetResource(context.Background(), "test-type", "test-name", &dest)
	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}

	if atomic.LoadInt32(&reqCount) != 3 {
		t.Errorf("Expected 3 requests, got %d", reqCount)
	}
}

func TestClient_NoRetryOn400(t *testing.T) {
	var reqCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		w.WriteHeader(http.StatusBadRequest) // 400 should not retry
		_, _ = w.Write([]byte(`{"success": false, "messages": ["Bad Request"]}`))
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "test-token", false)

	var dest ResourceResponse[map[string]any]
	err := c.GetResource(context.Background(), "test-type", "test-name", &dest)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if atomic.LoadInt32(&reqCount) != 1 {
		t.Errorf("Expected 1 request, got %d", reqCount)
	}
}

func TestClient_WaitLogic_SimulatedRestart(t *testing.T) {
	// Verify that we handle a "Wait" scenario.
	// This simulates a resource creation where the subsequent GET fails initially (e.g. restart)
	// but eventually succeeds.

	var postCount int32
	var getCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			atomic.AddInt32(&postCount, 1)
			// Create returns 200 OK immediately with a changeset
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"success": true,
				"changes": [{"name": "new-resource", "type": "test", "collection": "tests"}]
			}`))
			return
		}

		if r.Method == http.MethodGet {
			count := atomic.AddInt32(&getCount, 1)
			if count <= 3 {
				// Simulate Gateway restarting / API unavailable after config change
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"name": "new-resource", "config": {}}`))
			return
		}
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "test-token", false)
	c.HTTPClient.RetryWaitMin = 10 * time.Millisecond
	c.HTTPClient.RetryWaitMax = 50 * time.Millisecond

	// We use CreateResource which internally calls unmarshalResourceResponse -> GetResourceWithModule
	// The GetResourceWithModule should retry on the 503s
	item := map[string]string{"foo": "bar"}
	var dest ResourceResponse[map[string]any]

	err := c.CreateResource(context.Background(), "test", item, &dest)
	if err != nil {
		t.Fatalf("Expected CreateResource to handle transient get errors, got: %v", err)
	}

	if atomic.LoadInt32(&postCount) != 1 {
		t.Errorf("Expected 1 POST, got %d", postCount)
	}
	if atomic.LoadInt32(&getCount) != 4 {
		t.Errorf("Expected 4 GETs (3 fails + 1 success), got %d", getCount)
	}
}

func TestClient_CreateNotResentAfterTimeout(t *testing.T) {
	var posts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&posts, 1) == 1 {
			// the gateway creates the project, but answers after the client gave up
			time.Sleep(300 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"messages": ["already exists"]}`))
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "test-token", false)
	c.HTTPClient.HTTPClient.Timeout = 100 * time.Millisecond
	c.HTTPClient.RetryWaitMin = 10 * time.Millisecond
	c.HTTPClient.RetryWaitMax = 50 * time.Millisecond

	_, err := c.CreateProject(context.Background(), Project{Name: "slow"})
	if err == nil {
		t.Fatal("Expected the timeout to be reported")
	}
	if strings.Contains(err.Error(), "already exists") {
		t.Errorf("Expected the timeout, not a resent create: %v", err)
	}
	if n := atomic.LoadInt32(&posts); n != 1 {
		t.Errorf("Expected 1 POST, got %d", n)
	}
}

func TestClient_CreateNotResentAfter500(t *testing.T) {
	var posts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&posts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "test-token", false)
	c.HTTPClient.RetryWaitMin = 10 * time.Millisecond
	c.HTTPClient.RetryWaitMax = 50 * time.Millisecond

	var dest map[string]any
	if err := c.CreateResource(context.Background(), "t", map[string]any{}, &dest); err == nil {
		t.Fatal("Expected an error")
	}
	if n := atomic.LoadInt32(&posts); n != 1 {
		t.Errorf("Expected 1 POST, got %d", n)
	}
}

func TestClient_CreateRetriedOn503(t *testing.T) {
	var posts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&posts, 1) <= 2 {
			// the gateway is restarting and acted on nothing
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`[{"name": "x", "config": {}}]`))
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "test-token", false)
	c.HTTPClient.RetryWaitMin = 10 * time.Millisecond
	c.HTTPClient.RetryWaitMax = 50 * time.Millisecond

	var dest ResourceResponse[map[string]any]
	if err := c.CreateResource(context.Background(), "t", map[string]any{}, &dest); err != nil {
		t.Fatalf("Expected success after retries, got %v", err)
	}
	if n := atomic.LoadInt32(&posts); n != 3 {
		t.Errorf("Expected 3 POSTs, got %d", n)
	}
}

func TestRetryPolicy_ResendsOnlyWhatNeverArrived(t *testing.T) {
	once := context.WithValue(context.Background(), noResendKey{}, true)
	refused := &url.Error{Op: "Post", URL: "http://gw", Err: &net.OpError{Op: "dial", Err: errors.New("connection refused")}}
	reset := &url.Error{Op: "Post", URL: "http://gw", Err: &net.OpError{Op: "read", Err: errors.New("connection reset")}}
	cases := []struct {
		name string
		ctx  context.Context
		resp *http.Response
		err  error
		want bool
	}{
		{"create refused", once, nil, refused, true},
		{"create reset after sending", once, nil, reset, false},
		{"create timed out", once, nil, context.DeadlineExceeded, false},
		{"create 502", once, &http.Response{StatusCode: http.StatusBadGateway}, nil, false},
		{"create 429", once, &http.Response{StatusCode: http.StatusTooManyRequests}, nil, true},
		{"read reset", context.Background(), nil, reset, true},
		{"read 502", context.Background(), &http.Response{StatusCode: http.StatusBadGateway}, nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got, _ := retryPolicy(tc.ctx, tc.resp, tc.err); got != tc.want {
				t.Errorf("retry = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestClient_UpdateProjectWaitsForTheChange(t *testing.T) {
	var gets int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			return
		}
		// the first reads still show the project as it was
		desc := "old"
		if atomic.AddInt32(&gets, 1) > 2 {
			desc = "new"
		}
		_, _ = w.Write([]byte(`{"name": "p", "description": "` + desc + `"}`))
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "test-token", false)
	p, err := c.UpdateProject(context.Background(), Project{Name: "p", Description: "new"})
	if err != nil {
		t.Fatalf("UpdateProject failed: %v", err)
	}
	if p.Description != "new" {
		t.Errorf("Expected the updated project, got description %q", p.Description)
	}
}

func TestNewClient_TrimsTrailingSlash(t *testing.T) {
	var path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_, _ = w.Write([]byte(`{"name": "p"}`))
	}))
	defer server.Close()

	c, _ := NewClient(server.URL+"/", "test-token", false)
	if _, err := c.GetProject(context.Background(), "p"); err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}
	if path != "/data/api/v1/projects/find/p" {
		t.Errorf("Expected a single slash after the host, got %q", path)
	}
}

func TestNewClient_InsecureTLSKeepsTransportDefaults(t *testing.T) {
	c, _ := NewClient("https://gw", "test-token", true)
	tr, ok := c.HTTPClient.HTTPClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Expected *http.Transport, got %T", c.HTTPClient.HTTPClient.Transport)
	}
	if !tr.TLSClientConfig.InsecureSkipVerify {
		t.Error("Expected certificate checks off")
	}
	if tr.Proxy == nil || tr.TLSHandshakeTimeout == 0 {
		t.Error("Expected the proxy setting and handshake timeout kept")
	}
}

// A signed update the gateway applied but answered late isn't resent: the resend would carry
// the signature the update just replaced.
func TestClient_SignedUpdateNotResentAfterTimeout(t *testing.T) {
	for _, tc := range []struct {
		name   string
		update func(c *Client) error
	}{
		{"resource", func(c *Client) error {
			var dest ResourceResponse[map[string]any]
			return c.UpdateResource(context.Background(), "t", ResourceResponse[map[string]any]{Name: "x", Signature: "old"}, &dest)
		}},
		{"singleton", func(c *Client) error {
			_, err := c.UpdateGanGeneralSettings(context.Background(), ResourceResponse[GanGeneralSettingsConfig]{Signature: "old"})
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var puts int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					_, _ = w.Write([]byte(`{"name": "x", "signature": "new", "config": {}}`))
					return
				}
				if atomic.AddInt32(&puts, 1) == 1 {
					time.Sleep(300 * time.Millisecond)
					return
				}
				w.WriteHeader(http.StatusConflict)
				_, _ = w.Write([]byte(`{"messages": ["signature mismatch"]}`))
			}))
			defer server.Close()

			c, _ := NewClient(server.URL, "test-token", false)
			c.HTTPClient.HTTPClient.Timeout = 100 * time.Millisecond
			c.HTTPClient.RetryWaitMin = 10 * time.Millisecond
			c.HTTPClient.RetryWaitMax = 50 * time.Millisecond

			err := tc.update(c)
			if err == nil || strings.Contains(err.Error(), "signature mismatch") {
				t.Errorf("Expected the timeout, not a resent update: %v", err)
			}
			if n := atomic.LoadInt32(&puts); n != 1 {
				t.Errorf("Expected 1 PUT, got %d", n)
			}
		})
	}
}

// Updates that carry no signature (a project, the redundancy settings) are idempotent, so a
// gateway restarting mid-update is still retried.
func TestClient_UnsignedUpdateRetriedOn500(t *testing.T) {
	var puts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && atomic.AddInt32(&puts, 1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "test-token", false)
	c.HTTPClient.RetryWaitMin = 10 * time.Millisecond
	c.HTTPClient.RetryWaitMax = 50 * time.Millisecond
	if err := c.UpdateRedundancyConfig(context.Background(), RedundancyConfig{}); err != nil {
		t.Fatalf("Expected success after a retry, got %v", err)
	}
	if n := atomic.LoadInt32(&puts); n != 2 {
		t.Errorf("Expected 2 PUTs, got %d", n)
	}
}

// A delete is still retried: if the first one was applied, the retry finds nothing, and
// ErrNotFound is what the resources treat as already destroyed.
func TestClient_RetriedDeleteOfAnAppliedDeleteIsNotFound(t *testing.T) {
	var deletes int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&deletes, 1) == 1 {
			w.WriteHeader(http.StatusBadGateway) // applied, but the answer was lost
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "test-token", false)
	c.HTTPClient.RetryWaitMin = 10 * time.Millisecond
	c.HTTPClient.RetryWaitMax = 50 * time.Millisecond
	if err := c.DeleteResource(context.Background(), "t", "x", "sig"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

// A gateway that keeps failing is given up on after RequestTimeout, not after every retry
// and its backoff (minutes with the defaults), and the error says what it last answered.
func TestClient_RequestTimeoutBoundsRetries(t *testing.T) {
	var reqCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "test-token", false)
	if c.RequestTimeout != DefaultRequestTimeout {
		t.Fatalf("Expected the default request timeout, got %s", c.RequestTimeout)
	}
	c.RequestTimeout = 300 * time.Millisecond
	c.HTTPClient.RetryWaitMin = 100 * time.Millisecond
	c.HTTPClient.RetryWaitMax = 100 * time.Millisecond

	start := time.Now()
	var dest ResourceResponse[map[string]any]
	err := c.GetResource(context.Background(), "test-type", "test-name", &dest)
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("Expected to give up after about 300ms, took %s", elapsed)
	}
	if err == nil || !strings.Contains(err.Error(), "request_timeout") || !strings.Contains(err.Error(), "status 502") {
		t.Errorf("Expected a request_timeout error naming the last status, got %v", err)
	}
	if n := atomic.LoadInt32(&reqCount); n < 2 || n > 5 {
		t.Errorf("Expected a few attempts within the timeout, got %d", n)
	}
}

// A caller's own cancellation is reported as such, not as the request timeout.
func TestClient_RequestTimeoutLeavesCallerCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	c, _ := NewClient(server.URL, "test-token", false)
	c.HTTPClient.RetryWaitMin = 50 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	var dest ResourceResponse[map[string]any]
	err := c.GetResource(ctx, "test-type", "test-name", &dest)
	if !errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "request_timeout") {
		t.Errorf("Expected the caller's deadline, got %v", err)
	}
}
