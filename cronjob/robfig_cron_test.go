package main

import (
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCronExpr(t *testing.T) {
	expr := cron.New(cron.WithSeconds())

	// 这个任务的标识符
	// @every 是遍历语法
	id, err := expr.AddFunc("@every 1s", func() {
		t.Log("执行了")
	})
	assert.NoError(t, err)
	t.Log("任务", id)

	// 调度运行
	expr.Start()
	time.Sleep(10 * time.Second)
	ctx := expr.Stop() // 意思是，你不要调度新任务执行了，你正在执行的继续执行
	t.Log("发出来停止信号")
	// 等待正在运行中的任务运行结束
	<-ctx.Done()
	t.Log("彻底停下来了，没有任务在执行")
	// 这边，彻底停下来了
}

type JobFunc func()

func (j JobFunc) Run() {
	j()
}
