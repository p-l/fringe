package radiusd

import (
	"log"

	"github.com/p-l/fringe/internal/repos"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// NewRadiusServer Creates and configure the Radius Server.
func NewRadiusServer(repo *repos.UserRepository, secret string) *radius.PacketServer {
	handler := func(writer radius.ResponseWriter, request *radius.Request) {
		username := rfc2865.UserName_GetString(request.Packet)
		password := rfc2865.UserPassword_GetString(request.Packet)
		code := radius.CodeAccessReject

		log.Printf("Radius request for %s from %v", username, request.RemoteAddr)

		if len(password) == 0 {
			log.Printf("WARN: No password provided in radiusd request from: %v", request.RemoteAddr)
		}

		authenticated, err := repo.AuthenticateUser(username, password)
		if err != nil {
			log.Printf("ERR: Could not authenticate request from %v: %v", request.RemoteAddr, err)
		}

		if authenticated {
			if err = repo.TouchUser(username); err != nil {
				log.Printf("ERR: Could not update user's last seen: %v", err)
			}
			code = radius.CodeAccessAccept
		}

		log.Printf("Response %v to request from %v", code, request.RemoteAddr)

		err = writer.Write(request.Response(code))
		if err != nil {
			log.Printf("ERR: Could not send responde to %v: %v", request.RemoteAddr, err)
		}
	}

	log.Printf("Created radius server on :1812")
	server := radius.PacketServer{
		Addr:         "127.0.0.1:1812",
		Handler:      radius.HandlerFunc(handler),
		SecretSource: radius.StaticSecretSource([]byte(secret)),
	}

	return &server
}
