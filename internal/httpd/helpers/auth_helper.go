package helpers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

type AuthHelper struct {
	secret string
}

func NewAuthHelper(secret string) *AuthHelper {
	return &AuthHelper{
		secret: secret,
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

var ErrInvalidClaimsToken = errors.New("invalid claims token")

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
