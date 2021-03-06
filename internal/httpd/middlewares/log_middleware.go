package middlewares

import (
	"log"
	"net/http"
	"net/url"

	"github.com/mrz1836/go-sanitize"
)

type LogMiddleware struct {
	logger *log.Logger
}

func NewLogMiddleware(logger *log.Logger) *LogMiddleware {
	return &LogMiddleware{
		logger: logger,
	}
}

func (l *LogMiddleware) LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(httpResponse http.ResponseWriter, httpRequest *http.Request) {
		requestURL, err := url.Parse(httpRequest.URL.String())
		if err != nil {
			l.logger.Printf("%s <%v> [src:%v]", httpRequest.Method, err, httpRequest.RemoteAddr)
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

		l.logger.Printf("%s %v [src:%v]", httpRequest.Method, sanitize.URL(requestURL.String()), httpRequest.RemoteAddr)
		// Call the next handlers, which can be another middlewares in the chain, or the final handlers.
		next.ServeHTTP(httpResponse, httpRequest)
	})
}
