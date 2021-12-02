package radius

import (
	"log"

	"github.com/p-l/fringe/internal/db"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// ServeRadius starts a non-blocking Radius Server.
func ServeRadius(repo *db.Repository, secret string) {
	handler := func(writer radius.ResponseWriter, request *radius.Request) {
		username := rfc2865.UserName_GetString(request.Packet)
		password := rfc2865.UserPassword_GetString(request.Packet)
		code := radius.CodeAccessReject

		log.Printf("Radius request for %s from %v", username, request.RemoteAddr)

		if len(password) == 0 {
			log.Printf("WARN: No password provided in radius request from: %v", request.RemoteAddr)
		}

		authenticated, err := repo.AuthenticateUser(username, password)
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

	server := radius.PacketServer{
		Handler:      radius.HandlerFunc(handler),
		SecretSource: radius.StaticSecretSource([]byte(secret)),
	}

	log.Printf("Starting radius server on :1812")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
