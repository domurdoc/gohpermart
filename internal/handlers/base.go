package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/domurdoc/gophermart/internal/app"
	"github.com/domurdoc/gophermart/internal/auth/strategy"
	"github.com/domurdoc/gophermart/internal/auth/transport"
	"github.com/domurdoc/gophermart/internal/httputil"
	"github.com/domurdoc/gophermart/internal/models"
)

type Handler struct {
	app *app.App
}

func New(a *app.App) *Handler {
	return &Handler{app: a}
}

func (h *Handler) authRequest(w http.ResponseWriter, r *http.Request) (*models.User, bool) {
	user, err := h.app.Services.Auth.AuthenticateToken(r.Context(), r)
	if err != nil {
		var errNoToken *transport.NoTokenError
		var errInvalidToken *strategy.InvalidTokenError
		if errors.As(err, &errNoToken) || errors.As(err, &errInvalidToken) {
			w.WriteHeader(http.StatusUnauthorized)
			return nil, false
		}
		writeInternalServerError(w, err)
		return nil, false
	}
	return user, true
}

func readPlainTextRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	if !httputil.HasContentType(r.Header, httputil.ContentTypeTextPlain) {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return "", false
	}
	content, err := io.ReadAll(r.Body)
	if err != nil {
		writeInternalServerError(w, err)
		return "", false
	}
	return string(content), true
}

func readJSONRequest(w http.ResponseWriter, r *http.Request, schema any) bool {
	if !httputil.HasContentType(r.Header, httputil.ContentTypeJSON) {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return false
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(schema); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	if err := validator.New().Struct(schema); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

func writeJSONResponse(w http.ResponseWriter, data any, status int) {
	httputil.SetContentType(w.Header(), httputil.ContentTypeJSON)
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	if err := enc.Encode(data); err != nil {
		writeInternalServerError(w, err)
		return
	}
}

func writeInternalServerError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
