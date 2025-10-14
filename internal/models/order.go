package models

import (
	"time"
)

const (
	StatusNew        = "NEW"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusProcessed  = "PROCESSED"
)

type Order struct {
	Number     string
	Status     string
	Accrual    float64
	UploadedAt time.Time
	Version    int
}

type Bonus struct {
	OrderNumber string
	Status      string
	Accrual     float64
}
