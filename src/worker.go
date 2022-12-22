package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RunWorker(config Config) {
	fmt.Printf("starting with %s\n", config.Url)
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
	channel := make(chan pulsar.ConsumerMessage, 100)
	// Create a consumer
	jsonSchemaWithProperties := pulsar.NewJSONSchema(exampleSchemaDef, nil)
	options := pulsar.ConsumerOptions{
		Topic:                       config.ChannelName,
		SubscriptionName:            "worker",
		Type:                        pulsar.Shared,
		Schema:                      jsonSchemaWithProperties,
		SubscriptionInitialPosition: pulsar.SubscriptionPositionEarliest,
	}
	options.MessageChannel = channel
	consumer, err := client.Subscribe(options)
	if err != nil {
		log.Fatal(err)
	}

	defer consumer.Close()

	// Receive messages from channel. The channel returns a struct which contains message and the consumer from where
	// the message was received. It's not necessary here since we have 1 single consumer, but the channel could be
	// shared across multiple consumers as well
	for cm := range channel {
		msg := cm.Message
		var s Payload
		err2 := msg.GetSchemaValue(&s)
		if err2 != nil {
			log.Fatal("error")
		}
		fmt.Printf("Received message  msgId: %v -- content: '%s'\n",
			msg.ID(), string(msg.Payload()))
		for _, a := range s.Addresses {
			request, requestErr := http.NewRequest("POST", a, bytes.NewBuffer([]byte(s.Payload)))
			if (requestErr) != nil {
				continue // means the url is bad
			}
			request.Header.Set("Content-Type", "application/json; charset=UTF-8")
			client := &http.Client{
				Timeout: 1 * time.Second,
			}
			resp, httpErr := client.Do(request)
			if httpErr != nil {
				log.Fatal(httpErr)
			} else {
				fmt.Printf("status: %s\n", resp.Status)
			}
		}
		consumer.Ack(msg)
	}
}
