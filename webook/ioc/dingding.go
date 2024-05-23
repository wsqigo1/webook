package ioc

import (
	"github.com/wsqigo/basic-go/webook/internal/service/oauth2/dingding"
	"os"
)

func InitDingDingService() dingding.Service {
	appId, ok := os.LookupEnv("DINGDING_APP_ID")
	if !ok {
		panic("找不到环境变量 DINGDING_APP_ID")
	}
	appSecret, ok := os.LookupEnv("DINGDING_APP_SECRET")
	if !ok {
		panic("找不到环境变量 DINGDING_APP_SECRET")
	}
	return dingding.NewService(appId, appSecret)
}
