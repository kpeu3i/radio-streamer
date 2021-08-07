package httpapi

import (
	"fmt"
	"net/http"
)

func RecoverMiddleware(panicHandler func(v interface{})) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			defer func() {
				if r := recover(); r != nil {
					http.Error(writer, fmt.Sprintf("%v", r), http.StatusInternalServerError)
					panicHandler(r)
				}
			}()

			next(writer, request)
		}
	}
}
