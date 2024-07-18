package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var procTracer trace.Tracer

type DataEntry struct {
	source int
	data   string
}

func loadDataDumpEntries(source int, dataChan chan<- DataEntry, logger *logrus.Logger) {
	log := logger.WithFields(logrus.Fields{
		"source": source,
	})
	for {
		ms_sleep := time.Duration(200 + rand.Intn(500))
		time.Sleep(ms_sleep * time.Millisecond)
		log.Debugf("successfully loaded entry")
		datum := RandStringRunes(12)
		dataChan <- DataEntry{
			source: source,
			data:   datum,
		}
	}
}

func processDataEntry(datum DataEntry) {
	ctx, span := procTracer.Start(context.TODO(), "processing")
	defer span.End()

	log := logrus.WithFields(logrus.Fields{
		"source": datum.source,
	})

	decodeDatum(ctx, datum, log)
	calculateDatum(ctx, datum, log)
	commitDatum(ctx, datum, log)
}

func decodeDatum(ctx context.Context, datum DataEntry, log *logrus.Entry) {
	_, span := procTracer.Start(ctx, "decoding")
	defer span.End()
	log.Debugf("decoding datum %s\n", datum.data)
	time.Sleep(10 * time.Millisecond)
}

func calculateDatum(ctx context.Context, datum DataEntry, log *logrus.Entry) {
	_, span := procTracer.Start(ctx, "calculating")
	defer span.End()
	log.Debugf("decoding datum %s\n", datum.data)
	time.Sleep(10 * time.Millisecond)
}

func commitDatum(ctx context.Context, datum DataEntry, log *logrus.Entry) {
	_, span := procTracer.Start(ctx, "comitting")
	defer span.End()
	log.Debugf("comitting results for %s to memory\n", datum.data)
	time.Sleep(15 * time.Millisecond)
}

func processDataEntries(dataChan <-chan DataEntry) {
	for datum := range dataChan {
		processDataEntry(datum)
	}
}

const numOfProducers = 5
const numOfConsumers = 5

func startProcessing(numOfProducers, numOfConsumers int) {
	dataChan := make(chan DataEntry)

	loaderLogger := newCustomLogger("loader")

	for i := 0; i < numOfProducers; i++ {
		go loadDataDumpEntries(i, dataChan, loaderLogger)
	}

	procTracer = otel.Tracer("processor")

	for i := 0; i < numOfConsumers; i++ {
		go processDataEntries(dataChan)
	}

}

func setupOTelSDK() {
	traceExporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpointURL("http://jaeger-collector:4318"))
	if err != nil {
		panic(err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			sdktrace.WithBatchTimeout(time.Second)),
	)

	otel.SetTracerProvider(tracerProvider)

}

func main() {
	initLogging()
	setupOTelSDK()

	startProcessing(numOfProducers, numOfConsumers)

	// Block main goroutine forever.
	<-make(chan struct{})
}
