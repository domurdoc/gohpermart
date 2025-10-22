package router

import (
	"net/http"

	"github.com/domurdoc/gophermart/internal/handlers"
)

func New(handler *handlers.Handler) http.Handler {
	router := http.NewServeMux()
	setupRoutes(router, handler)
	return router
}

func setupRoutes(router *http.ServeMux, handler *handlers.Handler) {
	router.HandleFunc("POST /api/user/register", handler.RegisterUser)
	router.HandleFunc("POST /api/user/login", handler.Login)
	router.HandleFunc("POST /api/user/orders", handler.CreateOrder)
	router.HandleFunc("GET /api/user/orders", handler.GetUserOrders)
	router.HandleFunc("GET /api/user/balance", handler.GetUserBalance)
	router.HandleFunc("POST /api/user/balance/withdraw", handler.Withdraw)
	router.HandleFunc("GET /api/user/withdrawals", handler.GetUserWithdrawals)
}
