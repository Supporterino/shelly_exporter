package collector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/supporterino/shelly_exporter/internal/client"
)

func newTestFetcher(t *testing.T, payloads map[string]string) client.Fetcher {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, ok := payloads[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(payload))
	}))
	t.Cleanup(srv.Close)

	return client.NewAPIClient(strings.TrimPrefix(srv.URL, "http://"))
}

func TestCollectorsUseIndependentRegistries(t *testing.T) {
	constructors := map[string]func(prometheus.Registerer) error{
		"shelly_status": func(reg prometheus.Registerer) error { _, err := NewShellyStatusCollector(reg); return err },
		"shelly_config": func(reg prometheus.Registerer) error { _, err := NewShellyConfigCollector(reg); return err },
		"device_info":   func(reg prometheus.Registerer) error { _, err := NewDeviceInfoCollector(reg); return err },
		"cover_status":  func(reg prometheus.Registerer) error { _, err := NewCoverStatusCollector(reg); return err },
		"switch_status": func(reg prometheus.Registerer) error { _, err := NewSwitchStatusCollector(reg); return err },
		"switch_config": func(reg prometheus.Registerer) error { _, err := NewSwitchConfigCollector(reg); return err },
		"wifi_status":   func(reg prometheus.Registerer) error { _, err := NewWiFiStatusCollector(reg); return err },
		"up":            func(reg prometheus.Registerer) error { _, err := NewUpCollector(reg); return err },
	}

	for name, construct := range constructors {
		t.Run(name, func(t *testing.T) {
			if err := construct(prometheus.NewRegistry()); err != nil {
				t.Fatalf("first construction failed: %v", err)
			}
			if err := construct(prometheus.NewRegistry()); err != nil {
				t.Fatalf("second construction failed: %v", err)
			}
		})
	}
}

func TestShellyStatusCharacterization(t *testing.T) {
	fetcher := newTestFetcher(t, map[string]string{
		"/rpc/Shelly.GetStatus": `{"sys":{"mac":"AA:BB","uptime":3600,"ram_size":100,"ram_free":40,"fs_size":200,"fs_free":80},"wifi":{"sta_ip":"1.2.3.4","status":"got ip","ssid":"Net","rssi":-50}}`,
	})

	reg := prometheus.NewRegistry()
	c, err := NewShellyStatusCollector(reg)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}
	if err := c.Update(context.Background(), fetcher); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if got := testutil.ToFloat64(c.Uptime.WithLabelValues("AA:BB")); got != 3600 {
		t.Errorf("uptime = %v, want 3600", got)
	}
	if got := testutil.ToFloat64(c.WiFiRSSI.WithLabelValues("AA:BB", "Net", "1.2.3.4")); got != -50 {
		t.Errorf("wifi rssi = %v, want -50", got)
	}
	assertFamily(t, reg, "shelly_system_uptime")
}

func TestShellyStatusNilWiFiIsSafe(t *testing.T) {
	c, err := NewShellyStatusCollector(prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}

	status := client.ShellyGetStatusResponse{}
	status.Sys.MAC = "AA:BB"
	status.Sys.Uptime = 42
	status.Sys.RAMFree = 10

	c.UpdateMetrics(status)

	if got := testutil.ToFloat64(c.Uptime.WithLabelValues("AA:BB")); got != 42 {
		t.Errorf("uptime = %v, want 42", got)
	}
	if got := testutil.ToFloat64(c.RAM.WithLabelValues("AA:BB", "free")); got != 10 {
		t.Errorf("ram free = %v, want 10", got)
	}
	if got := testutil.ToFloat64(c.WiFiRSSI.WithLabelValues("AA:BB", "", "")); got != 0 {
		t.Errorf("wifi rssi = %v, want 0", got)
	}
}

func TestShellyConfigCharacterization(t *testing.T) {
	fetcher := newTestFetcher(t, map[string]string{
		"/rpc/Shelly.GetConfig": `{"ble":{"enable":true},"cloud":{"enable":false,"server":"srv"},"eth":{"enable":true,"ipv4mode":"dhcp"},"wifi":{"ap":{"enable":false},"sta":{"enable":true},"roam":{"rssi_thr":-70}},"sys":{"device":{"mac":"AA:BB","name":"x","fw_id":"1"}}}`,
	})

	reg := prometheus.NewRegistry()
	c, err := NewShellyConfigCollector(reg)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}
	if _, err := c.Update(context.Background(), fetcher); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if got := testutil.ToFloat64(c.BLEEnabled.WithLabelValues("AA:BB")); got != 1 {
		t.Errorf("ble = %v, want 1", got)
	}
	if got := testutil.ToFloat64(c.WiFiRoamingThreshold.WithLabelValues("AA:BB")); got != -70 {
		t.Errorf("roaming threshold = %v, want -70", got)
	}
	assertFamily(t, reg, "shelly_device_ble")
	assertFamily(t, reg, "selly_device_eth_ipv4_mode")
}

