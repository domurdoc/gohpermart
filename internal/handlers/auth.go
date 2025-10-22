package handlers

import (
	"errors"
	"net/http"

	"github.com/domurdoc/gophermart/internal/models"
)

type userCredentialsRequest struct {
	Username string `json:"login" validate:"required,alphanum"`
	Password string `json:"password" validate:"required"`
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req userCredentialsRequest

	if !readJSONRequest(w, r, &req) {
		return
	}
	user, err := h.app.Services.Auth.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, models.ErrUsernameExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		writeInternalServerError(w, err)
		return
	}
	if err := h.app.Services.Auth.Login(r.Context(), w, user); err != nil {
		writeInternalServerError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req userCredentialsRequest

	if !readJSONRequest(w, r, &req) {
		return
	}
	user, err := h.app.Services.Auth.AuthenticateCredentials(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		writeInternalServerError(w, err)
		return
	}
	if err := h.app.Services.Auth.Login(r.Context(), w, user); err != nil {
		writeInternalServerError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
