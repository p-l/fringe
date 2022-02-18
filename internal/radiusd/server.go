package radiusd

import (
	"log"

	"github.com/mrz1836/go-sanitize"
	"github.com/p-l/fringe/internal/repos"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// NewRadiusServer Creates and configure the Radius Server.
func NewRadiusServer(repo *repos.UserRepository, secret string, listenAddress string) *radius.PacketServer {
	handler := func(writer radius.ResponseWriter, request *radius.Request) {
		username := sanitize.Email(rfc2865.UserName_GetString(request.Packet), false)
		password := sanitize.SingleLine(rfc2865.UserPassword_GetString(request.Packet))
		code := radius.CodeAccessReject

		log.Printf("Radius request for %s from %v", username, request.RemoteAddr)

		if len(password) == 0 {
			log.Printf("WARN: No password provided in radiusd request from: %v", request.RemoteAddr)
		}

		authenticated, err := repo.Authenticate(username, password)
		if err != nil {
			log.Printf("ERR: Could not authenticate request from %v: %v", request.RemoteAddr, err)
		}

		if authenticated {
			code = radius.CodeAccessAccept
		}

		log.Printf("Response %v to request from %v", code, request.RemoteAddr)

		err = writer.Write(request.Response(code))
		if err != nil {
			log.Printf("ERR: Could not send responde to %v: %v", request.RemoteAddr, err)
		}
	}

	log.Printf("Created radius server on %s", listenAddress)
	server := radius.PacketServer{
		Addr:         listenAddress,
		Handler:      radius.HandlerFunc(handler),
		SecretSource: radius.StaticSecretSource([]byte(secret)),
	}

	return &server
}
