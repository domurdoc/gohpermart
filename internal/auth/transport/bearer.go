package transport

import (
	"fmt"
	"net/http"
	"strings"
)

type BearerTransport struct {
	header string
}

func NewBearer(header string) *BearerTransport {
	return &BearerTransport{header: header}
}

func (b *BearerTransport) Read(r *http.Request) (string, error) {
	authHeader := r.Header.Get(b.header)
	if authHeader == "" {
		return "", &NoTokenError{fmt.Errorf("no %s header set", b.header)}
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return "", &NoTokenError{fmt.Errorf("invalid token: undefined value")}
	}
	if parts[0] != "Bearer" {
		return "", &NoTokenError{fmt.Errorf("invalid token: wrong token type")}
	}
	return parts[1], nil
}

func (b *BearerTransport) Write(w http.ResponseWriter, tokenString string) error {
	w.Header().Set(b.header, fmt.Sprintf("Bearer %s", tokenString))
	return nil
}
