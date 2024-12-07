package xgo

import (
	"context"
	"log/slog"
	"runtime"
)

type AsyncContext struct {
	context.Context
}

func (AsyncContext) Done() <-chan struct{} {
	return nil
}

// Go encapsulation of goroutines to ensure timely recovery of panic.
// The execution of f is not controlled by the timeout of the original ctx, please customize in f if you want timeout ctx
// Example:
//
//	Go(ctx, func(ctx context.Context) {
//			funcCall(context.WithTimeout(ctx, 10*time.Second), args...)
//	})
func Go(ctx context.Context, f func(ctx context.Context)) {
	ctx = AsyncContext{ctx}
	go func(ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				slog.Error("program panic, err: %#v, detail: %s", err, buf[:n])
			}
		}()

		f(ctx)
	}(ctx)
}
