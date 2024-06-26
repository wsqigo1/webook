package aliyun

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"go.uber.org/zap"
	"strings"
)

type Service struct {
	client   *dysmsapi.Client
	signName string
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.SignName = s.signName
	request.TemplateCode = tplId
	request.PhoneNumbers = strings.Join(numbers, ",")

	byteParams, err := json.Marshal(map[string]string{
		"code": args[0],
	})
	request.TemplateParam = string(byteParams)

	response, err := s.client.SendSms(request)
	zap.L().Debug("请求阿里云SendSMS接口",
		zap.Any("req", request),
		zap.Any("resp", response))
	// 处理异常
	if err != nil {
		return err
	}
	if response.Code != "OK" {
		return errors.New("发送短信失败")
	}
	return nil
}

func NewService(client *dysmsapi.Client, signName string) *Service {
	return &Service{
		client:   client,
		signName: signName,
	}
}
