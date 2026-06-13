package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// parseDigestHeader extracts the fields of a Digest Authorization header using
// the same regex the client uses to parse challenges.
func parseDigestHeader(h string) map[string]string {
	out := map[string]string{}
	for _, m := range digestFieldRe.FindAllStringSubmatch(h, -1) {
		v := m[2]
		if v == "" {
			v = m[3]
		}
		out[strings.ToLower(m[1])] = v
	}
	return out
}

// validDigest independently recomputes the RFC digest response from the
// header the client sent and reports whether it matches (i.e. the client
// authenticated correctly).
func validDigest(header, method, realm, nonce, user, pass string) bool {
	f := parseDigestHeader(header)
	if f["username"] != user || f["realm"] != realm || f["nonce"] != nonce || f["qop"] != "auth" {
		return false
	}
	ha1 := sha256hex(user + ":" + realm + ":" + pass)
	ha2 := sha256hex(method + ":" + f["uri"])
	want := sha256hex(strings.Join([]string{ha1, nonce, f["nc"], f["cnonce"], "auth", ha2}, ":"))
	return f["response"] == want
}

func TestFetchDataDigestAuth(t *testing.T) {
	const realm, nonce, user, pass = "shellyplugus-test", "deadbeefnonce", "admin", "s3cret-pw"
	var authed bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if authz == "" || !validDigest(authz, r.Method, realm, nonce, user, pass) {
			w.Header().Set("WWW-Authenticate",
				`Digest qop="auth", realm="`+realm+`", nonce="`+nonce+`", algorithm=SHA-256`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		authed = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"value":42}`))
	}))
	defer srv.Close()

	c := NewAPIClient(strings.TrimPrefix(srv.URL, "http://"), user, pass, 5*time.Second)
	var out struct {
		Value int `json:"value"`
	}
	if err := c.FetchData("/rpc/Shelly.GetStatus", &out); err != nil {
		t.Fatalf("FetchData: %v", err)
	}
	if out.Value != 42 {
		t.Fatalf("value = %d, want 42", out.Value)
	}
	if !authed {
		t.Fatal("server never accepted a valid digest response")
	}
}

// After the first challenge the client should reuse the cached nonce and send
// Authorization pre-emptively (with an incrementing nc), not re-challenge.
func TestFetchDataReusesCachedChallenge(t *testing.T) {
	const realm, nonce, user, pass = "shelly-test", "noncevalue", "admin", "pw"
	challenges := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if authz == "" || !validDigest(authz, r.Method, realm, nonce, user, pass) {
			challenges++
			w.Header().Set("WWW-Authenticate",
				`Digest qop="auth", realm="`+realm+`", nonce="`+nonce+`", algorithm=SHA-256`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewAPIClient(strings.TrimPrefix(srv.URL, "http://"), user, pass, 5*time.Second)
	var out map[string]any
	for i := 0; i < 3; i++ {
		if err := c.FetchData("/rpc/Shelly.GetStatus", &out); err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
	}
	if challenges != 1 {
		t.Fatalf("got %d challenges, want exactly 1 (nonce should be cached and reused)", challenges)
	}
}

func TestFetchDataNoPasswordSendsNoAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("unexpected Authorization header when no password configured")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"value":7}`))
	}))
	defer srv.Close()

	c := NewAPIClient(strings.TrimPrefix(srv.URL, "http://"), "", "", 5*time.Second)
	var out struct {
		Value int `json:"value"`
	}
	if err := c.FetchData("/rpc/Shelly.GetStatus", &out); err != nil {
		t.Fatalf("FetchData: %v", err)
	}
	if out.Value != 7 {
		t.Fatalf("value = %d, want 7", out.Value)
	}
}

func TestFetchDataWrongPasswordSurfacesError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always reject - simulates a bad password.
		w.Header().Set("WWW-Authenticate",
			`Digest qop="auth", realm="r", nonce="n", algorithm=SHA-256`)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := NewAPIClient(strings.TrimPrefix(srv.URL, "http://"), "admin", "wrong", 5*time.Second)
	var out map[string]any
	err := c.FetchData("/rpc/Shelly.GetStatus", &out)
	if err == nil {
		t.Fatal("expected an error for a rejected password, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected a 401 error, got: %v", err)
	}
}
