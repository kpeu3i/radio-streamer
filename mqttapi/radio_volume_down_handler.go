package mqttapi

import (
	"log"

	"github.com/kpeu3i/radio-streamer/streaming"
)

func VolumeDownHandler(service *streaming.Service) Handler {
	return func() {
		err := service.DownVolume(volumeStep)
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
		}
	}
}
