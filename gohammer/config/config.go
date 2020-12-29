package config

type Config struct {
	Nodes []NodeConfig `json:"nodes"`
}

type NodeConfig struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
