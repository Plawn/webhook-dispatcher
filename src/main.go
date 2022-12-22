package main

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

const (
	httpPort       = 8000
	prometheusPort = 9000
)

var (
	exampleSchemaDef = "{\"type\":\"record\",\"name\":\"Example\",\"namespace\":\"test\"," +
		"\"fields\":[{\"name\":\"payload\",\"type\":\"string\"},{\"name\":\"addresses\",\"type\":\"array\",\"items\":{\"type\": \"string\"}  }]}"
)

type Config struct {
	isWorker    bool   `env:"IS_WORKER" envDefault:"true"`
	url         string `env:"PULSAR_URL"`
	channelName string `env:"CHANNEL_NAME" envDefault:"webhooks"`
}

func main() {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
	if cfg.isWorker {
		RunWorker(cfg)
	} else {
		RunGateway(cfg)
	}
}
