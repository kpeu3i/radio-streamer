package httpapi

import (
	"net/http"
)

type Middleware func(next http.HandlerFunc) http.HandlerFunc

func WrapHandler(handler http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	return handler
}
