package mqttapi

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Handler func()

type Listener struct {
	topic    string
	client   mqtt.Client
	handlers map[string]Handler
	quit     chan struct{}
}

func NewListener(
	address string,
	username string,
	password string,
	topic string,
) *Listener {
	opts := mqtt.NewClientOptions()
	opts.SetClientID("radio-streamer")
	opts.AddBroker(fmt.Sprintf("tcp://%s", address))
	opts.SetUsername(username)
	opts.SetPassword(password)

	client := mqtt.NewClient(opts)

	return &Listener{
		topic:    topic,
		client:   client,
		handlers: make(map[string]Handler),
		quit:     make(chan struct{}),
	}
}

func (l *Listener) Register(action string, handler Handler) *Listener {
	l.handlers[action] = handler

	return l
}

func (l *Listener) Close() error {
	if !l.client.IsConnected() {
		return nil
	}

	l.quit <- struct{}{}

	l.client.Disconnect(100)

	return nil
}

func (l *Listener) Listen() error {
	if token := l.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	l.client.Subscribe(l.topic, 0, func(client mqtt.Client, message mqtt.Message) {
		log.Printf("* [%s] %s\n", message.Topic(), string(message.Payload()))

		action := string(message.Payload())
		if h, ok := l.handlers[action]; ok {
			h()
		}
	})

	<-l.quit

	return nil
}