func TestDeviceInfoCharacterization(t *testing.T) {
	fetcher := newTestFetcher(t, map[string]string{
		"/rpc/Shelly.GetDeviceInfo": `{"name":"Dev","id":"123","mac":"AA:BB","model":"Plus2PM","fw_id":"20240101","app":"Plus2PM","auth_en":true,"profile":"cover"}`,
	})

	reg := prometheus.NewRegistry()
	c, err := NewDeviceInfoCollector(reg)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}
	info, err := c.Update(context.Background(), fetcher)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if info.Mac != "AA:BB" || info.App != "Plus2PM" || info.Profile != "cover" {
		t.Errorf("info = %+v, want mac AA:BB app Plus2PM profile cover", info)
	}

	if got := testutil.ToFloat64(c.AuthEnabled.WithLabelValues("AA:BB")); got != 1 {
		t.Errorf("auth = %v, want 1", got)
	}
	assertFamily(t, reg, "shelly_device_info")
}

func TestCoverStatusCharacterization(t *testing.T) {
	fetcher := newTestFetcher(t, map[string]string{
		"/rpc/Cover.GetStatus": `{"id":0,"state":"open","apower":12.5,"voltage":230,"current":0.1,"pf":1,"freq":50,"aenergy":{"total":100},"temperature":{"tC":30,"tF":86},"pos_control":true,"current_pos":50}`,
	})

	reg := prometheus.NewRegistry()
	c, err := NewCoverStatusCollector(reg)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}
	if err := c.Update(context.Background(), fetcher, 0, "Blinds", "AA:BB"); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if got := testutil.ToFloat64(c.State.WithLabelValues("AA:BB", "0", "Blinds")); got != 1 {
		t.Errorf("state = %v, want 1", got)
	}
	if got := testutil.ToFloat64(c.APower.WithLabelValues("AA:BB", "0", "Blinds")); got != 12.5 {
		t.Errorf("power = %v, want 12.5", got)
	}
	assertFamily(t, reg, "shelly_cover_state")
}

func TestSwitchStatusCharacterization(t *testing.T) {
	fetcher := newTestFetcher(t, map[string]string{
		"/rpc/Switch.GetStatus": `{"id":0,"output":true,"apower":10,"voltage":230,"current":0.05,"freq":50,"aenergy":{"total":5},"temperature":{"tC":40,"tF":104}}`,
	})

	reg := prometheus.NewRegistry()
	c, err := NewSwitchStatusCollector(reg)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}
	if err := c.Update(context.Background(), fetcher, 0, "Lamp", "AA:BB"); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if got := testutil.ToFloat64(c.State.WithLabelValues("AA:BB", "0", "Lamp")); got != 1 {
		t.Errorf("state = %v, want 1", got)
	}
	assertFamily(t, reg, "shelly_switch_state")
}

func TestSwitchConfigCharacterization(t *testing.T) {
	fetcher := newTestFetcher(t, map[string]string{
		"/rpc/Switch.GetConfig": `{"id":0,"name":"Lamp","initial_state":"on","auto_on":true,"auto_on_delay":1.5,"auto_off":false,"auto_off_delay":2.5,"autorecover_voltage_errors":true,"power_limit":100,"voltage_limit":250,"undervoltage_limit":180,"current_limit":2}`,
	})

	reg := prometheus.NewRegistry()
	c, err := NewSwitchConfigCollector(reg)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}
	if err := c.Update(context.Background(), fetcher, 0, "AA:BB"); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if got := testutil.ToFloat64(c.InitialState.WithLabelValues("AA:BB", "0", "Lamp")); got != 1 {
		t.Errorf("initial state = %v, want 1", got)
	}
	if got := testutil.ToFloat64(c.PowerLimit.WithLabelValues("AA:BB", "0", "Lamp")); got != 100 {
		t.Errorf("power limit = %v, want 100", got)
	}
	assertFamily(t, reg, "shelly_switch_initial_state")
}

