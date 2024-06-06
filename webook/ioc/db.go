package ioc

import (
	"fmt"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"github.com/wsqigo/basic-go/webook/internal/repository/dao"
	"github.com/wsqigo/basic-go/webook/pkg/gormx"
	"github.com/wsqigo/basic-go/webook/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/prometheus"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}

	var cfg Config
	err := viper.UnmarshalKey("db", &cfg)
	if err != nil {
		panic(fmt.Errorf("初始化配置失败 %v", err))
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		//// 使用 DEBUG 来打印
		//Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
		//	// 慢查询
		//	SlowThreshold: 0,
		//	LogLevel:      glogger.Info,
		//}),
	})
	if err != nil {
		// 我只会在初始化过程中 panic
		// panic 相当于整个 goroutine 结束
		panic(err)
	}

	// 接入 prometheus
	err = db.Use(prometheus.New(prometheus.Config{
		DBName: "webook",
		// 每 15 秒采集一些数据
		RefreshInterval: 15,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"thread_running"},
			}, // user defined metrics
		},
	}))
	if err != nil {
		panic(err)
	}
	cb := gormx.NewCallbacks(prometheus2.SummaryOpts{
		Namespace: "geektime_daming",
		Subsystem: "webook",
		Name:      "gorm_db",
		Help:      "统计 GORM 的数据库查询",
		ConstLabels: map[string]string{
			"instance_id": "my_instance",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	err = db.Use(cb)
	if err != nil {
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

// 函数衍生类型实现接口
type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...any) {
	g(msg, logger.Field{Key: "args", Val: args})
}
