package httpd

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mrz1836/go-sanitize"
	"github.com/p-l/fringe/internal/httpd/middlewares"
	"golang.org/x/crypto/acme/autocert"
)

const httpTimeouts = time.Second * 5

type redirectHandler struct {
	domain string
}

func (h *redirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	targetURL := sanitize.URL(fmt.Sprintf("https://%s%s", h.domain, r.RequestURI))
	log.Printf("REDIRECT [src:%v] %s to %s", r.RemoteAddr, sanitize.URL(r.URL.String()), targetURL)
	http.Redirect(w, r, targetURL, http.StatusFound)
}

// NewRedirectServer Create and configure the HTTP server that redirect to https server.
func NewRedirectServer(listenAddress string, serverDomain string, certManager *autocert.Manager) *http.Server {
	requestLogger := middlewares.NewLogMiddleware()
	redirect := new(redirectHandler)
	redirect.domain = serverDomain

	var handler http.Handler
	if certManager != nil {
		handler = certManager.HTTPHandler(redirect)
	} else {
		handler = redirect
	}

	log.Printf("Created http to https redirect server on %s", listenAddress)
	log.Printf("http://%s/* => https://%s/*", serverDomain, serverDomain)
	httpd := http.Server{
		Handler:      requestLogger.LogRequests(handler),
		Addr:         listenAddress,
		WriteTimeout: httpTimeouts,
		ReadTimeout:  httpTimeouts,
		IdleTimeout:  httpTimeouts,
	}

	return &httpd
}
