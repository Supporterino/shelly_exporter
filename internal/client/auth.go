package client

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/icholy/digest"
)

type authScheme int

const (
	authSchemeUnknown authScheme = iota
	authSchemeBasic
	authSchemeDigest
)

// authTransport selects an authentication scheme from the device's
// WWW-Authenticate challenge and caches it for subsequent requests.
type authTransport struct {
	base     http.RoundTripper
	username string
	password string

	mu         sync.Mutex
	scheme     authScheme
	challenge  *digest.Challenge
	nonceCount int
}

func newAuthTransport(base http.RoundTripper, username, password string) *authTransport {
	return &authTransport{base: base, username: username, password: password}
}

// RoundTrip probes an unauthenticated request when the scheme is unknown and
// retries with the scheme the device advertises. A cached scheme that is
// rejected with a 401 is re-negotiated once.
func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	scheme := t.scheme
	challenge := t.challenge
	t.mu.Unlock()

	var (
		resp *http.Response
		err  error
	)

	switch scheme {
	case authSchemeBasic:
		resp, err = t.roundTripBasic(req)
	case authSchemeDigest:
		resp, err = t.roundTripDigest(req, challenge)
	default:
		resp, err = t.base.RoundTrip(cloneRequest(req))
	}
	if err != nil {
		return resp, err
	}
	if resp.StatusCode != http.StatusUnauthorized {
		return resp, nil
	}

	return t.renegotiate(req, resp)
}

// renegotiate parses the WWW-Authenticate challenge from a rejected response,
// updates the cached scheme, and retries the request once.
func (t *authTransport) renegotiate(req *http.Request, resp *http.Response) (*http.Response, error) {
	challenge, digestErr := digest.FindChallenge(resp.Header)
	hasBasic := hasBasicChallenge(resp.Header)
	if digestErr != nil && !hasBasic {
		return resp, nil
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	if digestErr == nil {
		t.mu.Lock()
		t.scheme = authSchemeDigest
		t.challenge = challenge
		t.nonceCount = 0
		t.mu.Unlock()
		return t.roundTripDigest(req, challenge)
	}

	t.mu.Lock()
	t.scheme = authSchemeBasic
	t.mu.Unlock()
	return t.roundTripBasic(req)
}

func (t *authTransport) roundTripBasic(req *http.Request) (*http.Response, error) {
	clone := cloneRequest(req)
	clone.SetBasicAuth(t.username, t.password)
	return t.base.RoundTrip(clone)
}

func (t *authTransport) roundTripDigest(req *http.Request, challenge *digest.Challenge) (*http.Response, error) {
	if challenge == nil {
		return t.base.RoundTrip(cloneRequest(req))
	}

	clone := cloneRequest(req)

	t.mu.Lock()
	t.nonceCount++
	count := t.nonceCount
	t.mu.Unlock()

	credentials, err := digest.Digest(challenge, digest.Options{
		Method:   clone.Method,
		URI:      clone.URL.RequestURI(),
		GetBody:  clone.GetBody,
		Count:    count,
		Username: t.username,
		Password: t.password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to compute digest credentials: %w", err)
	}

	clone.Header.Set("Authorization", credentials.String())
	return t.base.RoundTrip(clone)
}

func cloneRequest(req *http.Request) *http.Request {
	clone := req.Clone(req.Context())
	clone.Header = req.Header.Clone()
	return clone
}

func hasBasicChallenge(header http.Header) bool {
	const prefix = "Basic"
	for _, value := range header.Values("WWW-Authenticate") {
		if len(value) >= len(prefix) && strings.EqualFold(value[:len(prefix)], prefix) {
			return true
		}
	}
	return false
}
