package client

type ShellyGetConfigResponse struct {
	BLE struct {
		Enable bool `json:"enable"`
		RPC    struct {
			Enable bool `json:"enable"`
		} `json:"rpc"`
		Observer struct {
			Enable bool `json:"enable"`
		} `json:"observer"`
	} `json:"ble"`
	Cloud struct {
		Enable bool   `json:"enable"`
		Server string `json:"server"`
	} `json:"cloud"`
	MQTT struct {
		Enable              bool   `json:"enable"`
		Server              string `json:"server"`
		ClientID            string `json:"client_id"`
		User                string `json:"user"`
		SSLCA               string `json:"ssl_ca"`
		TopicPrefix         string `json:"topic_prefix"`
		RPCNotifications    bool   `json:"rpc_ntf"`
		StatusNotifications bool   `json:"status_ntf"`
		UseClientCert       bool   `json:"use_client_cert"`
		EnableRPC           bool   `json:"enable_rpc"`
		EnableControl       bool   `json:"enable_control"`
	} `json:"mqtt"`
	PlugsUI struct {
		LEDs struct {
			Mode   string `json:"mode"`
			Colors map[string]struct {
				On struct {
					RGB        []float64 `json:"rgb"`
					Brightness float64   `json:"brightness"`
				} `json:"on"`
				Off struct {
					RGB        []float64 `json:"rgb"`
					Brightness float64   `json:"brightness"`
				} `json:"off"`
			} `json:"colors"`
			Power struct {
				Brightness float64 `json:"brightness"`
			} `json:"power"`
		} `json:"leds"`
		NightMode struct {
			Enable        bool     `json:"enable"`
			Brightness    float64  `json:"brightness"`
			ActiveBetween []string `json:"active_between"`
		} `json:"night_mode"`
	} `json:"plugs_ui"`
	Controls map[string]struct {
		InMode string `json:"in_mode"`
	} `json:"controls"`
	Switch0 struct {
		ID           int     `json:"id"`
		Name         string  `json:"name"`
		InitialState string  `json:"initial_state"`
		AutoOn       bool    `json:"auto_on"`
		AutoOnDelay  float64 `json:"auto_on_delay"`
		AutoOff      bool    `json:"auto_off"`
		AutoOffDelay float64 `json:"auto_off_delay"`
		PowerLimit   int     `json:"power_limit"`
		VoltageLimit int     `json:"voltage_limit"`
		CurrentLimit float64 `json:"current_limit"`
	} `json:"switch:0"`
	Sys struct {
		Device struct {
			Name         string `json:"name"`
			MAC          string `json:"mac"`
			FWID         string `json:"fw_id"`
			Discoverable bool   `json:"discoverable"`
			EcoMode      bool   `json:"eco_mode"`
		} `json:"device"`
		Location struct {
			TZ  string  `json:"tz"`
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"location"`
		Debug struct {
			Level     int  `json:"level"`
			FileLevel *int `json:"file_level"`
			MQTT      struct {
				Enable bool `json:"enable"`
			} `json:"mqtt"`
			Websocket struct {
				Enable bool `json:"enable"`
			} `json:"websocket"`
			UDP struct {
				Addr *string `json:"addr"`
			} `json:"udp"`
		} `json:"debug"`
	} `json:"sys"`
	WiFi struct {
		AP struct {
			SSID          string `json:"ssid"`
			IsOpen        bool   `json:"is_open"`
			Enable        bool   `json:"enable"`
			RangeExtender struct {
				Enable bool `json:"enable"`
			} `json:"range_extender"`
		} `json:"ap"`
		STA struct {
			SSID       string `json:"ssid"`
			IsOpen     bool   `json:"is_open"`
			Enable     bool   `json:"enable"`
			IPv4Mode   string `json:"ipv4mode"`
			IP         string `json:"ip"`
			Netmask    string `json:"netmask"`
			GW         string `json:"gw"`
			Nameserver string `json:"nameserver"`
		} `json:"sta"`
		STA1 struct {
			SSID       string `json:"ssid"`
			IsOpen     bool   `json:"is_open"`
			Enable     bool   `json:"enable"`
			IPv4Mode   string `json:"ipv4mode"`
			IP         string `json:"ip"`
			Netmask    string `json:"netmask"`
			GW         string `json:"gw"`
			Nameserver string `json:"nameserver"`
		} `json:"sta1"`
		Roam struct {
			RSSIThreshold int `json:"rssi_thr"`
			Interval      int `json:"interval"`
		} `json:"roam"`
	} `json:"wifi"`
	WS struct {
		Enable bool   `json:"enable"`
		Server string `json:"server"`
		SSLCA  string `json:"ssl_ca"`
	} `json:"ws"`
}
