package config

import (
	"encoding/json"
	"io"
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
	buf, err := io.ReadAll(fi)
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
	// provide env variable support for endpoint url
	for _, csmr := range cfg.Consumers {
		if csmr.Endpoint[0] == '$' {
			log.Infof("Parsing endpoint url %v...", csmr.Endpoint)
			url, ok := os.LookupEnv(csmr.Endpoint[1:])
			if ok {
				log.Infof("endpoint url %v from %v...", url, csmr.Endpoint)
				csmr.Endpoint = url
			}
		}
	}
	return cfg, err
}
