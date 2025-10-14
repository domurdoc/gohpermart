package models

import (
	"errors"
)

var (
	ErrNotEnoughBalance          = errors.New("not enough balance")
	ErrOrderAlreadyCreatedByUser = errors.New("order already created by user")
	ErrOrderCreatedByAnotherUser = errors.New("order created by another user")
	ErrUsernameExists            = errors.New("username already exists")
	ErrUserNotFound              = errors.New("user not found")
	ErrInvalidCredentials        = errors.New("credentials invalid")
	ErrInvalidOrderNumber        = errors.New("invalid order number")
)
