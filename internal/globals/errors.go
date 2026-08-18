package globals

import "errors"

var (
	ErrBadRequest = errors.New("bad request")
	ErrNotFound   = errors.New("resource not found")
	ErrConflict   = errors.New("resource conflict")
	ErrInternal   = errors.New("internal error")
	ErrUpstream   = errors.New("upstream service error")
)

const (
	CodeBadRequest = "BAD_REQUEST"
	CodeNotFound   = "NOT_FOUND"
	CodeConflict   = "CONFLICT"
	CodeInternal   = "INTERNAL_ERROR"
	CodeUpstream   = "UPSTREAM_ERROR"
)
