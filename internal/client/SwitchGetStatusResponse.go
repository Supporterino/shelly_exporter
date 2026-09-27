package client

type SwitchGetStatusResponse struct {
	ID          int          `json:"id"`
	Source      string       `json:"source"`
	Output      bool         `json:"output"`
	Apower      *float64     `json:"apower"`
	Voltage     *float64     `json:"voltage"`
	Current     *float64     `json:"current"`
	Freq        *float64     `json:"freq"`
	Aenergy     *Energy      `json:"aenergy"`
	RetAenergy  *Energy      `json:"ret_aenergy"`
	Temperature *Temperature `json:"temperature"`
}
