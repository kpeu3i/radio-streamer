package httpapi

import (
	"net/http"

	"github.com/kpeu3i/radio-streamer/streaming"
)

func RadioStreamPrevHandler(app *streaming.Service) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		err := app.PrevRadioStream()
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)

			return
		}
	}
}