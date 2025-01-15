package client

type ShellyGetDeviceInfoResponse struct {
	Name       string `json:"name"`        // Device name
	ID         string `json:"id"`          // Device ID
	MAC        string `json:"mac"`         // MAC address
	Slot       int    `json:"slot"`        // Slot number
	Model      string `json:"model"`       // Device model
	Gen        int    `json:"gen"`         // Generation (e.g., Gen 2)
	FwID       string `json:"fw_id"`       // Firmware ID
	Version    string `json:"ver"`         // Firmware version
	App        string `json:"app"`         // Application name
	AuthEn     bool   `json:"auth_en"`     // Indicates if authentication is enabled
	AuthDomain string `json:"auth_domain"` // Authentication domain, nullable
}
