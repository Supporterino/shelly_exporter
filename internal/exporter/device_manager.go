package exporter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/supporterino/shelly_exporter/internal/client"
)

const deviceRequestTimeout = 10 * time.Second

// FetcherFactory builds a Fetcher for a configured device.
type FetcherFactory func(host, username, password string) client.Fetcher

// DefaultFetcherFactory builds an APIClient backed by a real HTTP client.
func DefaultFetcherFactory(host, username, password string) client.Fetcher {
	return client.NewAPIClient(
		host,
		client.WithCredentials(username, password),
		client.WithTimeout(deviceRequestTimeout),
	)
}

type modelHandler func(ctx context.Context, dm *DeviceManager, d *deviceState, cfg client.ShellyGetConfigResponse) error

// modelHandlers maps a device model to its metric handler.
var modelHandlers = map[string]modelHandler{
	"Plus2PM":   handlePlus2PM,
	"PlusPlugS": handleSwitchDevice,
	"Mini1G3":   handleSwitchDevice,
	"Pro4PM":    handleSwitchDevice,
}

// DeviceManager owns the polling loops for all registered devices.
type DeviceManager struct {
	interval   time.Duration
	collectors *Collectors
	newFetcher FetcherFactory
	logger     *slog.Logger

	mu      sync.Mutex
	devices map[string]context.CancelFunc
	wg      sync.WaitGroup
}

// NewDeviceManager creates a DeviceManager.
func NewDeviceManager(interval time.Duration, collectors *Collectors, newFetcher FetcherFactory, logger *slog.Logger) *DeviceManager {
	if logger == nil {
		logger = slog.Default()
	}
	return &DeviceManager{
		interval:   interval,
		collectors: collectors,
		newFetcher: newFetcher,
		logger:     logger,
		devices:    make(map[string]context.CancelFunc),
	}
}

// Register starts a polling loop for the device. A host that is already
// registered is ignored.
func (dm *DeviceManager) Register(ctx context.Context, device Device) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if _, exists := dm.devices[device.Host]; exists {
		dm.logger.Warn("Device already registered", slog.String("host", device.Host))
		return
	}

	dm.logger.Info("Registering new device", slog.String("host", device.Host))

	loopCtx, cancel := context.WithCancel(ctx)
	dm.devices[device.Host] = cancel

	state := &deviceState{
		host:    device.Host,
		fetcher: dm.newFetcher(device.Host, device.Username, device.Password),
	}

	dm.wg.Add(1)
	go func() {
		defer dm.wg.Done()
		dm.runLoop(loopCtx, state)
	}()
}

// Wait blocks until every polling loop has stopped.
func (dm *DeviceManager) Wait() {
	dm.wg.Wait()
}

func (dm *DeviceManager) runLoop(ctx context.Context, d *deviceState) {
	if err := dm.poll(ctx, d); err != nil {
		dm.logger.Error("Error fetching metrics", slog.Any("error", err), slog.String("host", d.host))
	}

	ticker := time.NewTicker(dm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			dm.logger.Info("Stopping metrics update loop", slog.String("host", d.host))
			return
		case <-ticker.C:
			if err := dm.poll(ctx, d); err != nil {
				dm.logger.Error("Error fetching metrics", slog.Any("error", err), slog.String("host", d.host))
			}
		}
	}
}

