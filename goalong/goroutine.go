package goalong

import (
	"context"
	"log/slog"
	"runtime"
)

type GoFunc func()

// Go encapsulation of goroutines to ensure timely recovery of panic.
func Go(ctx context.Context, fn GoFunc) {
	go func() {
		fn.goWithRecover(ctx)
	}()
}

func (fn GoFunc) goWithRecover(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			slog.ErrorContext(ctx, "program panic, err: %#v, detail: %s", err, buf[:n])
		}
	}()

	if fn == nil {
		slog.ErrorContext(ctx, "function can not be nil")
		return
	}

	fn()
}
