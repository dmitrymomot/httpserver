package httpserver

import "errors"

var (
	ErrEmptyAddress     = errors.New("server address cannot be empty")
	ErrNilHandler       = errors.New("server handler cannot be nil")
	ErrServerStart      = errors.New("server failed to start")
	ErrServerStop       = errors.New("server failed to stop")
	ErrServerForceClose = errors.New("server force close failed")
)
