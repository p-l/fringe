package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/golang-jwt/jwt"
	"github.com/p-l/fringe/internal/repositories"
)

type googleAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	IDToken     string `json:"id_token"`
}

type googleUserInfoResponse struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

type jwtClaims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

const (
	generatedPasswordLen         = 24
	jwtValidityDurationInMinutes = 5
)

// ServeHTTP Starts blocking HTTP server.
func ServeHTTP(repo *repositories.UserRepository, rootURL string, googleClientID string, googleClientSecret string, allowedDomain string, jwtSecret string) {
	googleRedirectURI := fmt.Sprintf("%s/auth/google/callback", rootURL)

	http.HandleFunc("/", func(httpWriter http.ResponseWriter, httpRequest *http.Request) {
		log.Printf("GET / from %v", httpRequest.RemoteAddr)
		http.Redirect(httpWriter, httpRequest, "/auth", http.StatusFound)
	})

	http.HandleFunc("/user/", userHandler(repo, jwtSecret))
	http.HandleFunc("/auth/", authHandler(googleClientID, googleRedirectURI))
	http.HandleFunc("/auth/google/callback", googleCallbackHandler(googleClientID, googleClientSecret, googleRedirectURI, allowedDomain, jwtSecret))

	log.Printf("Starting http server on :9990")

	err := http.ListenAndServe(":9990", nil)
	if err != nil {
		log.Fatalf("FATAL: %v", err)
	}
}

