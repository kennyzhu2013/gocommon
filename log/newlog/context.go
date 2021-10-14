package log

import (
	"context"
)

type logKey struct{}

func FromContext(ctx context.Context) (Log, bool) {
	c, ok := ctx.Value(logKey{}).(Log)
	return c, ok
}

func NewContext(ctx context.Context, c Log) context.Context {
	return context.WithValue(ctx, logKey{}, c)
}
