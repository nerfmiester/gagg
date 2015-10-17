package Config

import "time"

// TomlConfig has the stuff we need
type TomlConfig struct {
	Title   string
	Owner   ownerInfo
	Agg     agg   `toml:"agg"`
	Debug   debug `toml:"debug"`
	Servers map[string]server
	Clients clients
}

type ownerInfo struct {
	Name string
	Org  string `toml:"organization"`
	Bio  string
	DOB  time.Time
}

type agg struct {
	Server  string
	Ports   []int
	ConnMax int `toml:"connection_max"`
	Enabled bool
}

type server struct {
	IP string
	DC string
}

type clients struct {
	Data  [][]interface{}
	Hosts []string
}

type debug struct {
	Enabled bool
}
