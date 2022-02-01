package handlers

import (
	"log"
	"net/http"

	"github.com/mrz1836/go-sanitize"
)

type DefaultHandler struct{}

func NewDefaultHandler() *DefaultHandler {
	return &DefaultHandler{}
}

func (u *DefaultHandler) NotFound(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	path := httpRequest.URL.Path
	log.Printf("Default/404 [%v]: %s", httpRequest.RemoteAddr, sanitize.URL(path))

	http.Error(httpResponse, "404 page not found", http.StatusNotFound)
}
