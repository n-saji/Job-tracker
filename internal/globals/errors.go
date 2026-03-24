package globals

import "errors"

var (
	ErrBadRequest = errors.New("bad request")
	ErrNotFound   = errors.New("resource not found")
	ErrConflict   = errors.New("resource conflict")
)

const (
	CodeBadRequest = "BAD_REQUEST"
	CodeNotFound   = "NOT_FOUND"
	CodeConflict   = "CONFLICT"
	CodeInternal   = "INTERNAL_ERROR"
)
