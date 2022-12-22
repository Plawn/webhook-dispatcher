package main

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

var (
	exampleSchemaDef = "[{\"name\":\"payload\",\"type\":\"string\"},{\"name\":\"addresses\",\"type\":\"array\",\"items\":{\"type\": \"string\"}  }]"
	httpPort         = 8000
	prometheusPort   = 9500
)

type Config struct {
	IsWorker       bool   `env:"IS_WORKER" envDefault:"true"`
	Url            string `env:"PULSAR_URL"`
	ChannelName    string `env:"CHANNEL_NAME" envDefault:"webhooks"`
	PrometheusPort int    `env:"PROMETHEUS_PORT" envDefault:"9500"`
	HttpPort       int    `env:"HTTP_PORT" envDefault:"8000"`
}

func main() {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("Missing context, %+v\n", err)
	}
	httpPort = cfg.HttpPort
	prometheusPort = cfg.PrometheusPort
	if cfg.IsWorker {
		RunWorker(cfg)
	} else {
		RunGateway(cfg)
	}
}
