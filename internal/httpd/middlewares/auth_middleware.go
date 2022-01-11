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
	log.Printf("DEBUG: path=%s", path)

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

		tokenCookie, err := httpRequest.Cookie("token")
		if err != nil {
			log.Printf("Auth [src:%v] %s requested without token, redirecting to %s", httpRequest.RemoteAddr, sanitize.URL(uri.Path), a.AuthPath)
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
