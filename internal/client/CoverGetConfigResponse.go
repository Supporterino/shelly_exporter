package client

type CoverGetConfigResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Motor struct {
		IdlePowerThr      float64 `json:"idle_power_thr"`
		IdleConfirmPeriod float64 `json:"idle_confirm_period"`
	} `json:"motor"`
	MaxtimeOpen      float64 `json:"maxtime_open"`
	MaxtimeClose     float64 `json:"maxtime_close"`
	InitialState     string  `json:"initial_state"`
	InvertDirections bool    `json:"invert_directions"`
	InMode           string  `json:"in_mode"`
	SwapInputs       bool    `json:"swap_inputs"`
	SafetySwitch     struct {
		Enable      bool   `json:"enable"`
		Direction   string `json:"direction"`
		Action      string `json:"action"`
		AllowedMove any    `json:"allowed_move"`
	} `json:"safety_switch"`
	PowerLimit           float64 `json:"power_limit"`
	VoltageLimit         float64 `json:"voltage_limit"`
	UndervoltageLimit    float64 `json:"undervoltage_limit"`
	CurrentLimit         float64 `json:"current_limit"`
	ObstructionDetection struct {
		Enable    bool    `json:"enable"`
		Direction string  `json:"direction"`
		Action    string  `json:"action"`
		PowerThr  float64 `json:"power_thr"`
		Holdoff   float64 `json:"holdoff"`
	} `json:"obstruction_detection"`
}