func (dm *DeviceManager) poll(ctx context.Context, d *deviceState) error {
	var errs []error

	info, err := dm.collectors.DeviceInfo.Update(ctx, d.fetcher)
	if err != nil {
		errs = append(errs, err)
	} else {
		dm.applyDeviceInfo(d, info)
	}

	if err := dm.collectors.ShellyStatus.Update(ctx, d.fetcher); err != nil {
		errs = append(errs, err)
	}
	cfg, cfgErr := dm.collectors.ShellyConfig.Update(ctx, d.fetcher)
	if cfgErr != nil {
		errs = append(errs, cfgErr)
	}

	if d.model != "" {
		handler, ok := modelHandlers[d.model]
		switch {
		case !ok:
			dm.logger.Warn("No model-specific handler registered", slog.String("host", d.host), slog.String("model", d.model))
		case cfgErr == nil:
			if err := handler(ctx, dm, d, cfg); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if err := errors.Join(errs...); err != nil {
		dm.collectors.Up.Set(0, d.mac, d.host)
		return err
	}

	dm.collectors.Up.Set(1, d.mac, d.host)
	return nil
}

func (dm *DeviceManager) applyDeviceInfo(d *deviceState, info client.ShellyGetDeviceInfoResponse) {
	previousMAC := d.mac
	d.mac = info.Mac
	d.model = info.App
	d.profile = info.Profile

	if d.mac != previousMAC {
		dm.collectors.Up.Delete(previousMAC, d.host)
	}
}

// handlePlus2PM dispatches a Plus2PM to the cover or switch handler based on
// the configured profile, falling back to the components present in the config.
func handlePlus2PM(ctx context.Context, dm *DeviceManager, d *deviceState, cfg client.ShellyGetConfigResponse) error {
	switch {
	case d.profile == "cover":
		return handleCoverDevice(ctx, dm, d, cfg)
	case d.profile == "switch":
		return handleSwitchDevice(ctx, dm, d, cfg)
	case len(cfg.CoverIDs()) > 0:
		return handleCoverDevice(ctx, dm, d, cfg)
	default:
		return handleSwitchDevice(ctx, dm, d, cfg)
	}
}

func handleCoverDevice(ctx context.Context, dm *DeviceManager, d *deviceState, cfg client.ShellyGetConfigResponse) error {
	var errs []error
	for _, id := range coverComponentIDs(cfg) {
		if err := dm.collectors.CoverStatus.Update(ctx, d.fetcher, id, coverName(cfg, id), d.mac); err != nil {
			errs = append(errs, err)
		}
	}
	if err := dm.collectors.WiFiStatus.Update(ctx, d.fetcher, d.mac); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func handleSwitchDevice(ctx context.Context, dm *DeviceManager, d *deviceState, cfg client.ShellyGetConfigResponse) error {
	var errs []error
	for _, id := range switchComponentIDs(cfg) {
		if err := dm.collectors.SwitchStatus.Update(ctx, d.fetcher, id, switchName(cfg, id), d.mac); err != nil {
			errs = append(errs, err)
		}
		if err := dm.collectors.SwitchConfig.Update(ctx, d.fetcher, id, d.mac); err != nil {
			errs = append(errs, err)
		}
	}
	if err := dm.collectors.WiFiStatus.Update(ctx, d.fetcher, d.mac); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// switchName returns the configured name of a switch component, or "" when the
// device does not expose one.
func switchName(cfg client.ShellyGetConfigResponse, id int) string {
	if sw, ok := cfg.Switches[fmt.Sprintf("switch:%d", id)]; ok && sw.Name != nil {
		return *sw.Name
	}
	return ""
}

// coverName returns the configured name of a cover component, or "" when the
// device does not expose one.
func coverName(cfg client.ShellyGetConfigResponse, id int) string {
	if cover, ok := cfg.Covers[fmt.Sprintf("cover:%d", id)]; ok {
		return cover.Name
	}
	return ""
}

// switchComponentIDs returns the switch component IDs discovered in the config,
// defaulting to the first switch when none can be discovered.
func switchComponentIDs(cfg client.ShellyGetConfigResponse) []int {
	if ids := cfg.SwitchIDs(); len(ids) > 0 {
		return ids
	}
	return []int{0}
}

// coverComponentIDs returns the cover component IDs discovered in the config,
// defaulting to the first cover when none can be discovered.
func coverComponentIDs(cfg client.ShellyGetConfigResponse) []int {
	if ids := cfg.CoverIDs(); len(ids) > 0 {
		return ids
	}
	return []int{0}
}
