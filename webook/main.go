package main

import (
	"bytes"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
)

func main() {
	initViperV1()
	initLogger()
	server := InitWebServer()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello，启动成功了")
	})
	// 作业：改成 8081
	//server.Run(":8081")
	server.Run(":8080")
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	// 设置全局的 logger
	// 你在你的代理里面就可以直接使用 zap.XXX 来记录日志
	zap.ReplaceGlobals(logger)
}

func initViper() {
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	// 当前工作目录的 config 子目录
	viper.AddConfigPath("config")
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	fmt.Print(viper.Get("test.key"))
}

func initViperWatch() {
	cfile := pflag.String("config",
		"config/config.yaml", "配置文件路径")
	// 这一步之后，cfile 里面才有值
	pflag.Parse()
	// Viper.Set("db.dsn", "localhost:3306")
	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Print(viper.GetString("redis.addr"))
	})
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	fmt.Print(viper.Get("test.key"))
}

func initViperV1() {
	cfile := pflag.String("config",
		"config/dev.yaml", "配置文件路径")
	// 这一步之后，cfile 里面才有值
	pflag.Parse()
	// Viper.Set("db.dsn", "localhost:3306")
	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	fmt.Print(viper.Get("test.key"))
}

func initViperV2() {
	cfg := `
test:
  key: value1

redis:
  addr: "localhost:6379"

db:
  dsn: "root:root@tcp(localhost:13316)/webook"
`
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewReader([]byte(cfg)))
	if err != nil {
		panic(err)
	}
}

func initViperRemote() {
	err := viper.AddRemoteProvider("etcd3",
		"http://localhost:12379", "C:/Program Files/Git/webook")
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			err = viper.WatchRemoteConfig()
			if err != nil {
				panic(err)
			}
			log.Println("watch", viper.GetString("redis.addr"))
			time.Sleep(3 * time.Second)
		}
	}()
}

func useSession(server *gin.Engine) {
	// 存储数据的，也就是你 userId 存哪里
	// 直接存 cookie
	store := cookie.NewStore([]byte("secret"))
	// 基于内存的实现，第一个参数是 authentication key，最好是 32 位或者 64 位
	// 第二个参数是 encryption key
	//store := memstore.NewStore([]byte("S4EWBerIvPWZCfH8jpFRBByIE5HcBfiP"),
	//	[]byte("anouWji8NjQi8wJ1LUI4TyZMM5xTz2zZ"))
	// 第一个参数是最大空闲连接数
	// 第二个参数就是 tcp，你不太可能用 udp
	// 第三、四个就是连接信息和密码
	// 第五和第六就是两个 key
	//store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
	//	[]byte("S4EWBerIvPWZCfH8jpFRBByIE5HcBfiP"), []byte("anouWji8NjQi8wJ1LUI4TyZMM5xTz2zZ"))
	//if err != nil {
	//	panic(err)
	//}
	// cookie 的名字叫做 mysession
	server.Use(sessions.Sessions("mysession", store))
}
