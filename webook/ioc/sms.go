package ioc

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentSMS "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"github.com/wsqigo/basic-go/webook/internal/service/sms"
	"github.com/wsqigo/basic-go/webook/internal/service/sms/aliyun"
	"github.com/wsqigo/basic-go/webook/internal/service/sms/localsms"
	"github.com/wsqigo/basic-go/webook/internal/service/sms/tencent"
	"os"
)

func InitSMSService() sms.Service {
	return localsms.NewService()
	// 如果有需要，就可以用这个
	//return initTencentSMSService()
	//return initAliyunSMSService()
}

func initTencentSMSService() sms.Service {
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		panic("找不到腾讯 SMS 的 secret id")
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		panic("找不到腾讯 SMS 的 secret key")
	}
	c, err := tencentSMS.NewClient(
		common.NewCredential(secretId, secretKey),
		"ap-nanjing",
		profile.NewClientProfile(),
	)
	if err != nil {
		panic(err)
	}
	return tencent.NewService(c, "1400842696", "妙影科技")
}

func initAliyunSMSService() sms.Service {
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		panic("找不到阿里云 SMS 的 secret id")
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		panic("找不到阿里云 SMS 的 secret key")
	}
	c, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", secretId, secretKey)
	if err != nil {
		panic(err)
	}
	return aliyun.NewService(c, "量链科技")
}
