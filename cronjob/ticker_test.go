package main

import (
	"context"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	// 间隔一秒钟的 ticker
	ticker := time.NewTicker(time.Second)
	// 这一句不要忘了
	// 避免潜在的 goroutine 泄露的问题
	defer ticker.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 每隔一秒钟就会有一个信号
	//for range ticker.C {
	//}
	for {
		select {
		case <-ctx.Done():
			// 循环结束
			t.Log("循环结束")
			// break 不会退出循环
			goto end
		case now := <-ticker.C:
			t.Log("过了一秒", now.Unix())
		}
	}
end:
	t.Log("goto 过来了，结束程序")
}
