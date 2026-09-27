package exporter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/supporterino/shelly_exporter/internal/client"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type mockFetcher struct {
	mu        sync.Mutex
	calls     map[string]int
	responses map[string]func(call int) (any, error)
}

func newMockFetcher(responses map[string]func(call int) (any, error)) *mockFetcher {
	return &mockFetcher{calls: make(map[string]int), responses: responses}
}

func (m *mockFetcher) FetchData(_ context.Context, endpoint string, result any) error {
	path := strings.SplitN(endpoint, "?", 2)[0]

	m.mu.Lock()
	m.calls[path]++
	call := m.calls[path]
	handler := m.responses[endpoint]
	if handler == nil {
		handler = m.responses[path]
	}
	m.mu.Unlock()

	if handler == nil {
		return fmt.Errorf("unexpected endpoint %s", endpoint)
	}
	value, err := handler(call)
	if err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

func static(value any) func(int) (any, error) {
	return func(int) (any, error) { return value, nil }
}

func deviceInfoPayload(app, profile string) map[string]any {
	return map[string]any{
		"name": "Dev", "id": "123", "mac": "AA:BB", "model": "M",
		"fw_id": "fw", "app": app, "auth_en": false, "profile": profile,
	}
}

func statusPayload() map[string]any {
	return map[string]any{
		"sys":  map[string]any{"mac": "AA:BB", "uptime": 3600, "ram_size": 100, "ram_free": 40, "fs_size": 200, "fs_free": 80},
		"wifi": map[string]any{"sta_ip": "1.2.3.4", "status": "got ip", "ssid": "Net", "rssi": -50},
	}
}

func configPayload() map[string]any {
	return map[string]any{
		"ble":   map[string]any{"enable": true},
		"cloud": map[string]any{"enable": false, "server": "srv"},
		"eth":   map[string]any{"enable": true, "ipv4mode": "dhcp"},
		"wifi":  map[string]any{"ap": map[string]any{"enable": false}, "sta": map[string]any{"enable": true}, "roam": map[string]any{"rssi_thr": -70}},
		"sys":   map[string]any{"device": map[string]any{"mac": "AA:BB", "name": "x", "fw_id": "1"}},
	}
}

func switchStatusPayload() map[string]any {
	return map[string]any{"id": 0, "output": true, "apower": 10, "voltage": 230, "current": 0.05, "freq": 50, "aenergy": map[string]any{"total": 5}, "temperature": map[string]any{"tC": 40, "tF": 104}}
}

func switchConfigPayload() map[string]any {
	return map[string]any{"id": 0, "initial_state": "on", "auto_on": true, "auto_on_delay": 1.5, "auto_off": false, "auto_off_delay": 2.5, "autorecover_voltage_errors": true, "power_limit": 100, "voltage_limit": 250, "undervoltage_limit": 180, "current_limit": 2}
}

func wifiPayload() map[string]any {
	return map[string]any{"sta_ip": "1.2.3.4", "status": "got ip", "ssid": "Net", "rssi": -55, "ap_client_count": 0}
}

func coverStatusPayload() map[string]any {
	return map[string]any{"id": 0, "state": "open", "apower": 12.5, "voltage": 230, "current": 0.1, "pf": 1, "freq": 50, "aenergy": map[string]any{"total": 100}, "temperature": map[string]any{"tC": 30, "tF": 86}, "pos_control": true, "current_pos": 50}
}

func fullResponses(app, profile string) map[string]func(int) (any, error) {
	return map[string]func(int) (any, error){
		"/rpc/Shelly.GetDeviceInfo": static(deviceInfoPayload(app, profile)),
		"/rpc/Shelly.GetStatus":     static(statusPayload()),
		"/rpc/Shelly.GetConfig":     static(configPayload()),
		"/rpc/Switch.GetStatus":     static(switchStatusPayload()),
		"/rpc/Switch.GetConfig":     static(switchConfigPayload()),
		"/rpc/WiFi.GetStatus":       static(wifiPayload()),
		"/rpc/Cover.GetStatus":      static(coverStatusPayload()),
	}
}

func upValue(t *testing.T, reg *prometheus.Registry, mac, host string) float64 {
	t.Helper()
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}
	for _, family := range families {
		if family.GetName() != "shelly_up" {
			continue
		}
		for _, metric := range family.GetMetric() {
			labels := make(map[string]string)
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			if labels["device_mac"] == mac && labels["host"] == host {
				return metric.GetGauge().GetValue()
			}
		}
	}
	return 0
}

func upDeviceMACs(t *testing.T, reg *prometheus.Registry, host string) []string {
	t.Helper()
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}
	var macs []string
	for _, family := range families {
		if family.GetName() != "shelly_up" {
			continue
		}
		for _, metric := range family.GetMetric() {
			labels := make(map[string]string)
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			if labels["host"] == host {
				macs = append(macs, labels["device_mac"])
			}
		}
	}
	return macs
}

