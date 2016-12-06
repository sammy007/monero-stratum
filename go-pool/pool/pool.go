package pool

type Config struct {
	Address                 string   `json:"address"`
	BypassAddressValidation bool     `json:"bypassAddressValidation"`
	BypassShareValidation   bool     `json:"bypassShareValidation"`
	Stratum                 Stratum  `json:"stratum"`
	Daemon                  Daemon   `json:"daemon"`
	EstimationWindow        string   `json:"estimationWindow"`
	LuckWindow              string   `json:"luckWindow"`
	LargeLuckWindow         string   `json:"largeLuckWindow"`
	Threads                 int      `json:"threads"`
	Frontend                Frontend `json:"frontend"`
	NewrelicName            string   `json:"newrelicName"`
	NewrelicKey             string   `json:"newrelicKey"`
	NewrelicVerbose         bool     `json:"newrelicVerbose"`
	NewrelicEnabled         bool     `json:"newrelicEnabled"`
}

type Stratum struct {
	Timeout              string `json:"timeout"`
	BlockRefreshInterval string `json:"blockRefreshInterval"`
	Ports                []Port `json:"listen"`
}

type Port struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Difficulty int64  `json:"diff"`
	MaxConn    int    `json:"maxConn"`
}

type Daemon struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Timeout string `json:"timeout"`
}

type Frontend struct {
	Listen   string `json:"listen"`
	Login    string `json:"login"`
	Password string `json:"password"`
	HideIP   bool   `json:"hideIP"`
}
