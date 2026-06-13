package client

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type APIClient struct {
	BaseURL  string
	Username string
	Password string
	Client   *http.Client

	mu   sync.Mutex
	auth *digestState // cached challenge, populated after the first 401
}

// digestState holds a parsed WWW-Authenticate digest challenge. nc is the
// request counter incremented on each authenticated request that reuses the
// cached nonce.
type digestState struct {
	realm  string
	nonce  string
	opaque string
	nc     uint32
}

// NewAPIClient initializes and returns a new APIClient. A non-empty
// username/password enables Shelly Gen2 HTTP digest auth (SHA-256, qop=auth);
// pass empty strings to talk to a device with auth disabled.
func NewAPIClient(baseURL, username, password string, timeout time.Duration) *APIClient {
	return &APIClient{
		BaseURL:  baseURL,
		Username: username,
		Password: password,
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

// FetchData makes a GET request to the specified endpoint and parses the
// response. When a password is configured and the device answers 401, it
// answers the digest challenge and retries once.
func (c *APIClient) FetchData(endpoint string, result interface{}) error {
	url := fmt.Sprintf("http://%s%s", c.BaseURL, endpoint)
	slog.Info("Fetching data", slog.String("url", url))

	resp, err := c.do(endpoint, url)
	if err != nil {
		slog.Error("Failed to fetch data", slog.String("url", url), slog.Any("error", err))
		return fmt.Errorf("failed to fetch data from %s: %w", url, err)
	}

	// Answer a digest challenge once, then retry, when we hold a password.
	if resp.StatusCode == http.StatusUnauthorized && c.Password != "" {
		challenge := resp.Header.Get("WWW-Authenticate")
		_, _ = io.Copy(io.Discard, resp.Body) // drain so the connection can be reused
		_ = resp.Body.Close()
		if err := c.setChallenge(challenge); err != nil {
			slog.Error("Digest auth failed", slog.String("url", url), slog.Any("error", err))
			return fmt.Errorf("digest auth for %s: %w", url, err)
		}
		resp, err = c.do(endpoint, url)
		if err != nil {
			slog.Error("Failed to fetch data", slog.String("url", url), slog.Any("error", err))
			return fmt.Errorf("failed to fetch data from %s: %w", url, err)
		}
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("Failed to close response body", slog.Any("error", err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("Non-200 status code", slog.String("url", url), slog.Int("status_code", resp.StatusCode))
		return fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body) // Use io.ReadAll instead of ioutil.ReadAll
	if err != nil {
		slog.Error("Failed to read response body", slog.String("url", url), slog.Any("error", err))
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, result); err != nil {
		slog.Error("Failed to parse JSON response", slog.String("url", url), slog.Any("error", err))
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	slog.Info("Successfully fetched data", slog.String("url", url))
	return nil
}

// do issues a GET, attaching an Authorization header when a digest challenge
// has already been cached.
func (c *APIClient) do(endpoint, url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	if c.auth != nil && c.Password != "" {
		c.auth.nc++
		req.Header.Set("Authorization", c.authorizationLocked(http.MethodGet, endpoint, c.auth.nc))
	}
	c.mu.Unlock()
	return c.Client.Do(req)
}

var digestFieldRe = regexp.MustCompile(`(\w+)=(?:"([^"]*)"|([^",\s]+))`)

// setChallenge parses a WWW-Authenticate digest header and caches its
// parameters for subsequent requests.
func (c *APIClient) setChallenge(header string) error {
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(header)), "digest") {
		return fmt.Errorf("not a digest challenge: %q", header)
	}
	fields := map[string]string{}
	for _, m := range digestFieldRe.FindAllStringSubmatch(header, -1) {
		v := m[2]
		if v == "" {
			v = m[3]
		}
		fields[strings.ToLower(m[1])] = v
	}
	if fields["realm"] == "" || fields["nonce"] == "" {
		return fmt.Errorf("incomplete digest challenge: %q", header)
	}
	c.mu.Lock()
	c.auth = &digestState{realm: fields["realm"], nonce: fields["nonce"], opaque: fields["opaque"]}
	c.mu.Unlock()
	return nil
}

// authorizationLocked builds the Digest Authorization header value. The caller
// must hold c.mu (it reads c.auth).
func (c *APIClient) authorizationLocked(method, uri string, nc uint32) string {
	ha1 := sha256hex(c.Username + ":" + c.auth.realm + ":" + c.Password)
	ha2 := sha256hex(method + ":" + uri)
	ncStr := fmt.Sprintf("%08x", nc)
	cnonce := randomHex(8)
	response := sha256hex(strings.Join([]string{ha1, c.auth.nonce, ncStr, cnonce, "auth", ha2}, ":"))
	h := fmt.Sprintf(
		`Digest username=%q, realm=%q, nonce=%q, uri=%q, response=%q, qop=auth, nc=%s, cnonce=%q, algorithm=SHA-256`,
		c.Username, c.auth.realm, c.auth.nonce, uri, response, ncStr, cnonce,
	)
	if c.auth.opaque != "" {
		h += fmt.Sprintf(`, opaque=%q`, c.auth.opaque)
	}
	return h
}

func sha256hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand.Read does not fail in practice; keep a valid-length
		// fallback so the header stays well-formed.
		return strings.Repeat("0", 2*n)
	}
	return hex.EncodeToString(b)
}

var componentKeyRe = regexp.MustCompile(`^(switch|cover):(\d+)$`)

// DiscoverComponents fetches Shelly.GetStatus and returns the IDs of
// switch and cover components found in the response keys.
func (c *APIClient) DiscoverComponents() (switchIDs []int, coverIDs []int, err error) {
	var raw map[string]json.RawMessage
	if err := c.FetchData("/rpc/Shelly.GetStatus", &raw); err != nil {
		return nil, nil, fmt.Errorf("failed to discover components: %w", err)
	}

	for key := range raw {
		m := componentKeyRe.FindStringSubmatch(key)
		if m == nil {
			continue
		}
		id, _ := strconv.Atoi(m[2])
		switch m[1] {
		case "switch":
			switchIDs = append(switchIDs, id)
		case "cover":
			coverIDs = append(coverIDs, id)
		}
	}

	sort.Ints(switchIDs)
	sort.Ints(coverIDs)

	slog.Info("Discovered components",
		slog.Any("switchIDs", switchIDs),
		slog.Any("coverIDs", coverIDs))
	return switchIDs, coverIDs, nil
}
