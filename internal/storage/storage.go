package storage

import "errors"

var (
	ErrClientExists   = errors.New("client already exists")
	ErrClientNotFound = errors.New("client not found")
	ErrAppExists      = errors.New("app already exists")
	ErrAppNotFound    = errors.New("app not found")
)