func TestPollSetsUpMetricOnSuccessAndFailure(t *testing.T) {
	reg := prometheus.NewRegistry()
	cols, err := NewCollectors(reg)
	if err != nil {
		t.Fatalf("NewCollectors failed: %v", err)
	}
	dm := NewDeviceManager(time.Hour, cols, nil, testLogger())

	healthy := true
	fetcher := newMockFetcher(fullResponses("PlusPlugS", "switch"))
	fetcher.responses["/rpc/Shelly.GetDeviceInfo"] = func(int) (any, error) {
		if !healthy {
			return nil, errors.New("unreachable")
		}
		return deviceInfoPayload("PlusPlugS", "switch"), nil
	}

	state := &deviceState{host: "host1", fetcher: fetcher}

	if err := dm.poll(context.Background(), state); err != nil {
		t.Fatalf("poll returned error: %v", err)
	}
	if got := upValue(t, reg, "AA:BB", "host1"); got != 1 {
		t.Errorf("up = %v, want 1", got)
	}

	healthy = false
	if err := dm.poll(context.Background(), state); err == nil {
		t.Fatal("expected poll to fail")
	}
	if got := upValue(t, reg, "AA:BB", "host1"); got != 0 {
		t.Errorf("up = %v, want 0", got)
	}
}

func TestPollUnreachableDeviceEmitsEmptyMACZero(t *testing.T) {
	reg := prometheus.NewRegistry()
	cols, err := NewCollectors(reg)
	if err != nil {
		t.Fatalf("NewCollectors failed: %v", err)
	}
	dm := NewDeviceManager(time.Hour, cols, nil, testLogger())

	fetcher := newMockFetcher(map[string]func(int) (any, error){
		"/rpc/Shelly.GetDeviceInfo": func(int) (any, error) { return nil, errors.New("unreachable") },
		"/rpc/Shelly.GetStatus":     func(int) (any, error) { return nil, errors.New("unreachable") },
		"/rpc/Shelly.GetConfig":     func(int) (any, error) { return nil, errors.New("unreachable") },
	})

	state := &deviceState{host: "host1", fetcher: fetcher}
	if err := dm.poll(context.Background(), state); err == nil {
		t.Fatal("expected poll to fail")
	}

	if got := upValue(t, reg, "", "host1"); got != 0 {
		t.Errorf("shelly_up{device_mac=\"\",host=host1} = %v, want 0", got)
	}
}

func TestEmptyMACSeriesRemovedAfterRecovery(t *testing.T) {
	reg := prometheus.NewRegistry()
	cols, err := NewCollectors(reg)
	if err != nil {
		t.Fatalf("NewCollectors failed: %v", err)
	}
	dm := NewDeviceManager(time.Hour, cols, nil, testLogger())

	available := false
	fetcher := newMockFetcher(fullResponses("PlusPlugS", "switch"))
	fetcher.responses["/rpc/Shelly.GetDeviceInfo"] = func(int) (any, error) {
		if !available {
			return nil, errors.New("unreachable")
		}
		return deviceInfoPayload("PlusPlugS", "switch"), nil
	}

	state := &deviceState{host: "host1", fetcher: fetcher}

	if err := dm.poll(context.Background(), state); err == nil {
		t.Fatal("expected first poll to fail")
	}
	if macs := upDeviceMACs(t, reg, "host1"); len(macs) != 1 || macs[0] != "" {
		t.Fatalf("up series after failure = %v, want one empty MAC", macs)
	}

	available = true
	if err := dm.poll(context.Background(), state); err != nil {
		t.Fatalf("poll returned error: %v", err)
	}

	macs := upDeviceMACs(t, reg, "host1")
	if len(macs) != 1 || macs[0] != "AA:BB" {
		t.Fatalf("up series after recovery = %v, want one AA:BB MAC", macs)
	}
}

func TestMACChangeReplacesUpSeries(t *testing.T) {
	reg := prometheus.NewRegistry()
	cols, err := NewCollectors(reg)
	if err != nil {
		t.Fatalf("NewCollectors failed: %v", err)
	}
	dm := NewDeviceManager(time.Hour, cols, nil, testLogger())

	mac := "AA:BB"
	fetcher := newMockFetcher(fullResponses("PlusPlugS", "switch"))
	fetcher.responses["/rpc/Shelly.GetDeviceInfo"] = func(int) (any, error) {
		return map[string]any{
			"name": "Dev", "id": "123", "mac": mac, "model": "M",
			"fw_id": "fw", "app": "PlusPlugS", "auth_en": false, "profile": "switch",
		}, nil
	}

	state := &deviceState{host: "host1", fetcher: fetcher}
	if err := dm.poll(context.Background(), state); err != nil {
		t.Fatalf("poll returned error: %v", err)
	}

	mac = "CC:DD"
	if err := dm.poll(context.Background(), state); err != nil {
		t.Fatalf("poll returned error: %v", err)
	}

	macs := upDeviceMACs(t, reg, "host1")
	if len(macs) != 1 || macs[0] != "CC:DD" {
		t.Fatalf("up series after MAC change = %v, want one CC:DD MAC", macs)
	}
}

