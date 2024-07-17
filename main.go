package main

import (
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

type DataEntry struct {
	source int
	data   string
}

func loadDataDumpEntries(source int, dataChan chan<- DataEntry) {
	log := logrus.WithFields(logrus.Fields{
		"source": source,
	})
	for {
		ms_sleep := time.Duration(200 + rand.Intn(500))
		time.Sleep(ms_sleep * time.Millisecond)
		log.Infof("successfully loaded entry")
		datum := RandStringRunes(12)
		dataChan <- DataEntry{
			source: source,
			data:   datum,
		}
	}
}

func processDataEntries(dataChan <-chan DataEntry) {
	for datum := range dataChan {
		log := logrus.WithFields(logrus.Fields{
			"source": datum.source,
		})

		log.Debugf("decoding datum %s\n", datum.data)
		time.Sleep(10 * time.Millisecond)
		log.Debugf("calculating something based on datum %s\n", datum.data)
		time.Sleep(35 * time.Millisecond)
		log.Debugf("comitting results for %s to memory\n", datum.data)
		time.Sleep(15 * time.Millisecond)
	}
}

const numOfProducers = 5
const numOfConsumers = 5

func startProcessing(numOfProducers, numOfConsumers int) {
	dataChan := make(chan DataEntry)

	for i := 0; i < numOfProducers; i++ {
		go loadDataDumpEntries(i, dataChan)
	}

	for i := 0; i < numOfConsumers; i++ {
		go processDataEntries(dataChan)
	}

}

func main() {
	initLogging()

	startProcessing(numOfProducers, numOfConsumers)

	// Block main goroutine forever.
	<-make(chan struct{})
}
