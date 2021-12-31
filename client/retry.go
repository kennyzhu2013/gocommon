package client

import (
	"context"
)

// note that returning either false or a non-nil error will result in the call not being retried
type RetryFunc func(ctx context.Context, req Request, statusCode int) (bool, error)

// RetryAlways always retry on error
func RetryAlways(ctx context.Context, req Request, retryCount int, err error) (bool, error) {
	return true, nil
}

// RetryOnError retries a request on a 500 or timeout error
func RetryOnError(ctx context.Context, req Request, statusCode int) (bool, error) {
	if statusCode <= 300 {
		return false, nil
	}

	switch statusCode {
	// retry on timeout or internal server error
	case 408, 500:
		return true, nil
	default:
		return false, nil
	}
}
