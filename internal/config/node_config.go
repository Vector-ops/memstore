package config

type NodeConfig struct {
	Id   string `json:"id"`
	Host string `json:"host"`
	Port string `json:"port"`
	Role string `json:"role"`
}
