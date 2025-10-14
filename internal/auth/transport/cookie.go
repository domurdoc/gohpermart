package transport

import "net/http"

type CookieTransport struct {
	name     string
	maxAge   int
	secure   bool
	httpOnly bool
	path     string
	sameSite http.SameSite
}

func NewCookie(name string, maxAge int, secure bool) *CookieTransport {
	return &CookieTransport{
		name:     name,
		maxAge:   maxAge,
		secure:   secure,
		httpOnly: true,
		path:     "/",
		sameSite: http.SameSiteStrictMode,
	}
}

func (c *CookieTransport) Read(r *http.Request) (string, error) {
	cookie, err := r.Cookie(c.name)
	if err != nil {
		return "", &NoTokenError{err}
	}
	if err := cookie.Valid(); err != nil {
		return "", &NoTokenError{err}
	}
	return cookie.Value, nil
}

func (c *CookieTransport) Write(w http.ResponseWriter, tokenString string) error {
	cookie := http.Cookie{
		Name:     c.name,
		Value:    tokenString,
		SameSite: c.sameSite,
		MaxAge:   c.maxAge,
		Secure:   c.secure,
		HttpOnly: c.httpOnly,
		Path:     c.path,
	}
	if err := cookie.Valid(); err != nil {
		return err
	}
	http.SetCookie(w, &cookie)
	return nil
}
