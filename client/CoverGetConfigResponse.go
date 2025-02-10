package client

type CoverGetConfigResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Motor struct {
		IdlePowerThr      int     `json:"idle_power_thr"`
		IdleConfirmPeriod float64 `json:"idle_confirm_period"`
	} `json:"motor"`
	MaxtimeOpen      int    `json:"maxtime_open"`
	MaxtimeClose     int    `json:"maxtime_close"`
	InitialState     string `json:"initial_state"`
	InvertDirections bool   `json:"invert_directions"`
	InMode           string `json:"in_mode"`
	SwapInputs       bool   `json:"swap_inputs"`
	SafetySwitch     struct {
		Enable      bool        `json:"enable"`
		Direction   string      `json:"direction"`
		Action      string      `json:"action"`
		AllowedMove interface{} `json:"allowed_move"`
	} `json:"safety_switch"`
	PowerLimit           int `json:"power_limit"`
	VoltageLimit         int `json:"voltage_limit"`
	UndervoltageLimit    int `json:"undervoltage_limit"`
	CurrentLimit         int `json:"current_limit"`
	ObstructionDetection struct {
		Enable    bool   `json:"enable"`
		Direction string `json:"direction"`
		Action    string `json:"action"`
		PowerThr  int    `json:"power_thr"`
		Holdoff   int    `json:"holdoff"`
	} `json:"obstruction_detection"`
}
