package repository

import (
	"context"
	"github.com/wsqigo/basic-go/webook/internal/domain"
)

//go:generate mockgen -source=./async_sms_repository.go -package=repomocks -destination=mocks/async_sms_repository.mock.go AsyncSmsRepository
type AsyncSmsRepository interface {
	// Add 添加一个异步 SMS 记录。
	// 你叫做 Create 或者 Insert 也可以
	Add(ctx context.Context, s domain.AsyncSms) error
}
