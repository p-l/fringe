package middlewares

import (
	"log"
	"net/http"
)

type LogMiddleware struct{}

func NewLogMiddleware() *LogMiddleware {
	return &LogMiddleware{}
}

func (u *LogMiddleware) LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(httpResponse http.ResponseWriter, httpRequest *http.Request) {
		log.Printf("%s %s [src:%v]", httpRequest.Method, httpRequest.RequestURI, httpRequest.RemoteAddr)
		// Call the next handlers, which can be another middlewares in the chain, or the final handlers.
		next.ServeHTTP(httpResponse, httpRequest)
	})
}
