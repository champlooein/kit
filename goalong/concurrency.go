package goalong

import (
	"context"
	"sync"
)

const defaultLimit = 10

func GoWaits(ctx context.Context, fns ...GoFunc) {
	GoWaitsWithLimit(ctx, defaultLimit, fns...)
}

func GoWaitsWithLimit(ctx context.Context, limit int, fns ...GoFunc) {
	wg, ch := sync.WaitGroup{}, make(chan struct{}, limit)
	wg.Add(len(fns))
	defer close(ch)

	for i := 0; i < limit; i++ {
		ch <- struct{}{}
	}

	for _, fn := range fns {
		<-ch

		f := fn // 创建临时变量，避免闭包问题
		Go(ctx, func() {
			defer func() {
				wg.Done()
				ch <- struct{}{}
			}()

			f()
		})
	}
	wg.Wait()
}
