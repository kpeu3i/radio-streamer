package main

import (
	"time"

	"github.com/joeshaw/envdecode"
)

type Config struct {
	HTTPServer struct {
		Address string `env:"HTTP_SERVER_ADDRESS,default=:7070"`
	}

	MQTTServer struct {
		Address  string `env:"MQTT_SERVER_ADDRESS,default=localhost:1883"`
		User     string `env:"MQTT_SERVER_USER,default=admin"`
		Password string `env:"MQTT_SERVER_PASSWORD,default=admin"`
		Topic    string `env:"MQTT_SERVER_Topic,default=zigbee2mqtt/0x00124b000cc8d641/action"`
	}

	ErrorHandling struct {
		RecoveryDelay time.Duration `env:"HTTP_SERVER_ADDRESS,default=1s"`
	}
}

func newConfig() (*Config, error) {
	var config Config

	err := envdecode.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
