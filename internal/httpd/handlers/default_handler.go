package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/mrz1836/go-sanitize"
	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/p-l/fringe/internal/repos"
)

type DefaultHandler struct {
	UserRepo   *repos.UserRepository
	PageHelper *helpers.PageHelper
}

type NotFoundTemplateData struct {
	Path string
}

func NewDefaultHandler(userRepo *repos.UserRepository, pageHelper *helpers.PageHelper) *DefaultHandler {
	return &DefaultHandler{
		UserRepo:   userRepo,
		PageHelper: pageHelper,
	}
}

func (u *DefaultHandler) Root(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	claims, success := helpers.AuthClaimsFromContext(httpRequest.Context())
	if !success {
		log.Printf("Home [%v]: Could not read context", httpRequest.RemoteAddr)
		http.Error(httpResponse, "could not retrieve user information", http.StatusInternalServerError)

		return
	}

	log.Printf("Home/Root [%v]: claims %v", httpRequest.RemoteAddr, claims.Email)

	user, err := u.UserRepo.UserWithEmail(claims.Email)
	if errors.Is(err, repos.ErrUserNotFound) {
		createURL := fmt.Sprintf("/user/%s/enroll", claims.Email)
		http.Redirect(httpResponse, httpRequest, createURL, http.StatusFound)

		return
	}

	if err != nil {
		log.Printf("Home/Root [%v]: %v", httpRequest.RemoteAddr, err)
		http.Error(httpResponse, "could not retrieve user information", http.StatusInternalServerError)

		return
	}

	redirectPath := fmt.Sprintf("/user/%s/", claims.Email)
	if claims.IsAdmin() {
		redirectPath = "/user/"
	}

	log.Printf("Home [%v]: Redirecting to user's (%s) page %s", httpRequest.RequestURI, user.Email, redirectPath)
	http.Redirect(httpResponse, httpRequest, redirectPath, http.StatusFound)
}

func (u *DefaultHandler) NotFound(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	path := httpRequest.URL.Path

	log.Printf("Home/404 [%v]: %s", httpRequest.RemoteAddr, sanitize.URL(path))

	err := u.PageHelper.RenderPage(httpResponse, "default/404.gohtml", NotFoundTemplateData{Path: path})
	if err != nil {
		log.Printf("Could not render 404 template: %v", err)
		http.Error(httpResponse, "404 page not found", http.StatusNotFound)
	}
}
