package main

import (
	"io/ioutil"
	"os"

	"github.com/fsnotify/fsnotify"
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

func initLogging(levelPath string) {
	go watchLogLevel(levelPath, logrus.StandardLogger())

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
}

func setLogLevelFromMount(path string, logger *logrus.Logger) logrus.Level {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.Panicf("reading log level from %q: %s", path, err)
	}

	lvl, err := logrus.ParseLevel(string(data))
	if err != nil {
		logrus.Panicf("parsing log level %q: %s", string(data), err)
	}
	logger.SetLevel(lvl)
	return lvl
}

func watchLogLevel(filePath string, logger *logrus.Logger) {
	setLogLevelFromMount(filePath, logger)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Panic("creating fs watcher", err)
	}
	defer watcher.Close()

	err = watcher.Add(filePath)
	if err != nil {
		logrus.Panicf("watching path %q: %s", filePath, err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Remove) {
				if err = watcher.Add(filePath); err != nil {
					logrus.Panicf("watching path %q: %s", filePath, err)
				}

				setLogLevelFromMount(filePath, logger)
				logrus.Warnf("updated watcher for: %q", filePath)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logrus.Warnf("error: %+v", err)
		}
	}
}

func newCustomLogger(logLevelMountPath string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	logger.SetOutput(os.Stdout)

	go watchLogLevel(logLevelMountPath, logger)
	return logger
}
