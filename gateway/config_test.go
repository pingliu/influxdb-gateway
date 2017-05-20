package gateway

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	c, err := LoadConfig("sample.toml")
	if err != nil {
		t.Error(err)
	} else if c.Sender.Addr == "" {
		t.Error("addr is null")
	}
}
