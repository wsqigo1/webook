package service

import (
	"context"
	"fmt"
	"github.com/wsqigo/basic-go/webook/internal/repository"
	"github.com/wsqigo/basic-go/webook/internal/service/sms"
	"math/rand"
)

var ErrCodeSendTooMany = repository.ErrCodeVerifyTooMany

type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type codeService struct {
	repo repository.CodeRepository
	sms  sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &codeService{
		repo: repo,
		sms:  smsSvc,
	}
}

func (svc *codeService) Send(ctx context.Context, biz, phone string) error {
	code := svc.generate()
	err := svc.repo.Set(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	const codeTplId = "SMS_186030080"
	return svc.sms.Send(ctx, codeTplId, []string{code}, phone)
}

func (svc *codeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	ok, err := svc.repo.Verify(ctx, biz, phone, inputCode)
	if err == repository.ErrCodeVerifyTooMany {
		// 相当于，我们对外面屏蔽了验证次数过多的错误，我们就是告诉调用者，你这个不对
		return false, nil
	}
	return ok, err
}

func (svc *codeService) generate() string {
	// 0-999999
	code := rand.Intn(1000000)
	return fmt.Sprintf("%6d", code)
}
