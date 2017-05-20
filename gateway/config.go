package gateway

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/influxdata/influxdb/services/udp"
)

type Config struct {
	Sender SenderConfig
}

type SenderConfig struct {
	Addr        string       `toml:"addr"`
	Username    string       `toml:"username"`
	Password    string       `toml:"password"`
	UserAgent   string       `toml:"user-agent"`
	Timeout     int          `toml:"timeout"`
	Gzip        bool         `toml:"gzip"`
	Precision   string       `toml:"precision"`
	Consistency string       `toml:"consistency"`
	UDPs        []udp.Config `toml:"udp"`
}

func LoadConfig(path string) (c Config, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = toml.Unmarshal(data, &c)
	return
}
