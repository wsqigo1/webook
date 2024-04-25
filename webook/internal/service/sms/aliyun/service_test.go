package aliyun

import (
	"context"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSender(t *testing.T) {
	secretId := os.Getenv("SMS_SECRET_ID")
	secretKey := os.Getenv("SMS_SECRET_KEY")

	c, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", secretId, secretKey)
	if !assert.NoError(t, err) {
		return
	}

	s := NewService(c, "", "")

	testCases := []struct {
		name    string
		tplId   string
		params  []string
		numbers []string
		wantErr error
	}{
		{
			name:   "发送验证码",
			tplId:  "187756",
			params: []string{"123456"},
			// 改成你的手机号
			numbers: []string{""},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := s.Send(context.Background(), tc.tplId, tc.params, tc.numbers...)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
