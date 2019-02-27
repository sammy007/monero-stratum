package pool

type Config struct {
	Address                 string     `json:"address"`
	BypassAddressValidation bool       `json:"bypassAddressValidation"`
	BypassShareValidation   bool       `json:"bypassShareValidation"`
	Stratum                 Stratum    `json:"stratum"`
	BlockRefreshInterval    string     `json:"blockRefreshInterval"`
	UpstreamCheckInterval   string     `json:"upstreamCheckInterval"`
	Upstream                []Upstream `json:"upstream"`
	EstimationWindow        string     `json:"estimationWindow"`
	LuckWindow              string     `json:"luckWindow"`
	LargeLuckWindow         string     `json:"largeLuckWindow"`
	Threads                 int        `json:"threads"`
	Frontend                Frontend   `json:"frontend"`
	NewrelicName            string     `json:"newrelicName"`
	NewrelicKey             string     `json:"newrelicKey"`
	NewrelicVerbose         bool       `json:"newrelicVerbose"`
	NewrelicEnabled         bool       `json:"newrelicEnabled"`
}

type Stratum struct {
	Timeout string `json:"timeout"`
	Ports   []Port `json:"listen"`
}

type Port struct {
	Difficulty int64  `json:"diff"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	MaxConn    int    `json:"maxConn"`
}

type Upstream struct {
	Name    string `json:"name"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Timeout string `json:"timeout"`
}

type Frontend struct {
	Enabled  bool   `json:"enabled"`
	Listen   string `json:"listen"`
	Login    string `json:"login"`
	Password string `json:"password"`
	HideIP   bool   `json:"hideIP"`
}
