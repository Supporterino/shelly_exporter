package client

type WiFiGetStatusResponse struct {
	StaIP         string `json:"sta_ip"`
	Status        string `json:"status"`
	Ssid          string `json:"ssid"`
	Rssi          int    `json:"rssi"`
	ApClientCount int    `json:"ap_client_count"`
}
