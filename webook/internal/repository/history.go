package repository

import (
	"context"
	"github.com/wsqigo/basic-go/webook/internal/domain"
)

type HistoryRecordRepository interface {
	AddRecord(ctx context.Context, record domain.HistoryRecord) error
}
