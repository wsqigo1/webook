package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/wsqigo/basic-go/webhook/internal/repository"
	"github.com/wsqigo/basic-go/webhook/internal/repository/dao"
	"github.com/wsqigo/basic-go/webhook/internal/service"
	"github.com/wsqigo/basic-go/webhook/internal/web"
	"github.com/wsqigo/basic-go/webhook/internal/web/middleware"
	"github.com/wsqigo/basic-go/webhook/pkg/ginx/middleware/ratelimit"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func main() {
	db := initDB()
	server := initWebServer()

	initUserHdl(db, server)
	server.Run(":8080")
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	server.Use(cors.New(cors.Config{
		//AllowAllOrigins: true,
		//AllowOrigins:     []string{"http://localhost:3000"},
		//AllowMethods:     []string{"POST", "GET"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 这个是允许前端访问你的后端响应中带的头部
		ExposeHeaders: []string{"x-jwt-token"},
		// 是否允许带上用户认证信息，比如 cookie 之类的东西
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				// 你的开发环境
				return true
			}
			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	server.Use(ratelimit.NewBuilder(redisClient,
		time.Second, 1).Build())

	useJWT(server)
	//useSession(server)

	return server
}

func useJWT(server *gin.Engine) {
	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login").CheckLogin())
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
	server.Use(middleware.NewLoginMiddlewareBuilder().
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login").CheckLogin())
}

func initUserHdl(db *gorm.DB, server *gin.Engine) {
	ud := dao.NewUserDao(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	u.RegisterRoutes(server)
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	if err != nil {
		// 我只会在初始化过程中 panic
		// panic 相当于整个 goroutine 结束
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
