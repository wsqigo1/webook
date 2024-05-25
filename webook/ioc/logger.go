package ioc

import (
	"github.com/spf13/viper"
	"github.com/wsqigo/basic-go/webook/pkg/logger"
	"go.uber.org/zap"
)

func InitLogger() logger.LoggerV1 {
	// 这里我们用一个小技巧
	// 就是直接使用 zap 本身的配置结构体来处理
	cfg := zap.NewDevelopmentConfig()
	err := viper.UnmarshalKey("log", &cfg)
	if err != nil {
		panic(err)
	}

	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}
