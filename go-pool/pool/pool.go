package pool

type Config struct {
	Address                 string  `json:"address"`
	BypassAddressValidation bool    `json:"bypassAddressValidation"`
	Stratum                 Stratum `json:"stratum"`
	Daemon                  Daemon  `json:"daemon"`
	Redis                   Redis   `json:"redis"`

	Threads int    `json:"threads"`
	Coin    string `json:"coin"`

	Policy Policy `json:"policy"`

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

type Redis struct {
	Endpoint string `json:"endpoint"`
	Password string `json:"password"`
	Database int64  `json:"database"`
	PoolSize int    `json:"poolSize"`
}

type Policy struct {
	Workers         int     `json:"workers"`
	Banning         Banning `json:"banning"`
	Limits          Limits  `json:"limits"`
	ResetInterval   string  `json:"resetInterval"`
	RefreshInterval string  `json:"refreshInterval"`
}

type Banning struct {
	Enabled        bool    `json:"enabled"`
	IPSet          string  `json:"ipset"`
	Timeout        int64   `json:"timeout"`
	InvalidPercent float32 `json:"invalidPercent"`
	CheckThreshold uint32  `json:"checkThreshold"`
	MalformedLimit uint32  `json:"malformedLimit"`
}

type Limits struct {
	Enabled   bool   `json:"enabled"`
	Limit     int32  `json:"limit"`
	Grace     string `json:"grace"`
	LimitJump int32  `json:"limitJump"`
}
