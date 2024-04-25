package aliyun

import (
	"context"
	"errors"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"math/rand"
	"strings"
	"time"
)

type Service struct {
	client   *dysmsapi.Client
	appId    string
	signName string
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.SignName = s.signName
	request.TemplateCode = tplId
	request.TemplateCode = fmt.Sprintf(`{"code":%s}`, args[0])
	request.PhoneNumbers = strings.Join(numbers, ",")
	
	response, err := s.client.SendSms(request)
	if err != nil {
		return err
	}
	if !response.IsSuccess() {
		return errors.New("发送短信失败")
	}
	return nil
}

func (s *Service) GenerateSmsCode(width int) string {
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	rand.Seed(time.Now().UnixNano())

	var sb strings.Builder
	for i := 0; i < width; i++ {
		fmt.Fprintf(&sb, "%d", numeric[rand.Intn(r)])
	}

	return sb.String()
}

func NewService(client *dysmsapi.Client, appId string, signName string) *Service {
	return &Service{
		client:   client,
		appId:    appId,
		signName: signName,
	}
}
