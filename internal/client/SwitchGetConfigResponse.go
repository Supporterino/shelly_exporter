package client

type SwitchGetConfigResponse struct {
	ID                       int      `json:"id"`
	Name                     string   `json:"name"`
	InMode                   string   `json:"in_mode"`
	InitialState             string   `json:"initial_state"`
	AutoOn                   *bool    `json:"auto_on"`
	AutoOnDelay              *float64 `json:"auto_on_delay"`
	AutoOff                  *bool    `json:"auto_off"`
	AutoOffDelay             *float64 `json:"auto_off_delay"`
	AutorecoverVoltageErrors *bool    `json:"autorecover_voltage_errors"`
	PowerLimit               *float64 `json:"power_limit"`
	VoltageLimit             *float64 `json:"voltage_limit"`
	UndervoltageLimit        *float64 `json:"undervoltage_limit"`
	CurrentLimit             *float64 `json:"current_limit"`
}
