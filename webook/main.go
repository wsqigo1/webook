package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	server := InitWebServer()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello，启动成功了")
	})
	// 作业：改成 8081
	//server.Run(":8081")
	server.Run(":8080")
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
