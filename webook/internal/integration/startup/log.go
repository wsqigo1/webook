package startup

import "github.com/wsqigo/basic-go/webook/pkg/logger"

func InitLogger() logger.LoggerV1 {
	return logger.NewNopLogger()
}
