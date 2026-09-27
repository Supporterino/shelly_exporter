package client

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/icholy/digest"
)

type testResponse struct {
	Value string `json:"value"`
}

func hostOf(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	return strings.TrimPrefix(srv.URL, "http://")
}

func TestFetchDataSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"value":"ok"}`))
	}))
	defer srv.Close()

	c := NewAPIClient(hostOf(t, srv))
	var result testResponse
	if err := c.FetchData(context.Background(), "/rpc/Test", &result); err != nil {
		t.Fatalf("FetchData returned error: %v", err)
	}
	if result.Value != "ok" {
		t.Errorf("Value = %q, want ok", result.Value)
	}
}

func TestFetchDataNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewAPIClient(hostOf(t, srv))
	if err := c.FetchData(context.Background(), "/rpc/Test", &testResponse{}); err == nil {
		t.Fatal("expected non-200 status to return an error")
	}
}

func TestFetchDataMalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"value":`))
	}))
	defer srv.Close()

	c := NewAPIClient(hostOf(t, srv))
	if err := c.FetchData(context.Background(), "/rpc/Test", &testResponse{}); err == nil {
		t.Fatal("expected malformed JSON to return an error")
	}
}

func TestFetchDataTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewAPIClient(hostOf(t, srv), WithTimeout(20*time.Millisecond))
	if err := c.FetchData(context.Background(), "/rpc/Test", &testResponse{}); err == nil {
		t.Fatal("expected timeout to return an error")
	}
}

func TestFetchDataContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	c := NewAPIClient(hostOf(t, srv))
	if err := c.FetchData(ctx, "/rpc/Test", &testResponse{}); err == nil {
		t.Fatal("expected cancelled context to return an error")
	}
}

type recordingTransport struct {
	mu    sync.Mutex
	calls int
}

func (t *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	t.calls++
	t.mu.Unlock()
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("{}")),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func TestInjectableHTTPClient(t *testing.T) {
	transport := &recordingTransport{}
	c := NewAPIClient("example.invalid", WithHTTPClient(&http.Client{Transport: transport}))

	if err := c.FetchData(context.Background(), "/rpc/Test", &testResponse{}); err != nil {
		t.Fatalf("FetchData returned error: %v", err)
	}

	transport.mu.Lock()
	defer transport.mu.Unlock()
	if transport.calls != 1 {
		t.Errorf("injected transport calls = %d, want 1", transport.calls)
	}
}

func TestNoCredentialsSendsNoAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewAPIClient(hostOf(t, srv))
	if err := c.FetchData(context.Background(), "/rpc/Test", &testResponse{}); err != nil {
		t.Fatalf("FetchData returned error: %v", err)
	}
	if gotAuth != "" {
		t.Errorf("Authorization header = %q, want empty", gotAuth)
	}
}

