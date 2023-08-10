package utils

import (
	log "github.com/sirupsen/logrus"
)

func SetLogLevel(level string) {
	lvl, err := log.ParseLevel(level)
	if err != nil {
		lvl = log.InfoLevel
	}
	log.SetLevel(lvl)
}
