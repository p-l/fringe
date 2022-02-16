package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/p-l/fringe/internal/system"
)

type ConfigHandler struct {
	GoogleConfig system.GoogleConfig
}

type configResponse struct {
	GoogleClientID string `json:"google_client_id"`
}

func NewConfigHandler(googleConfig system.GoogleConfig) *ConfigHandler {
	return &ConfigHandler{
		GoogleConfig: googleConfig,
	}
}

func (h *ConfigHandler) Root(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	config := configResponse{GoogleClientID: h.GoogleConfig.ClientID}

	jsonResponse, err := json.Marshal(config)
	if err != nil {
		http.Error(httpResponse, err.Error(), http.StatusInternalServerError)

		return
	}

	httpResponse.Header().Set("Content-Type", "application/json")

	_, err = httpResponse.Write(jsonResponse)
	if err != nil {
		http.Error(httpResponse, err.Error(), http.StatusInternalServerError)

		return
	}
}
