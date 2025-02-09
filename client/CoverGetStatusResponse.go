package client

type CoverGetStatusResponse struct {
	ID      int     `json:"id"`
	Source  string  `json:"source"`
	State   string  `json:"state"`
	Apower  float64 `json:"apower"`
	Voltage float64 `json:"voltage"`
	Current float64 `json:"current"`
	Pf      float64 `json:"pf"`
	Freq    float64 `json:"freq"`
	Aenergy struct {
		Total    float64 `json:"total"`
		ByMinute []int   `json:"by_minute"`
		MinuteTs int     `json:"minute_ts"`
	} `json:"aenergy"`
	Temperature struct {
		TC float64 `json:"tC"`
		TF float64 `json:"tF"`
	} `json:"temperature"`
	PosControl    bool   `json:"pos_control"`
	LastDirection string `json:"last_direction"`
	CurrentPos    int    `json:"current_pos"`
}
