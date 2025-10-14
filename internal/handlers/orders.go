package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/domurdoc/gophermart/internal/models"
)

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authRequest(w, r)
	if !ok {
		return
	}
	orderNumber, ok := readPlainTextRequest(w, r)
	if !ok {
		return
	}
	_, err := h.app.OrderService.CreateOrder(r.Context(), user, orderNumber)
	if err != nil {
		if errors.Is(err, models.ErrInvalidOrderNumber) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, models.ErrOrderAlreadyCreatedByUser) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, models.ErrOrderCreatedByAnotherUser) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		writeInternalServerError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

type responseOrder struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func (h *Handler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authRequest(w, r)
	if !ok {
		return
	}
	orders, err := h.app.OrderService.GetUserOrders(r.Context(), user)
	if err != nil {
		writeInternalServerError(w, err)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	res := make([]responseOrder, len(orders))
	for i, order := range orders {
		res[i] = responseOrder{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		}
	}
	writeJSONResponse(w, res, http.StatusOK)
}
