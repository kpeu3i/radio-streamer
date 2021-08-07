package mqttapi

import "log"

func RecoverMiddleware(panicHandler func(v interface{})) Middleware {
	return func(next Handler) Handler {
		return func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[PANIC] %v\n", r)
					panicHandler(r)
				}
			}()

			next()
		}
	}
}
