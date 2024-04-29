package tencent

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"os"
	"testing"
)

func TestSender(t *testing.T) {
	secretId, ok := os.LookupEnv("TX_ACCESS_KEY_ID")
	if !ok {
		t.Fatal()
	}
	secretKey, ok := os.LookupEnv("TX_ACCESS_KEY_SECRET")
	if !ok {
		t.Fatal()
	}

	client, err := sms.NewClient(common.NewCredential(secretId, secretKey),
		"ap-guangzhou",
		profile.NewClientProfile())
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(client, "1400088380", "量链科技")

	testCases := []struct {
		name    string
		tplId   string
		params  []string
		numbers []string
		wantErr error
	}{
		{
			name:   "发送验证码",
			tplId:  "1574317",
			params: []string{"stage", "1", "1", "1"},
			// 改成你的手机号码
			numbers: []string{"19124155294"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			er := s.Send(context.Background(), tc.tplId, tc.params, tc.numbers...)
			assert.Equal(t, tc.wantErr, er)
		})
	}
}
