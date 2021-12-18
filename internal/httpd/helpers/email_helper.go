package helpers

import (
	"fmt"
	"net/mail"
	"strings"
)

func IsEmailValid(email string) bool {
	_, err := mail.ParseAddress(email)

	return err == nil
}

func IsEmailInDomain(email string, domain string) bool {
	return IsEmailValid(email) && strings.HasSuffix(email, fmt.Sprintf("@%s", domain))
}
