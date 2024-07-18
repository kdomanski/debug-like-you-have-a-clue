package main

import (
	"io/ioutil"
	"os"
	"strings"

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

const logLevelMountPath = "/conf/log_level"
const logDebugModulesMountPath = "/conf/log_debug_modules"

func initLogging() {
	go watchLogLevel(logrus.StandardLogger(), "main")

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
}

func isModuleInDebugList(moduleName string) bool {
	data, err := ioutil.ReadFile(logDebugModulesMountPath)
	if err != nil {
		logrus.Panicf("reading log debug modules from %q: %s", logDebugModulesMountPath, err)
	}
	debugModules := strings.Split(string(data), ",")

	for _, module := range debugModules {
		if moduleName == module {
			return true
		}
	}
	return false
}

func readLogLevelFromMount() logrus.Level {
	data, err := ioutil.ReadFile(logLevelMountPath)
	if err != nil {
		logrus.Panicf("reading log level from %q: %s", logLevelMountPath, err)
	}

	lvl, err := logrus.ParseLevel(string(data))
	if err != nil {
		logrus.Panicf("parsing log level %q: %s", string(data), err)
	}

	return lvl
}

func setLogLevelFromMount(logger *logrus.Logger, moduleName string) logrus.Level {
	lvl := readLogLevelFromMount()

	if lvl >= logrus.DebugLevel && moduleName != "" && !isModuleInDebugList(moduleName) {
		// If the module is not in the debug list, set the log level to Info
		lvl = logrus.InfoLevel
	}

	logger.SetLevel(lvl)
	return lvl
}

func watchLogLevel(logger *logrus.Logger, moduleName string) {
	setLogLevelFromMount(logger, moduleName)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Panic("creating fs watcher", err)
	}
	defer watcher.Close()

	err = watcher.Add(logLevelMountPath)
	if err != nil {
		logrus.Panicf("watching path %q: %s", logLevelMountPath, err)
	}
	err = watcher.Add(logDebugModulesMountPath)
	if err != nil {
		logrus.Panicf("watching path %q: %s", logDebugModulesMountPath, err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Remove) {
				err = watcher.Add(logLevelMountPath)
				if err != nil {
					logrus.Panicf("watching path %q: %s", logLevelMountPath, err)
				}
				err = watcher.Add(logDebugModulesMountPath)
				if err != nil {
					logrus.Panicf("watching path %q: %s", logDebugModulesMountPath, err)
				}

				setLogLevelFromMount(logger, moduleName)
				logrus.Warnf("updated log level watchers for module %q", moduleName)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logrus.Warnf("error: %+v", err)
		}
	}
}

func newCustomLogger(moduleName string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	logger.SetOutput(os.Stdout)

	go watchLogLevel(logger, moduleName)
	return logger
}