func TestWiFiStatusCharacterization(t *testing.T) {
	fetcher := newTestFetcher(t, map[string]string{
		"/rpc/WiFi.GetStatus": `{"sta_ip":"1.2.3.4","status":"got ip","ssid":"Net","rssi":-55,"ap_client_count":0}`,
	})

	reg := prometheus.NewRegistry()
	c, err := NewWiFiStatusCollector(reg)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}
	if err := c.Update(context.Background(), fetcher, "AA:BB"); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if got := testutil.ToFloat64(c.RSSI.WithLabelValues("AA:BB")); got != -55 {
		t.Errorf("rssi = %v, want -55", got)
	}
	assertFamily(t, reg, "shelly_wifi_rssi")
}

func TestUpCollector(t *testing.T) {
	reg := prometheus.NewRegistry()
	c, err := NewUpCollector(reg)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}

	c.Set(1, "AA:BB", "host1")
	c.Set(0, "", "host2")

	if got := testutil.ToFloat64(c.up.WithLabelValues("AA:BB", "host1")); got != 1 {
		t.Errorf("up = %v, want 1", got)
	}
	if got := testutil.ToFloat64(c.up.WithLabelValues("", "host2")); got != 0 {
		t.Errorf("up = %v, want 0", got)
	}
	assertFamily(t, reg, "shelly_up")
}

func TestSwitchStatusSkipsAbsentFields(t *testing.T) {
	fetcher := newTestFetcher(t, map[string]string{
		"/rpc/Switch.GetStatus": `{"id":0,"output":true,"temperature":{"tC":38,"tF":100}}`,
	})

	reg := prometheus.NewRegistry()
	c, err := NewSwitchStatusCollector(reg)
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}
	if err := c.Update(context.Background(), fetcher, 0, "", "AA:BB"); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if got := testutil.ToFloat64(c.State.WithLabelValues("AA:BB", "0", "")); got != 1 {
		t.Errorf("state = %v, want 1", got)
	}
	if got := testutil.ToFloat64(c.Temperature.WithLabelValues("AA:BB", "0", "", "dC")); got != 38 {
		t.Errorf("temperature = %v, want 38", got)
	}

	absent := map[string]*prometheus.GaugeVec{
		"power": c.APower, "voltage": c.Voltage, "current": c.Current,
		"frequency": c.Freq, "energy": c.Energy,
	}
	for name, vec := range absent {
		if count := testutil.CollectAndCount(vec); count != 0 {
			t.Errorf("%s series = %d, want 0 for a device that omits the field", name, count)
		}
	}
}

func TestSwitchConfigSkipsAbsentFields(t *testing.T) {
	c, err := NewSwitchConfigCollector(prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}

	c.UpdateMetrics(client.SwitchGetConfigResponse{ID: 0, InitialState: "restore_last"}, "AA:BB")

	absent := map[string]*prometheus.GaugeVec{
		"auto_on": c.AutoOn, "auto_off": c.AutoOff,
		"recover_voltage_errors": c.RecoverVoltageErrors,
		"power_limit":            c.PowerLimit, "voltage_limit": c.VoltageLimit,
		"current_limit": c.CurrentLimit,
	}
	for name, vec := range absent {
		if count := testutil.CollectAndCount(vec); count != 0 {
			t.Errorf("%s series = %d, want 0 for a device that omits the field", name, count)
		}
	}
}

func TestSwitchConfigInitialStateMapping(t *testing.T) {
	cases := map[string]float64{
		"on":           1,
		"off":          0,
		"restore_last": 2,
		"match_input":  3,
		"something":    -1,
	}

	for state, want := range cases {
		c, err := NewSwitchConfigCollector(prometheus.NewRegistry())
		if err != nil {
			t.Fatalf("constructor failed: %v", err)
		}
		c.UpdateMetrics(client.SwitchGetConfigResponse{ID: 0, InitialState: state}, "AA:BB")

		if got := testutil.ToFloat64(c.InitialState.WithLabelValues("AA:BB", "0", "")); got != want {
			t.Errorf("initial_state %q = %v, want %v", state, got, want)
		}
	}
}

func TestShellyConfigSkipsEmptyEthMode(t *testing.T) {
	c, err := NewShellyConfigCollector(prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}

	c.UpdateMetrics(client.ShellyGetConfigResponse{})

	if count := testutil.CollectAndCount(c.EthIPv4Mode); count != 0 {
		t.Errorf("eth ipv4 mode series = %d, want 0 when no mode is configured", count)
	}
}

func assertFamily(t *testing.T, reg *prometheus.Registry, name string) {
	t.Helper()
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}
	for _, family := range families {
		if family.GetName() == name {
			return
		}
	}
	t.Errorf("metric family %q not found", name)
}
