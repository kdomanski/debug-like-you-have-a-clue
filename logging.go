package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

func setLogLevelFromEnv(env string) {
	if logEnv := os.Getenv(env); logEnv != "" {
		lvl, err := logrus.ParseLevel(logEnv)
		if err != nil {
			logrus.Panicf("parsing log level %q: %s", logEnv, err)
		}
		logrus.SetLevel(lvl)
	}
}

func initLogging() {
	setLogLevelFromEnv("LOG_LEVEL")

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
}
