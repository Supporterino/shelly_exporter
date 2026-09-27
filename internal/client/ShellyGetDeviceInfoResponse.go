package client

type ShellyGetDeviceInfoResponse struct {
	Name         string `json:"name"`
	ID           string `json:"id"`
	Mac          string `json:"mac"`
	Model        string `json:"model"`
	Gen          int    `json:"gen"`
	FwID         string `json:"fw_id"`
	Ver          string `json:"ver"`
	App          string `json:"app"`
	AuthEn       bool   `json:"auth_en"`
	AuthDomain   string `json:"auth_domain"`
	Discoverable bool   `json:"discoverable"`
	Profile      string `json:"profile"`
}
