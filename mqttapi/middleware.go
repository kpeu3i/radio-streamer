package mqttapi

type Middleware func(next Handler) Handler

func WrapHandler(handler Handler, middlewares ...Middleware) Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	return handler
}
