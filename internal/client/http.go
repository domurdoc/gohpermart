package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"resty.dev/v3"

	"github.com/domurdoc/gophermart/internal/models"
)

type HTTPBonusClient struct {
	client *resty.Client
}

type HTTPBonusParams struct {
	URL              string
	PoolSize         int
	PoolTimeout      time.Duration
	MaxReries        int
	RetryWaitTime    time.Duration
	RetryMaxWaitTime time.Duration
}

func NewBonusClient(params *HTTPBonusParams) *HTTPBonusClient {
	url := params.URL
	if !strings.HasPrefix(url, "http://") {
		url = "http://" + url
	}
	client := resty.
		New().
		SetBaseURL(url).
		SetTransport(&http.Transport{
			MaxConnsPerHost: params.PoolSize,
			MaxIdleConns:    params.PoolSize,
			IdleConnTimeout: params.PoolTimeout,
		}).
		SetRetryStrategy(nil).
		SetRetryCount(params.MaxReries).
		SetRetryWaitTime(params.RetryWaitTime).
		SetRetryMaxWaitTime(params.RetryMaxWaitTime)
	return &HTTPBonusClient{client: client}
}

func (c *HTTPBonusClient) Close() error {
	return c.client.Close()
}

type bonusResponse struct {
	Order   string  `json:"order" validate:"required"`
	Status  string  `json:"status" validate:"required,oneof=REGISTERED INVALID PROCESSING PROCESSED"`
	Accrual float64 `json:"accrual"`
}

var statusMap = map[string]string{
	"REGISTERED": models.StatusNew,
	"INVALID":    models.StatusInvalid,
	"PROCESSING": models.StatusProcessing,
	"PROCESSED":  models.StatusProcessed,
}

func (c *HTTPBonusClient) GetForOrder(ctx context.Context, orderNumber string) (*models.Bonus, error) {
	var response bonusResponse

	r, err := c.client.R().
		SetPathParam("orderNumber", orderNumber).
		SetResult(&response).
		Get("/api/orders/{orderNumber}")
	if err != nil {
		return nil, err
	}
	if r.StatusCode() == http.StatusNoContent {
		return nil, ErrOrderNotRegistered
	}
	if r.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("status code %d", r.StatusCode())
	}
	if err := validator.New().Struct(response); err != nil {
		return nil, err
	}
	bonus := models.Bonus{
		OrderNumber: orderNumber,
		Status:      statusMap[response.Status],
		Accrual:     response.Accrual,
	}
	return &bonus, nil
}
