package mqttapi

import (
	"log"

	"github.com/kpeu3i/radio-streamer/streaming"
)

func RadioStreamPrevHandler(service *streaming.Service) Handler {
	return func() {
		err := service.PrevRadioStream()
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
		}
	}
}
