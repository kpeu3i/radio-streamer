package httpapi

import (
	"net/http"

	"github.com/kpeu3i/radio-streamer/streaming"
)

func RadioPowerHandler(service *streaming.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if service.IsRadioPlaying() {
			service.StopRadio()
		} else {
			err := service.PlayRadio()
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)

				return
			}
		}
	}
}
