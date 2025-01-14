package client

type ShellyGetStatusResponse struct {
	BLE   map[string]interface{} `json:"ble"` // No defined structure, use a generic map
	Cloud struct {
		Connected bool `json:"connected"`
	} `json:"cloud"`
	MQTT struct {
		Connected bool `json:"connected"`
	} `json:"mqtt"`
	PlugsUI map[string]interface{} `json:"plugs_ui"` // No defined structure, use a generic map
	Switch0 struct {
		ID      int     `json:"id"`
		Source  string  `json:"source"`
		Output  bool    `json:"output"`
		APower  float64 `json:"apower"`
		Voltage float64 `json:"voltage"`
		Current float64 `json:"current"`
		AEnergy struct {
			Total    float64   `json:"total"`
			ByMinute []float64 `json:"by_minute"`
			MinuteTS int64     `json:"minute_ts"`
		} `json:"aenergy"`
		Temperature struct {
			TC float64 `json:"tC"`
			TF float64 `json:"tF"`
		} `json:"temperature"`
	} `json:"switch:0"`
	Sys struct {
		MAC              string `json:"mac"`
		RestartRequired  bool   `json:"restart_required"`
		Time             string `json:"time"`
		UnixTime         int64  `json:"unixtime"`
		Uptime           int64  `json:"uptime"`
		RAMSize          int64  `json:"ram_size"`
		RAMFree          int64  `json:"ram_free"`
		FSSize           int64  `json:"fs_size"`
		FSFree           int64  `json:"fs_free"`
		CfgRev           int    `json:"cfg_rev"`
		KvsRev           int    `json:"kvs_rev"`
		ScheduleRev      int    `json:"schedule_rev"`
		WebhookRev       int    `json:"webhook_rev"`
		AvailableUpdates struct {
			Beta struct {
				Version string `json:"version"`
			} `json:"beta"`
		} `json:"available_updates"`
		ResetReason int `json:"reset_reason"`
	} `json:"sys"`
	WiFi struct {
		STAIP  string `json:"sta_ip"`
		Status string `json:"status"`
		SSID   string `json:"ssid"`
		RSSI   int    `json:"rssi"`
	} `json:"wifi"`
	WS struct {
		Connected bool `json:"connected"`
	} `json:"ws"`
}
