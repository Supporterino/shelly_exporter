package exporter

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/supporterino/shelly_exporter/internal/client"
	"github.com/supporterino/shelly_exporter/internal/config"
)

func TestHandlers(t *testing.T) {
	reg := prometheus.NewRegistry()
	if _, err := NewCollectors(reg); err != nil {
		t.Fatalf("NewCollectors failed: %v", err)
	}
	handler := NewHandler(reg)

	tests := []struct {
		path string
	}{
		{path: "/"},
		{path: "/health"},
		{path: "/metrics"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if rec.Code != http.StatusOK {
				t.Errorf("GET %s status = %d, want 200", tt.path, rec.Code)
			}
		})
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(rec.Body.String(), "/metrics") {
		t.Error("landing page does not link to /metrics")
	}
}

func TestHealthRemainsOKWithUnreachableDevice(t *testing.T) {
	reg := prometheus.NewRegistry()
	cols, err := NewCollectors(reg)
	if err != nil {
		t.Fatalf("NewCollectors failed: %v", err)
	}
	dm := NewDeviceManager(time.Hour, cols, nil, testLogger())

	fetcher := newMockFetcher(map[string]func(int) (any, error){
		"/rpc/Shelly.GetDeviceInfo": func(int) (any, error) { return nil, context.DeadlineExceeded },
		"/rpc/Shelly.GetStatus":     func(int) (any, error) { return nil, context.DeadlineExceeded },
		"/rpc/Shelly.GetConfig":     func(int) (any, error) { return nil, context.DeadlineExceeded },
	})
	state := &deviceState{host: "host1", fetcher: fetcher}
	if err := dm.poll(context.Background(), state); err == nil {
		t.Fatal("expected poll to fail")
	}

	rec := httptest.NewRecorder()
	NewHandler(reg).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("GET /health status = %d, want 200 despite device failure", rec.Code)
	}
}

func TestLoggerLevel(t *testing.T) {
	ctx := context.Background()

	if NewLogger(false, io.Discard).Enabled(ctx, slog.LevelDebug) {
		t.Error("debug level enabled without the debug flag")
	}
	if !NewLogger(true, io.Discard).Enabled(ctx, slog.LevelDebug) {
		t.Error("debug level not enabled with the debug flag")
	}
}

func TestRunReturnsServerErrorAndStopsPolling(t *testing.T) {
	blocker := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer blocker.Close()

	cfg := config.Default()
	cfg.ListenAddress = strings.TrimPrefix(blocker.URL, "http://")
	cfg.Devices = []config.Device{{Host: "127.0.0.1:1"}}

	app, err := New(&cfg, testLogger())
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- app.Run(context.Background()) }()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected Run to return the server error")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after the server failed")
	}
}

func TestExporterServesMetricsAndHealth(t *testing.T) {
	payloads := map[string]string{
		"/rpc/Shelly.GetDeviceInfo": `{"name":"Dev","id":"123","mac":"AA:BB","model":"PlusPlugS","fw_id":"fw","app":"PlusPlugS","auth_en":false,"profile":"switch"}`,
		"/rpc/Shelly.GetStatus":     `{"sys":{"mac":"AA:BB","uptime":3600},"wifi":{"sta_ip":"1.2.3.4","status":"got ip","ssid":"Net","rssi":-50}}`,
		"/rpc/Shelly.GetConfig":     `{"sys":{"device":{"mac":"AA:BB"}}}`,
		"/rpc/Switch.GetStatus":     `{"id":0,"output":true,"apower":10}`,
		"/rpc/Switch.GetConfig":     `{"id":0,"initial_state":"on","power_limit":100}`,
		"/rpc/WiFi.GetStatus":       `{"sta_ip":"1.2.3.4","status":"got ip","ssid":"Net","rssi":-55}`,
	}

	deviceSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, ok := payloads[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(payload))
	}))
	defer deviceSrv.Close()

	host := strings.TrimPrefix(deviceSrv.URL, "http://")

	reg := prometheus.NewRegistry()
	cols, err := NewCollectors(reg)
	if err != nil {
		t.Fatalf("NewCollectors failed: %v", err)
	}
	factory := func(_, _, _ string) client.Fetcher {
		return client.NewAPIClient(host)
	}
	dm := NewDeviceManager(5*time.Millisecond, cols, factory, testLogger())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dm.Register(ctx, Device{Host: host})

	deadline := time.Now().Add(2 * time.Second)
	for upValue(t, reg, "AA:BB", host) != 1 {
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for a successful poll")
		}
		time.Sleep(5 * time.Millisecond)
	}

	srv := httptest.NewServer(NewHandler(reg))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	for _, name := range []string{"shelly_up", "shelly_system_uptime", "shelly_device_info", "shelly_switch_state"} {
		if !strings.Contains(string(body), name) {
			t.Errorf("/metrics does not expose %s", name)
		}
	}

	health, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	_ = health.Body.Close()
	if health.StatusCode != http.StatusOK {
		t.Errorf("GET /health status = %d, want 200", health.StatusCode)
	}

	cancel()
	dm.Wait()
}
