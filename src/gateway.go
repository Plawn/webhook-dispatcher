package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Payload struct {
	Payload   string   `json:"payload"`
	Addresses []string `json:"addresses"`
}

// HandleFunc type defines the func that is used as a callback in ToFunc
type HandleFunc func(low int, high int)

// partition calls provided func per fragment
func partition(totalLength int, partitionLength int, hf HandleFunc) {
	if partitionLength <= 0 || totalLength <= 0 {
		return
	}
	partitions := totalLength / partitionLength
	var i int
	for i = 0; i < partitions; i++ {
		hf(i*partitionLength, i*partitionLength+partitionLength)
	}
	if rest := totalLength % partitionLength; rest != 0 {
		hf(i*partitionLength, i*partitionLength+rest)
	}
}

func RunGateway(config Config) {
	// Create a Pulsar client
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: config.Url,
	})
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	// Start a separate goroutine for Prometheus metrics
	// In this case, Prometheus metrics can be accessed via http://localhost:2112/metrics
	go func() {
		log.Printf("Starting Prometheus metrics at http://localhost:%v/metrics\n", prometheusPort)
		http.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(":"+strconv.Itoa(prometheusPort), nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Create a producer
	jsonSchemaWithProperties := pulsar.NewJSONSchema(exampleSchemaDef, nil)
	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic:  config.ChannelName,
		Schema: jsonSchemaWithProperties,
	})
	if err != nil {
		log.Fatal(err)
	}

	defer producer.Close()

	ctx := context.Background()

	// Write your business logic here
	// In this case, you build a simple Web server. You can produce messages by requesting http://localhost:8082/produce
	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		var payload Payload
		err := json.NewDecoder(r.Body).Decode(&payload)
		// TODO: split for faster distribution
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		partition(len(payload.Addresses), batchSize, func(l int, h int) {
			var currentPayload = &Payload{
				Addresses: payload.Addresses[l:h],
				Payload:   payload.Payload,
			}
			msgId, err := producer.Send(ctx, &pulsar.ProducerMessage{
				Value: &currentPayload,
			})
			if err != nil {
				log.Fatal(err)
			} else {
				log.Printf("Published message: %v", msgId)
				fmt.Fprint(w, "OK")
			}
		})

	})

	err = http.ListenAndServe(":"+strconv.Itoa(httpPort), nil)
	if err != nil {
		log.Fatal(err)
	}
}
