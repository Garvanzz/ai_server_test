package serverdb

import "errors"

var (
	ErrNoEngine     = errors.New("serverdb: no engine")
	ErrWrongServer  = errors.New("serverdb: server id not match this process")
	ErrNotStarted   = errors.New("serverdb: manager not started")
)
