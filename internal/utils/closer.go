package utils

import (
	"errors"
	"slices"
)

type Closer struct {
	closes []func() error
}

func NewCloser() *Closer {
	return &Closer{}
}

func (c *Closer) Register(close func() error) {
	c.closes = append(c.closes, close)
}

func (c *Closer) Close() error {
	errs := make([]error, 0, len(c.closes))
	slices.Reverse(c.closes)
	for _, c := range c.closes {
		errs = append(errs, c())
	}
	c.closes = nil
	return errors.Join(errs...)
}
