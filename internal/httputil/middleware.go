package httputil

import "net/http"

type Middleware func(http.Handler) http.Handler

func AddMiddlewares(root http.Handler, middlewares ...Middleware) http.Handler {
	handler := root
	for _, mw := range middlewares {
		handler = mw(handler)
	}
	return handler
}
