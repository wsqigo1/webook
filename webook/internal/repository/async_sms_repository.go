package repository

import (
	"context"
	"github.com/wsqigo/basic-go/webook/internal/domain"
	"gorm.io/gorm"
)

var ErrWaitingSMSNotFound = gorm.ErrRecordNotFound

//go:generate mockgen -source=./async_sms_repository.go -package=repomocks -destination=mocks/async_sms_repository.mock.go AsyncSmsRepository
type AsyncSmsRepository interface {
	// Add 添加一个异步 SMS 记录。
	// 你叫做 Create 或者 Insert 也可以
	Add(ctx context.Context, s domain.AsyncSms) error
	PreemptWaitingSMS(ctx context.Context) (domain.AsyncSms, error)
	ReportScheduleResult(ctx context.Context, id int64, res bool) error
}
