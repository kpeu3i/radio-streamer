package mqttapi

import (
	"log"

	"github.com/kpeu3i/radio-streamer/streaming"
)

func VolumeUpHandler(service *streaming.Service) Handler {
	return func() {
		err := service.UpVolume(volumeStep)
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
		}
	}
}
