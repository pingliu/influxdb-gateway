package gateway

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Sender SenderConfig `toml:"sender"`
}

func LoadConfig(path string) (c Config, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = toml.Unmarshal(data, &c)
	return
}
