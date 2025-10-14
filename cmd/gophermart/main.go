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

	a.Log.Infow(
		"starting server",
		"addr", a.Options.RunAddress,
		"database_uri", a.Options.DatabaseURI,
		"accrual_system_address", a.Options.AccrualSystemAddress,
		"log_level", a.Options.LogLevel,
		"jwt_duration", a.Options.JWTDuration,
		"cookie_max_age", a.Options.CookieMaxAge,
		"bonus_client", fmt.Sprintf("%T", a.BonusClient),
	)
	handler := handlers.New(a)
	router := router.New(handler)
	router = httputil.AddMiddlewares(
		router,
		logger.NewRequestLogger(a.Log),
		compressor.GZIPMiddleware,
	)
	log.Fatal(http.ListenAndServe(a.Options.RunAddress, router))
}
