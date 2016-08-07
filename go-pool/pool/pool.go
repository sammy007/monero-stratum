package pool

type Config struct {
	Address                 string  `json:"address"`
	BypassAddressValidation bool    `json:"bypassAddressValidation"`
	Stratum                 Stratum `json:"stratum"`
	Daemon                  Daemon  `json:"daemon"`

	Threads int    `json:"threads"`
	Coin    string `json:"coin"`

	NewrelicName    string `json:"newrelicName"`
	NewrelicKey     string `json:"newrelicKey"`
	NewrelicVerbose bool   `json:"newrelicVerbose"`
	NewrelicEnabled bool   `json:"newrelicEnabled"`
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
