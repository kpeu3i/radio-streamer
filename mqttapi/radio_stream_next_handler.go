package mqttapi

import (
	"log"

	"github.com/kpeu3i/radio-streamer/streaming"
)

func RadioStreamNextHandler(service *streaming.Service) Handler {
	return func() {
		err := service.NextRadioStream()
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
		}
	}
}
