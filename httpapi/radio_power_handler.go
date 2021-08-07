package httpapi

import (
	"net/http"

	"github.com/kpeu3i/radio-streamer/streaming"
)

func RadioPowerHandler(app *streaming.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if app.IsRadioPlaying() {
			app.StopRadio()
		} else {
			err := app.PlayRadio()
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)

				return
			}
		}
	}
}
