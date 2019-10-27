package config

type KitConfig struct {
	Module   string   `json:"module"`
	Services []string `json:"services"`
}
