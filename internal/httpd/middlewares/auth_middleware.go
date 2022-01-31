package middlewares

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/mrz1836/go-sanitize"
	"github.com/p-l/fringe/internal/httpd/helpers"
)

type AuthMiddleware struct {
	authHelper     *helpers.AuthHelper
	AuthPath       string
	ExcludedPaths  []string
	ProtectedPaths []string
}

func NewAuthMiddleware(redirectToAuthPath string, protectedPaths []string, excludedPaths []string, authHelper *helpers.AuthHelper) *AuthMiddleware {
	return &AuthMiddleware{
		authHelper:     authHelper,
		AuthPath:       redirectToAuthPath,
		ExcludedPaths:  excludedPaths,
		ProtectedPaths: protectedPaths,
	}
}

func (a *AuthMiddleware) IsProtected(path string) bool {
	for _, protected := range a.ProtectedPaths {
		if strings.HasPrefix(path, protected) {
			// Is it an exception?
			for _, excluded := range a.ExcludedPaths {
				if strings.HasPrefix(path, excluded) {
					return false
				}
			}

			// Not an exception
			return true
		}
	}

	// Is not under the protected list
	return false
}

func (a *AuthMiddleware) EnsureAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(httpResponse http.ResponseWriter, httpRequest *http.Request) {
		uri, _ := url.Parse(httpRequest.RequestURI)

		protected := a.IsProtected(uri.Path)
		if !protected {
			log.Printf("Auth [src:%v] %s is not protected, skipping auth", httpRequest.RemoteAddr, sanitize.URL(uri.Path))
			next.ServeHTTP(httpResponse, httpRequest)

			return
		}

		authorization := httpRequest.Header.Get("Authorization")

		token, err := extractBearerTokenFromAuthorization(authorization)
		if err != nil {
			log.Printf("Auth [src:%v] %s requested without valid Bearer token, redirecting to %s: %v", httpRequest.RemoteAddr, sanitize.URL(uri.Path), a.AuthPath, err)
			http.Error(httpResponse, "Invalid token", http.StatusForbidden)

			return
		}

		claims, err := a.authHelper.AuthClaimsFromSignedToken(token)
		if err != nil {
			log.Printf("Auth [src:%v] %v ", httpRequest.RemoteAddr, err)
			http.Error(httpResponse, "Invalid claims", http.StatusForbidden)

			return
		}

		// Success add claims to context
		ctx := claims.ContextWithClaims(httpRequest.Context())

		// Call the next handlers, which can be another middlewares in the chain, or the final handlers.
		next.ServeHTTP(httpResponse, httpRequest.WithContext(ctx))
	})
}

var errInvalidAuthorizationString = errors.New("authorization string is invalid")

func extractBearerTokenFromAuthorization(authorization string) (token string, err error) {
	if !strings.HasPrefix(authorization, "Bearer ") {
		return "", errInvalidAuthorizationString
	}

	splitAuthorization := strings.Split(authorization, "Bearer ")
	if len(splitAuthorization) <= 1 {
		return "", errInvalidAuthorizationString
	}

	return splitAuthorization[1], nil
}
