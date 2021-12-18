package middlewares

import (
	"log"
	"net/http"
	"net/url"
)

type LogMiddleware struct{}

func NewLogMiddleware() *LogMiddleware {
	return &LogMiddleware{}
}

func (u *LogMiddleware) LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(httpResponse http.ResponseWriter, httpRequest *http.Request) {
		requestURL, err := url.Parse(httpRequest.URL.String())
		if err != nil {
			log.Printf("%s <%v> [src:%v]", httpRequest.Method, err, httpRequest.RemoteAddr)
			next.ServeHTTP(httpResponse, httpRequest)

			return
		}

		redactedKeys := []string{"code", "password", "pass", "secret", "token"}
		urlQuery := requestURL.Query()
		for _, key := range redactedKeys {
			if len(urlQuery.Get(key)) > 1 {
				urlQuery.Set(key, "redacted")
			}
		}
		requestURL.RawQuery = urlQuery.Encode()

		log.Printf("%s %v [src:%v]", httpRequest.Method, requestURL, httpRequest.RemoteAddr)
		// Call the next handlers, which can be another middlewares in the chain, or the final handlers.
		next.ServeHTTP(httpResponse, httpRequest)
	})
}
