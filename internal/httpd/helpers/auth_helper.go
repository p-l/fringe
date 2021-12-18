package helpers

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

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

func (h *AuthHelper) NewJWTCookieFromClaims(claims *AuthClaims) *http.Cookie {
	expirationTime := time.Unix(claims.StandardClaims.ExpiresAt, 0)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey := []byte(h.secret)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Fatalf("!!! Error creating JWT signed string: %v", err)
	}

	return &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  expirationTime,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}
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

func (h *AuthHelper) RemoveJWTCookie() *http.Cookie {
	expirationTime := time.Unix(0, 0)

	return &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  expirationTime,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}
}

func (h *AuthHelper) InAllowedDomain(email string) bool {
	return IsEmailInDomain(email, h.AllowedDomain)
}

func (h *AuthHelper) PermissionsForEmail(email string) string {
	for _, adminEmail := range h.admins {
		if strings.EqualFold(adminEmail, email) {
			return "admin"
		}
	}

	return "user"
}
