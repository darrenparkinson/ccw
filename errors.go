package ccw

// Err implements the error interface so we can have constant errors.
type Err string

func (e Err) Error() string {
	return string(e)
}

// Error Constants
const (
	ErrBadRequest    = Err("ccw: bad request")
	ErrUnauthorized  = Err("ccw: unauthorized request")
	ErrForbidden     = Err("ccw: forbidden")
	ErrInternalError = Err("ccw: internal error")
	ErrUnknown       = Err("ccw: unexpected error occurred")
	ErrNotFound      = Err("ccw: not found")
)
