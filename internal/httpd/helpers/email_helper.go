package helpers

import (
	"fmt"
	"regexp"
	"strings"
)

func IsEmailValid(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

	return emailRegex.MatchString(email)
}

func IsEmailInDomain(email string, domain string) bool {
	return IsEmailValid(email) && strings.HasSuffix(email, fmt.Sprintf("@%s", domain))
}
