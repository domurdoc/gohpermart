package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/domurdoc/gophermart/internal/models"
)

type balanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func (h *Handler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authRequest(w, r)
	if !ok {
		return
	}
	balance, err := h.app.Services.Balance.GetUserBalance(r.Context(), user)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	writeJSONResponse(w, balanceResponse(*balance), http.StatusOK)
}

type withdrawRequest struct {
	OrderNumber string  `json:"order" validate:"required"`
	Sum         float64 `json:"sum" validate:"required"`
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	var req withdrawRequest

	user, ok := h.authRequest(w, r)
	if !ok {
		return
	}
	if !readJSONRequest(w, r, &req) {
		return
	}
	_, err := h.app.Services.Balance.Withdraw(r.Context(), user, req.OrderNumber, req.Sum)
	if err != nil {
		if errors.Is(err, models.ErrInvalidOrderNumber) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, models.ErrNotEnoughBalance) {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		writeInternalServerError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type responseWithdrawal struct {
	OrderNumber string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (h *Handler) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authRequest(w, r)
	if !ok {
		return
	}
	withdrawals, err := h.app.Services.Balance.GetUserWithdrawals(r.Context(), user)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	res := make([]responseWithdrawal, len(withdrawals))
	for i, withdrawal := range withdrawals {
		res[i] = responseWithdrawal(*withdrawal)
	}
	writeJSONResponse(w, res, http.StatusOK)
}