func TestBasicAuthChallenge(t *testing.T) {
	var unauthenticated, authenticated int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "user" || pass != "pass" {
			unauthenticated++
			w.Header().Set("WWW-Authenticate", `Basic realm="test"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		authenticated++
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewAPIClient(hostOf(t, srv), WithCredentials("user", "pass"))
	if err := c.FetchData(context.Background(), "/rpc/Test", &testResponse{}); err != nil {
		t.Fatalf("FetchData returned error: %v", err)
	}
	if unauthenticated != 1 || authenticated != 1 {
		t.Errorf("unauthenticated=%d authenticated=%d, want 1 and 1", unauthenticated, authenticated)
	}
}

func TestBasicSchemeReused(t *testing.T) {
	var unauthenticated, authenticated int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := r.BasicAuth(); !ok {
			unauthenticated++
			w.Header().Set("WWW-Authenticate", `Basic realm="test"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		authenticated++
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewAPIClient(hostOf(t, srv), WithCredentials("user", "pass"))
	for i := 0; i < 3; i++ {
		if err := c.FetchData(context.Background(), "/rpc/Test", &testResponse{}); err != nil {
			t.Fatalf("FetchData returned error: %v", err)
		}
	}
	if unauthenticated != 1 {
		t.Errorf("unauthenticated probes = %d, want 1", unauthenticated)
	}
	if authenticated != 3 {
		t.Errorf("authenticated requests = %d, want 3", authenticated)
	}
}

func digestChallengeServer(t *testing.T, algorithm string) (*httptest.Server, func() int) {
	t.Helper()

	challenge := &digest.Challenge{
		Realm:     "test",
		Nonce:     "abc123",
		Algorithm: algorithm,
		QOP:       []string{"auth"},
	}

	authCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !digest.IsDigest(header) {
			w.Header().Set("WWW-Authenticate", challenge.String())
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		cred, err := digest.ParseCredentials(header)
		if err != nil {
			t.Errorf("failed to parse credentials: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		expected, err := digest.Digest(challenge, digest.Options{
			Method:   http.MethodGet,
			URI:      r.URL.RequestURI(),
			Count:    cred.Nc,
			Username: "user",
			Password: "pass",
			Cnonce:   cred.Cnonce,
		})
		if err != nil {
			t.Errorf("failed to compute expected digest: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if cred.Username != "user" || cred.Response != expected.Response {
			t.Errorf("digest credentials mismatch: got %+v want response %s", cred, expected.Response)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		authCount++
		_, _ = w.Write([]byte(`{"value":"ok"}`))
	}))
	t.Cleanup(srv.Close)

	return srv, func() int { return authCount }
}

func TestDigestAuthAlgorithms(t *testing.T) {
	for _, algorithm := range []string{"MD5", "SHA-256"} {
		t.Run(algorithm, func(t *testing.T) {
			srv, authCount := digestChallengeServer(t, algorithm)

			c := NewAPIClient(hostOf(t, srv), WithCredentials("user", "pass"))
			var result testResponse
			if err := c.FetchData(context.Background(), "/rpc/Test", &result); err != nil {
				t.Fatalf("FetchData returned error: %v", err)
			}
			if result.Value != "ok" {
				t.Errorf("Value = %q, want ok", result.Value)
			}
			if got := authCount(); got != 1 {
				t.Errorf("authenticated requests = %d, want 1", got)
			}
		})
	}
}

func TestDigestSchemeReused(t *testing.T) {
	srv, authCount := digestChallengeServer(t, "MD5")

	c := NewAPIClient(hostOf(t, srv), WithCredentials("user", "pass"))
	for i := 0; i < 2; i++ {
		if err := c.FetchData(context.Background(), "/rpc/Test", &testResponse{}); err != nil {
			t.Fatalf("FetchData returned error: %v", err)
		}
	}
	if got := authCount(); got != 2 {
		t.Errorf("authenticated requests = %d, want 2", got)
	}
}

func TestDigestNonceRotationTriggersRenegotiation(t *testing.T) {
	var mu sync.Mutex
	nonce := "nonce-1"
	ncByNonce := map[string]int{}
	authCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		current := nonce
		mu.Unlock()

		challenge := &digest.Challenge{
			Realm:     "test",
			Nonce:     current,
			Algorithm: "MD5",
			QOP:       []string{"auth"},
		}

		creds, err := digest.ParseCredentials(r.Header.Get("Authorization"))
		if err == nil {
			mu.Lock()
			ncByNonce[creds.Nonce] = creds.Nc
			mu.Unlock()
		}
		if err != nil || creds.Nonce != current {
			w.Header().Set("WWW-Authenticate", challenge.String())
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		expected, err := digest.Digest(challenge, digest.Options{
			Method:   http.MethodGet,
			URI:      r.URL.RequestURI(),
			Count:    creds.Nc,
			Username: "user",
			Password: "pass",
			Cnonce:   creds.Cnonce,
		})
		if err != nil {
			t.Errorf("failed to compute expected digest: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if creds.Username != "user" || creds.Response != expected.Response {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		mu.Lock()
		authCount++
		mu.Unlock()
		_, _ = w.Write([]byte(`{"value":"ok"}`))
	}))
	defer srv.Close()

	c := NewAPIClient(hostOf(t, srv), WithCredentials("user", "pass"))
	if err := c.FetchData(context.Background(), "/rpc/Test", &testResponse{}); err != nil {
		t.Fatalf("first FetchData returned error: %v", err)
	}

	mu.Lock()
	nonce = "nonce-2"
	mu.Unlock()

	var result testResponse
	if err := c.FetchData(context.Background(), "/rpc/Test", &result); err != nil {
		t.Fatalf("FetchData after nonce rotation returned error: %v", err)
	}
	if result.Value != "ok" {
		t.Errorf("Value = %q, want ok", result.Value)
	}

	mu.Lock()
	defer mu.Unlock()
	if authCount != 2 {
		t.Errorf("authenticated requests = %d, want 2", authCount)
	}
	if got := ncByNonce["nonce-2"]; got != 1 {
		t.Errorf("nonce-count for the rotated nonce = %d, want 1", got)
	}
}

func TestCachedSchemeRenegotiatesAfterAuthModeChange(t *testing.T) {
	var mu sync.Mutex
	digestMode := false
	authCount := 0

	challenge := &digest.Challenge{
		Realm:     "test",
		Nonce:     "abc123",
		Algorithm: "MD5",
		QOP:       []string{"auth"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		mode := digestMode
		mu.Unlock()

		if !mode {
			if user, pass, ok := r.BasicAuth(); ok && user == "user" && pass == "pass" {
				mu.Lock()
				authCount++
				mu.Unlock()
				_, _ = w.Write([]byte(`{"value":"ok"}`))
				return
			}
			w.Header().Set("WWW-Authenticate", `Basic realm="test"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		creds, err := digest.ParseCredentials(r.Header.Get("Authorization"))
		if err != nil || creds.Nonce != challenge.Nonce {
			w.Header().Set("WWW-Authenticate", challenge.String())
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		expected, err := digest.Digest(challenge, digest.Options{
			Method:   http.MethodGet,
			URI:      r.URL.RequestURI(),
			Count:    creds.Nc,
			Username: "user",
			Password: "pass",
			Cnonce:   creds.Cnonce,
		})
		if err != nil {
			t.Errorf("failed to compute expected digest: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if creds.Username != "user" || creds.Response != expected.Response {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		mu.Lock()
		authCount++
		mu.Unlock()
		_, _ = w.Write([]byte(`{"value":"ok"}`))
	}))
	defer srv.Close()

	c := NewAPIClient(hostOf(t, srv), WithCredentials("user", "pass"))
	if err := c.FetchData(context.Background(), "/rpc/Test", &testResponse{}); err != nil {
		t.Fatalf("basic FetchData returned error: %v", err)
	}

	mu.Lock()
	digestMode = true
	mu.Unlock()

	var result testResponse
	if err := c.FetchData(context.Background(), "/rpc/Test", &result); err != nil {
		t.Fatalf("FetchData after scheme change returned error: %v", err)
	}
	if result.Value != "ok" {
		t.Errorf("Value = %q, want ok", result.Value)
	}

	mu.Lock()
	defer mu.Unlock()
	if authCount != 2 {
		t.Errorf("authenticated requests = %d, want 2", authCount)
	}
}

func TestUnsatisfiableChallenge(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Digest realm="test", nonce="abc", algorithm=SHA-1`)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := NewAPIClient(hostOf(t, srv), WithCredentials("user", "pass"))
	if err := c.FetchData(context.Background(), "/rpc/Test", &testResponse{}); err == nil {
		t.Fatal("expected unsatisfiable challenge to return an error")
	}
}

func TestCredentialsNeverLogged(t *testing.T) {
	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	defer slog.SetDefault(previous)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm="test"`)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := NewAPIClient(hostOf(t, srv), WithCredentials("secret-user", "secret-pass"))
	_ = c.FetchData(context.Background(), "/rpc/Test", &testResponse{})

	logs := buf.String()
	if strings.Contains(logs, "secret-user") || strings.Contains(logs, "secret-pass") {
		t.Errorf("credentials appeared in logs: %s", logs)
	}
}