func TestUnreachableDeviceResolvesModelLater(t *testing.T) {
	reg := prometheus.NewRegistry()
	cols, err := NewCollectors(reg)
	if err != nil {
		t.Fatalf("NewCollectors failed: %v", err)
	}
	dm := NewDeviceManager(time.Hour, cols, nil, testLogger())

	available := false
	fetcher := newMockFetcher(fullResponses("PlusPlugS", "switch"))
	fetcher.responses["/rpc/Shelly.GetDeviceInfo"] = func(int) (any, error) {
		if !available {
			return nil, errors.New("unreachable")
		}
		return deviceInfoPayload("PlusPlugS", "switch"), nil
	}

	state := &deviceState{host: "host1", fetcher: fetcher}

	if err := dm.poll(context.Background(), state); err == nil {
		t.Fatal("expected first poll to fail")
	}
	if got := testutil.ToFloat64(cols.ShellyStatus.Uptime.WithLabelValues("AA:BB")); got != 3600 {
		t.Errorf("generic status metric = %v, want 3600", got)
	}
	if got := testutil.ToFloat64(cols.SwitchStatus.State.WithLabelValues("AA:BB", "0", "")); got != 0 {
		t.Errorf("switch metric before model resolved = %v, want 0", got)
	}

	available = true
	if err := dm.poll(context.Background(), state); err != nil {
		t.Fatalf("poll returned error: %v", err)
	}
	if state.model != "PlusPlugS" {
		t.Errorf("model = %q, want PlusPlugS", state.model)
	}
	if got := testutil.ToFloat64(cols.SwitchStatus.State.WithLabelValues("AA:BB", "0", "")); got != 1 {
		t.Errorf("switch metric after model resolved = %v, want 1", got)
	}
}

func TestModelDispatch(t *testing.T) {
	tests := []struct {
		name       string
		app        string
		profile    string
		wantSwitch bool
		wantCover  bool
	}{
		{name: "cover model", app: "Plus2PM", profile: "cover", wantCover: true},
		{name: "switch model", app: "PlusPlugS", profile: "switch", wantSwitch: true},
		{name: "unknown model", app: "Mystery", profile: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			cols, err := NewCollectors(reg)
			if err != nil {
				t.Fatalf("NewCollectors failed: %v", err)
			}
			dm := NewDeviceManager(time.Hour, cols, nil, testLogger())
			fetcher := newMockFetcher(fullResponses(tt.app, tt.profile))
			state := &deviceState{host: "host1", fetcher: fetcher}

			if err := dm.poll(context.Background(), state); err != nil {
				t.Fatalf("poll returned error: %v", err)
			}

			switchState := testutil.ToFloat64(cols.SwitchStatus.State.WithLabelValues("AA:BB", "0", ""))
			coverState := testutil.ToFloat64(cols.CoverStatus.State.WithLabelValues("AA:BB", "0", ""))

			if (switchState == 1) != tt.wantSwitch {
				t.Errorf("switch state = %v, want switch=%v", switchState, tt.wantSwitch)
			}
			if (coverState == 1) != tt.wantCover {
				t.Errorf("cover state = %v, want cover=%v", coverState, tt.wantCover)
			}
		})
	}
}

func configWithSwitches(n int) map[string]any {
	cfg := configPayload()
	for i := 0; i < n; i++ {
		cfg[fmt.Sprintf("switch:%d", i)] = map[string]any{
			"id":            i,
			"name":          fmt.Sprintf("Channel %d", i),
			"initial_state": "match_input",
		}
	}
	return cfg
}

func channelResponses(app, profile string, switches int) map[string]func(int) (any, error) {
	responses := fullResponses(app, profile)
	responses["/rpc/Shelly.GetConfig"] = static(configWithSwitches(switches))
	for id := 0; id < switches; id++ {
		status := switchStatusPayload()
		status["id"] = id
		status["apower"] = 10.0 + float64(id)
		responses[fmt.Sprintf("/rpc/Switch.GetStatus?id=%d", id)] = static(status)

		cfg := switchConfigPayload()
		cfg["id"] = id
		responses[fmt.Sprintf("/rpc/Switch.GetConfig?id=%d", id)] = static(cfg)
	}
	return responses
}

