package client

type Energy struct {
	Total    float64   `json:"total"`
	ByMinute []float64 `json:"by_minute"`
	MinuteTS int64     `json:"minute_ts"`
}

type Temperature struct {
	TC float64 `json:"tC"`
	TF float64 `json:"tF"`
}

type Sys struct {
	MAC              string `json:"mac"`
	RestartRequired  bool   `json:"restart_required"`
	Time             string `json:"time"`
	Unixtime         int64  `json:"unixtime"`
	LastSyncTS       int64  `json:"last_sync_ts"`
	Uptime           int64  `json:"uptime"`
	RAMSize          int    `json:"ram_size"`
	RAMFree          int    `json:"ram_free"`
	FSSize           int    `json:"fs_size"`
	FSFree           int    `json:"fs_free"`
	CfgRev           int    `json:"cfg_rev"`
	KvsRev           int    `json:"kvs_rev"`
	ScheduleRev      int    `json:"schedule_rev"`
	WebhookRev       int    `json:"webhook_rev"`
	AvailableUpdates struct {
		Stable struct {
			Version string `json:"version"`
		} `json:"stable"`
	} `json:"available_updates"`
}

type Wifi struct {
	StaIP  *string `json:"sta_ip"`
	Status string  `json:"status"`
	SSID   *string `json:"ssid"`
	RSSI   int     `json:"rssi"`
}

type ShellyGetStatusResponse struct {
	BLE   map[string]interface{}   `json:"ble"`
	Cloud struct{ Connected bool } `json:"cloud"`
	Eth   struct{ IP string }      `json:"eth"`
	MQTT  struct{ Connected bool } `json:"mqtt"`
	Sys   Sys                      `json:"sys"`
	Wifi  Wifi                     `json:"wifi"`
	WS    struct{ Connected bool } `json:"ws"`
}
