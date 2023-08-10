package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
)

func LoadConfig(configPath string) (*Config, error) {

	fi, err := os.Open(configPath)
	if err != nil {
		log.Errorf("Error: %s", err)
		return nil, err
	}
	defer fi.Close()
	buf, err := ioutil.ReadAll(fi)
	if err != nil {
		log.Errorf("Error: %s", err)
		return nil, err
	}

	cfg := &Config{}
	err = json.Unmarshal(buf, cfg)
	if err != nil {
		log.Errorf("Error: %s", err)
		return nil, err
	}
	return cfg, err
}
