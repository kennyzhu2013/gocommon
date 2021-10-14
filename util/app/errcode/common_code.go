package errcode

var (
	Success = NewError(0, "success")
	ErrInvalidParams = NewError(1, "invalid params")
	ErrServerUnavailable = NewError(2, "server unavailable")
	ErrServerInternal = NewError(3, "server internal error")
)
