package store

import "errors"

var (
	ErrNotFound  = errors.New("Document not found")
	ErrTransient = errors.New("Transient Error")
)
