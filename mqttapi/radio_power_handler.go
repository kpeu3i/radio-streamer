package mqttapi

import (
	"log"

	"github.com/kpeu3i/radio-streamer/streaming"
)

func RadioPowerHandler(service *streaming.Service) Handler {
	return func() {
		if service.IsRadioPlaying() {
			service.StopRadio()
		} else {
			err := service.PlayRadio()
			if err != nil {
				log.Printf("[ERROR] %v\n", err)
			}
		}
	}
}
