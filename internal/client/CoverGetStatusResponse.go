package client

type CoverGetStatusResponse struct {
	ID            int          `json:"id"`
	Source        string       `json:"source"`
	State         string       `json:"state"`
	Apower        *float64     `json:"apower"`
	Voltage       *float64     `json:"voltage"`
	Current       *float64     `json:"current"`
	Pf            *float64     `json:"pf"`
	Freq          *float64     `json:"freq"`
	Aenergy       *Energy      `json:"aenergy"`
	Temperature   *Temperature `json:"temperature"`
	PosControl    *bool        `json:"pos_control"`
	LastDirection string       `json:"last_direction"`
	CurrentPos    *int         `json:"current_pos"`
}
