package client

type ShellyGetConfigResponse struct {
	BLE      BLE   `json:"ble"`
	Cloud    Cloud `json:"cloud"`
	Eth      Eth   `json:"eth"`
	Inputs   map[string]ShellyGetConfigResponseInput
	Switches map[string]ShellyGetConfigResponseSwitch
	MQTT     MQTT                        `json:"mqtt"`
	Sys      ShellyGetConfigResponseSys  `json:"sys"`
	Wifi     ShellyGetConfigResponseWifi `json:"wifi"`
}

type BLE struct {
	Enable bool `json:"enable"`
}

type Cloud struct {
	Enable bool   `json:"enable"`
	Server string `json:"server"`
}

type Eth struct {
	Enable     bool    `json:"enable"`
	IPv4Mode   string  `json:"ipv4mode"`
	IP         *string `json:"ip"`
	Netmask    *string `json:"netmask"`
	Gateway    *string `json:"gw"`
	Nameserver *string `json:"nameserver"`
}

type ShellyGetConfigResponseInput struct {
	ID     int     `json:"id"`
	Name   *string `json:"name"`
	Type   string  `json:"type"`
	Invert bool    `json:"invert"`
}

type ShellyGetConfigResponseSwitch struct {
	ID           int     `json:"id"`
	Name         *string `json:"name"`
	InMode       string  `json:"in_mode"`
	InitialState string  `json:"initial_state"`
	AutoOn       bool    `json:"auto_on"`
	AutoOnDelay  float64 `json:"auto_on_delay"`
	AutoOff      bool    `json:"auto_off"`
	AutoOffDelay float64 `json:"auto_off_delay"`
	PowerLimit   float64 `json:"power_limit"`
	VoltageLimit float64 `json:"voltage_limit"`
	CurrentLimit float64 `json:"current_limit"`
}

type MQTT struct {
	Enable bool    `json:"enable"`
	Server *string `json:"server"`
	User   *string `json:"user"`
	Pass   *string `json:"pass"`
}

type ShellyGetConfigResponseSys struct {
	Device struct {
		Name *string `json:"name"`
		MAC  string  `json:"mac"`
		FWID string  `json:"fw_id"`
	} `json:"device"`
	Location struct {
		TZ  string  `json:"tz"`
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"location"`
	UIData map[string]interface{} `json:"ui_data"`
	SNTP   struct {
		Server string `json:"server"`
	} `json:"sntp"`
}

type ShellyGetConfigResponseWifi struct {
	AP struct {
		SSID   string `json:"ssid"`
		IsOpen bool   `json:"is_open"`
		Enable bool   `json:"enable"`
	} `json:"ap"`
	STA struct {
		SSID       *string `json:"ssid"`
		IsOpen     bool    `json:"is_open"`
		Enable     bool    `json:"enable"`
		IPv4Mode   string  `json:"ipv4mode"`
		IP         *string `json:"ip"`
		Netmask    *string `json:"netmask"`
		Gateway    *string `json:"gw"`
		Nameserver *string `json:"nameserver"`
	} `json:"sta"`
	STA1 struct {
		SSID       *string `json:"ssid"`
		IsOpen     bool    `json:"is_open"`
		Enable     bool    `json:"enable"`
		IPv4Mode   string  `json:"ipv4mode"`
		IP         *string `json:"ip"`
		Netmask    *string `json:"netmask"`
		Gateway    *string `json:"gw"`
		Nameserver *string `json:"nameserver"`
	} `json:"sta1"`
	WS struct {
		Enable bool    `json:"enable"`
		Server *string `json:"server"`
		SSLCA  string  `json:"ssl_ca"`
	} `json:"ws"`
	Roam struct {
		RSSIThreshold int `json:"rssi_thr"`
		Interval      int `json:"interval"`
	} `json:"roam"`
}