func TestPro4PMCollectsEveryChannel(t *testing.T) {
	reg := prometheus.NewRegistry()
	cols, err := NewCollectors(reg)
	if err != nil {
		t.Fatalf("NewCollectors failed: %v", err)
	}
	dm := NewDeviceManager(time.Hour, cols, nil, testLogger())

	fetcher := newMockFetcher(channelResponses("Pro4PM", "", 4))
	state := &deviceState{host: "rack", fetcher: fetcher}

	if err := dm.poll(context.Background(), state); err != nil {
		t.Fatalf("poll returned error: %v", err)
	}

	for id := 0; id < 4; id++ {
		want := 10.0 + float64(id)
		name := fmt.Sprintf("Channel %d", id)
		got := testutil.ToFloat64(cols.SwitchStatus.APower.WithLabelValues("AA:BB", fmt.Sprintf("%d", id), name))
		if got != want {
			t.Errorf("switch %d power = %v, want %v", id, got, want)
		}
	}
}

func TestPlus2PMSwitchProfileCollectsSwitches(t *testing.T) {
	reg := prometheus.NewRegistry()
	cols, err := NewCollectors(reg)
	if err != nil {
		t.Fatalf("NewCollectors failed: %v", err)
	}
	dm := NewDeviceManager(time.Hour, cols, nil, testLogger())

	fetcher := newMockFetcher(channelResponses("Plus2PM", "switch", 2))
	state := &deviceState{host: "shutter", fetcher: fetcher}

	if err := dm.poll(context.Background(), state); err != nil {
		t.Fatalf("poll returned error: %v", err)
	}

	for id := 0; id < 2; id++ {
		name := fmt.Sprintf("Channel %d", id)
		got := testutil.ToFloat64(cols.SwitchStatus.State.WithLabelValues("AA:BB", fmt.Sprintf("%d", id), name))
		if got != 1 {
			t.Errorf("switch %d state = %v, want 1", id, got)
		}
	}
	if got := testutil.ToFloat64(cols.CoverStatus.State.WithLabelValues("AA:BB", "0", "")); got != 0 {
		t.Errorf("cover state = %v, want 0 for switch profile", got)
	}
}

func TestPollLoopStopsOnContextCancel(t *testing.T) {
	reg := prometheus.NewRegistry()
	cols, err := NewCollectors(reg)
	if err != nil {
		t.Fatalf("NewCollectors failed: %v", err)
	}
	factory := func(_, _, _ string) client.Fetcher {
		return newMockFetcher(fullResponses("PlusPlugS", "switch"))
	}
	dm := NewDeviceManager(time.Millisecond, cols, factory, testLogger())

	ctx, cancel := context.WithCancel(context.Background())
	dm.Register(ctx, Device{Host: "host1"})
	time.Sleep(20 * time.Millisecond)
	cancel()

	done := make(chan struct{})
	go func() {
		dm.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("polling loops did not stop after context cancellation")
	}
}

func TestDuplicateRegistrationStartsSingleLoop(t *testing.T) {
	reg := prometheus.NewRegistry()
	cols, err := NewCollectors(reg)
	if err != nil {
		t.Fatalf("NewCollectors failed: %v", err)
	}
	factory := func(_, _, _ string) client.Fetcher {
		return newMockFetcher(fullResponses("PlusPlugS", "switch"))
	}
	dm := NewDeviceManager(time.Millisecond, cols, factory, testLogger())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dm.Register(ctx, Device{Host: "host1"})
	dm.Register(ctx, Device{Host: "host1"})

	dm.mu.Lock()
	count := len(dm.devices)
	dm.mu.Unlock()
	if count != 1 {
		t.Errorf("registered devices = %d, want 1", count)
	}

	cancel()
	dm.Wait()
}

func TestMultipleDevicesNoRace(t *testing.T) {
	reg := prometheus.NewRegistry()
	cols, err := NewCollectors(reg)
	if err != nil {
		t.Fatalf("NewCollectors failed: %v", err)
	}
	factory := func(host, _, _ string) client.Fetcher {
		fetcher := newMockFetcher(fullResponses("PlusPlugS", "switch"))
		fetcher.responses["/rpc/Shelly.GetDeviceInfo"] = static(map[string]any{
			"name": host, "id": "123", "mac": "AA:" + host, "model": "M",
			"fw_id": "fw", "app": "PlusPlugS", "auth_en": false, "profile": "switch",
		})
		return fetcher
	}
	dm := NewDeviceManager(time.Millisecond, cols, factory, testLogger())

	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < 5; i++ {
		dm.Register(ctx, Device{Host: fmt.Sprintf("%02d", i)})
	}

	time.Sleep(50 * time.Millisecond)
	cancel()
	dm.Wait()
}
