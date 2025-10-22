package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/domurdoc/gophermart/internal/app"
	"github.com/domurdoc/gophermart/internal/compressor"
	"github.com/domurdoc/gophermart/internal/handlers"
	"github.com/domurdoc/gophermart/internal/httputil"
	"github.com/domurdoc/gophermart/internal/logger"
	"github.com/domurdoc/gophermart/internal/router"
)

func main() {
	a, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	a.Services.Log.Infow(
		"starting server",
		"addr", a.Options.Server.RunAddress,
		"database_uri", a.Options.Repositories.DatabaseURI,
		"accrual_system_address", a.Options.Services.AccrualSystemAddress,
		"log_level", a.Options.Services.LogLevel,
		"jwt_duration", a.Options.Services.JWTDuration,
		"cookie_max_age", a.Options.Services.CookieMaxAge,
		"bonus_client", fmt.Sprintf("%T", a.Services.Client),
	)
	handler := handlers.New(a)
	router := router.New(handler)
	router = httputil.AddMiddlewares(
		router,
		logger.NewRequestLogger(a.Services.Log),
		compressor.GZIPMiddleware,
	)
	log.Fatal(http.ListenAndServe(a.Options.Server.RunAddress, router))
}