func fetchGoogleTokenFromCallbackCode(code string, googleClientID string, googleClientSecret string, googleRedirectURI string) (auth googleAuthResponse, err error) {
	postParams := url.Values{}
	postParams.Add("code", code)
	postParams.Add("client_id", googleClientID)
	postParams.Add("client_secret", googleClientSecret)
	postParams.Add("redirect_uri", googleRedirectURI)
	postParams.Add("grant_type", "authorization_code")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", postParams)
	if err != nil {
		return googleAuthResponse{}, fmt.Errorf("could not create request to google token api: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return googleAuthResponse{}, fmt.Errorf("could not parse Google's response body: %w", err)
	}

	var googleAuth googleAuthResponse
	if err = json.Unmarshal(body, &googleAuth); err != nil {
		return googleAuthResponse{}, fmt.Errorf("could not unmarshal JSON from Google's response: %w", err)
	}

	return googleAuth, nil
}

func fetchGoogleUserInfoWithToken(tokenType string, token string) (userInfo googleUserInfoResponse, err error) {
	var googleUserInfo googleUserInfoResponse

	req, err := http.NewRequestWithContext(context.Background(), "GET", "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return googleUserInfoResponse{}, fmt.Errorf("could not request user info: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("%s %s", tokenType, token))

	userInfoResponse, err := http.DefaultClient.Do(req)
	if err != nil {
		return googleUserInfoResponse{}, fmt.Errorf("could not read response from user info: %w", err)
	}

	defer func() { _ = userInfoResponse.Body.Close() }()

	userInfoBody, err := ioutil.ReadAll(userInfoResponse.Body)
	if err != nil {
		return googleUserInfoResponse{}, fmt.Errorf("could not read user_info body: %w", err)
	}

	err = json.Unmarshal(userInfoBody, &googleUserInfo)
	if err != nil {
		return googleUserInfoResponse{}, fmt.Errorf("cannot unmarshal JSON from Google's response %w", err)
	}

	return googleUserInfo, nil
}

func googleCallbackHandler(googleClientID string, googleClientSecret string, googleRedirectURI string, allowedDomain string, jwtSecret string) func(w http.ResponseWriter, r *http.Request) {
	return func(httpWriter http.ResponseWriter, httpRequest *http.Request) {
		log.Printf("GET /auth/google/callback from %v", httpRequest.RemoteAddr)

		parsedQuery := httpRequest.URL.Query()
		code := parsedQuery.Get("code")

		googleAuth, err := fetchGoogleTokenFromCallbackCode(code, googleClientID, googleClientSecret, googleRedirectURI)
		if err != nil {
			log.Printf("ERR: %v", err)
			writeMessage(httpWriter, "Sorry, could not authenticate with Google")

			return
		}

		googleUserInfo, err := fetchGoogleUserInfoWithToken(googleAuth.TokenType, googleAuth.AccessToken)
		if err != nil {
			log.Printf("ERR: %v", err)
			writeMessage(httpWriter, "Sorry, could not fetch user's info from Google")

			return
		}

		if strings.Contains(googleUserInfo.Email, "@"+allowedDomain) {
			expirationTime := time.Now().Add(jwtValidityDurationInMinutes * time.Minute)
			// Create the JWT claims, which includes the username and expiry time
			claims := &jwtClaims{
				Email: googleUserInfo.Email,
				StandardClaims: jwt.StandardClaims{
					// In JWT, the expiry time is expressed as unix milliseconds
					ExpiresAt: expirationTime.Unix(),
				},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			jwtKey := []byte(jwtSecret)
			tokenString, _ := token.SignedString(jwtKey)

			http.SetCookie(httpWriter, &http.Cookie{
				Name:     "token",
				Value:    tokenString,
				Expires:  expirationTime,
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
			})
			http.Redirect(httpWriter, httpRequest, "/user/", http.StatusFound)
		}

		log.Printf("ERR: Rejecting %s from %v", googleUserInfo.Email, httpRequest.RemoteAddr)
		writeMessage(httpWriter, "Sorry, your email domain is not allowed here.")
	}
}

func writeMessage(w io.Writer, message string) {
	if _, err := fmt.Fprint(w, message); err != nil {
		log.Printf("ERR: could not send message: %v", err)
	}
}

func authHandler(googleClientID string, googleRedirectURI string) func(httpWriter http.ResponseWriter, httpRequest *http.Request) {
	return func(httpWriter http.ResponseWriter, httpRequest *http.Request) {
		log.Printf("GET /auth/ from %v", httpRequest.RemoteAddr)

		googleAuthScope := "https://www.googleapis.com/auth/userinfo.email"
		googleAuthURL := fmt.Sprintf(
			"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s",
			url.QueryEscape(googleClientID),
			url.QueryEscape(googleRedirectURI),
			url.QueryEscape(googleAuthScope))
		http.Redirect(httpWriter, httpRequest, googleAuthURL, http.StatusFound)
	}
}

func userHandler(repo *repositories.UserRepository, jwtSecret string) func(w http.ResponseWriter, r *http.Request) {
	return func(httpWriter http.ResponseWriter, httpRequest *http.Request) {
		log.Printf("GET /user/ from %v", httpRequest.RemoteAddr)
		// We can obtain the session token from the requests cookies, which come with every request
		tokenCookie, err := httpRequest.Cookie("token")
		if err != nil {
			log.Printf("ERR: Invalid token cookie received from %v: %v", httpRequest.RemoteAddr, err)
			http.Redirect(httpWriter, httpRequest, "/auth", http.StatusFound)

			return
		}

		// Get the JWT from the cookie
		tknStr := tokenCookie.Value
		claims := &jwtClaims{}
		jwtKey := []byte(jwtSecret)

		tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			log.Printf("ERR: received unparseable JWT from %v: %v", httpRequest.RemoteAddr, err)
			httpWriter.WriteHeader(http.StatusBadRequest)
			writeMessage(httpWriter, "Invalid claims token")

			return
		}

		if !tkn.Valid {
			log.Printf("ERR: received invalid JWT from %v: %v", httpRequest.RemoteAddr, err)
			httpWriter.WriteHeader(http.StatusBadRequest)
			writeMessage(httpWriter, "Invalid claims token")

			return
		}

		password := uniuri.NewLen(generatedPasswordLen)

		user, err := repo.UserWithEmail(claims.Email)
		if !errors.Is(err, repositories.ErrUserNotFound) && err != nil {
			log.Fatalf("FATAL: fail to query for user: %s", err.Error())
		}

		if user != nil {
			_, err = repo.UpdateUserPassword(claims.Email, password)
			if err != nil {
				log.Fatalf("FATAL: Fail to update user %s: %v", claims.Email, err)
			}
			_, _ = httpWriter.Write([]byte(fmt.Sprintf("Welcome back %s! Your new password is: %s", claims.Email, password)))
		} else {
			_, err = repo.CreateUser(claims.Email, password)
			if err != nil {
				log.Fatalf("FATAL: Fail to create user %s: %v", claims.Email, err)
			}
			_, _ = httpWriter.Write([]byte(fmt.Sprintf("Welcome %s! Your password is: %s", claims.Email, password)))
		}
	}
}
