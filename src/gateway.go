package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RunGateway(config Config) {
	// Create a Pulsar client
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: config.url,
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
	properties := make(map[string]string)
	properties["pulsar"] = "hello"
	jsonSchemaWithProperties := pulsar.NewJSONSchema(exampleSchemaDef, properties)
	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic:  "topic-1",
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
		msgId, err := producer.Send(ctx, &pulsar.ProducerMessage{
			Payload: []byte(fmt.Sprintf("hello world")),
		})
		if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Published message: %v", msgId)
			fmt.Fprintf(w, "Published message: %v", msgId)
		}
	})

	err = http.ListenAndServe(":"+strconv.Itoa(httpPort), nil)
	if err != nil {
		log.Fatal(err)
	}
}
