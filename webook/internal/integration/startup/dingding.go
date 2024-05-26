package startup

import (
	"github.com/wsqigo/basic-go/webook/internal/service/oauth2/dingding"
	"github.com/wsqigo/basic-go/webook/pkg/logger"
)

func InitDingDingService(l logger.LoggerV1) dingding.Service {
	return dingding.NewService("", "", l)
}
