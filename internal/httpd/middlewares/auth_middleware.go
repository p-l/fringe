package middlewares

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/mrz1836/go-sanitize"
	"github.com/p-l/fringe/internal/httpd/helpers"
)

type AuthMiddleware struct {
	authHelper    *helpers.AuthHelper
	ExcludedPaths []string
	AuthPath      string
}

func NewAuthMiddleware(redirectToAuthPath string, excludedPaths []string, authHelper *helpers.AuthHelper) *AuthMiddleware {
	return &AuthMiddleware{
		authHelper:    authHelper,
		ExcludedPaths: append(excludedPaths, redirectToAuthPath),
		AuthPath:      redirectToAuthPath,
	}
}

func (a *AuthMiddleware) EnsureAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(httpResponse http.ResponseWriter, httpRequest *http.Request) {
		uri, _ := url.Parse(httpRequest.RequestURI)

		// Skip auth validation for excluded path
		for _, path := range a.ExcludedPaths {
			if strings.HasPrefix(uri.Path, path) {
				log.Printf("Auth [src:%v] %s matches excluded path %s, skipping auth", httpRequest.RemoteAddr, sanitize.URL(uri.Path), sanitize.URL(path))
				next.ServeHTTP(httpResponse, httpRequest)

				return
			}
		}

		tokenCookie, err := httpRequest.Cookie("token")
		if err != nil {
			http.Redirect(httpResponse, httpRequest, a.AuthPath, http.StatusFound)

			return
		}

		claims, err := a.authHelper.AuthClaimsFromSignedToken(tokenCookie.Value)
		if err != nil {
			log.Printf("Auth [src:%v] %v ", httpRequest.RemoteAddr, err)

			http.SetCookie(httpResponse, a.authHelper.RemoveJWTCookie())
			http.Redirect(httpResponse, httpRequest, a.AuthPath, http.StatusFound)

			return
		}

		// Success add claims to context
		claims.Refresh()
		ctx := claims.ContextWithClaims(httpRequest.Context())

		// Refresh token cookie
		http.SetCookie(httpResponse, a.authHelper.NewJWTCookieFromClaims(claims))

		// Call the next handlers, which can be another middlewares in the chain, or the final handlers.
		next.ServeHTTP(httpResponse, httpRequest.WithContext(ctx))
	})
}
