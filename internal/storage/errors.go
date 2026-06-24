package storage

import "errors"

// ErrNotFound indicates that a requested resource does not exist.
var ErrNotFound = errors.New("not found")

// ErrConflict indicates that a requested write would conflict with existing data.
var ErrConflict = errors.New("conflict")
