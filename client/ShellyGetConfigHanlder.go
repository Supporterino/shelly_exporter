package client

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (r *ShellyGetConfigResponse) UnmarshalJSON(data []byte) error {
	// Parse JSON into a generic map
	raw := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Initialize dynamic maps
	r.Inputs = make(map[string]ShellyGetConfigResponseInput)
	r.Switches = make(map[string]ShellyGetConfigResponseSwitch)

	// Iterate through raw data
	for key, value := range raw {
		switch {
		case strings.HasPrefix(key, "input:"):
			// Parse dynamic input keys
			var input ShellyGetConfigResponseInput
			if err := json.Unmarshal(value, &input); err != nil {
				return fmt.Errorf("failed to unmarshal input '%s': %w", key, err)
			}
			r.Inputs[key] = input

		case strings.HasPrefix(key, "switch:"):
			// Parse dynamic switch keys
			var sw ShellyGetConfigResponseSwitch
			if err := json.Unmarshal(value, &sw); err != nil {
				return fmt.Errorf("failed to unmarshal switch '%s': %w", key, err)
			}
			r.Switches[key] = sw

		case key == "ble":
			if err := json.Unmarshal(value, &r.BLE); err != nil {
				return fmt.Errorf("failed to unmarshal BLE: %w", err)
			}

		case key == "cloud":
			if err := json.Unmarshal(value, &r.Cloud); err != nil {
				return fmt.Errorf("failed to unmarshal cloud: %w", err)
			}

		case key == "eth":
			if err := json.Unmarshal(value, &r.Eth); err != nil {
				return fmt.Errorf("failed to unmarshal eth: %w", err)
			}

		case key == "mqtt":
			if err := json.Unmarshal(value, &r.MQTT); err != nil {
				return fmt.Errorf("failed to unmarshal MQTT: %w", err)
			}

		case key == "sys":
			if err := json.Unmarshal(value, &r.Sys); err != nil {
				return fmt.Errorf("failed to unmarshal sys: %w", err)
			}

		case key == "wifi":
			if err := json.Unmarshal(value, &r.Wifi); err != nil {
				return fmt.Errorf("failed to unmarshal wifi: %w", err)
			}

		default:
			// Ignore unknown keys
		}
	}

	return nil
}
