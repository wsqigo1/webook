package aliyun

import (
	"context"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSender(t *testing.T) {
	secretId, _ := os.LookupEnv("ALI_ACCESS_KEY_ID")
	secretKey, _ := os.LookupEnv("ALI_ACCESS_KEY_SECRET")

	c, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", secretId, secretKey)
	if !assert.NoError(t, err) {
		return
	}

	s := NewService(c, "量链科技")

	testCases := []struct {
		name    string
		tplId   string
		params  []string
		numbers []string
		wantErr error
	}{
		{
			name:   "发送验证码",
			tplId:  "SMS_186030080",
			params: []string{"123456"},
			// 改成你的手机号
			numbers: []string{"19124155294"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := s.Send(context.Background(), tc.tplId, tc.params, tc.numbers...)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
