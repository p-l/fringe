package helpers

import (
	"errors"
	"log"
	"strings"

	"github.com/golang-jwt/jwt"
)

type AuthHelper struct {
	secret        string
	admins        []string
	AllowedDomain string
}

var ErrInvalidClaimsToken = errors.New("invalid claims token")

func NewAuthHelper(allowedDomain string, secret string, adminsEmail []string) *AuthHelper {
	return &AuthHelper{
		secret:        secret,
		admins:        adminsEmail,
		AllowedDomain: allowedDomain,
	}
}

func (h *AuthHelper) NewJWTSignedString(claims *AuthClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey := []byte(h.secret)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Fatalf("!!! Error creating JWT signed string: %v", err)
	}

	return tokenString
}

func (h *AuthHelper) AuthClaimsFromSignedToken(tokenString string) (*AuthClaims, error) {
	claims := &AuthClaims{}
	jwtKey := []byte(h.secret)

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) { return jwtKey, nil })
	if err != nil {
		return nil, ErrInvalidClaimsToken
	}

	// JWT content and expiry validation
	if !token.Valid {
		return nil, ErrInvalidClaimsToken
	}

	return claims, nil
}

func (h *AuthHelper) InAllowedDomain(email string) bool {
	return IsEmailInDomain(email, h.AllowedDomain)
}

func (h *AuthHelper) RoleForEmail(email string) string {
	for _, adminEmail := range h.admins {
		if strings.EqualFold(adminEmail, email) {
			return "admin"
		}
	}

	return "user"
}
