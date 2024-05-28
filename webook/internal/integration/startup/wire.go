//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/wsqigo/basic-go/webook/internal/repository"
	"github.com/wsqigo/basic-go/webook/internal/repository/cache"
	"github.com/wsqigo/basic-go/webook/internal/repository/dao"
	"github.com/wsqigo/basic-go/webook/internal/service"
	"github.com/wsqigo/basic-go/webook/internal/web"
	jwt2 "github.com/wsqigo/basic-go/webook/internal/web/jwt"
	"github.com/wsqigo/basic-go/webook/ioc"
)

var thirdPartySet = wire.NewSet(
	// 第三方依赖
	InitRedis, InitDB,
	InitLogger)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		// DAO 部分
		dao.NewUserDao,
		dao.NewArticleGORMDAO,

		// cache 部分
		cache.NewCodeCache, cache.NewUserCache,

		// repository 部分
		repository.NewUserRepository,
		repository.NewCodeRepository,
		repository.NewCachedArticleRepository,

		// Service 部分
		ioc.InitSMSService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,
		InitDingDingService,

		// handler 部分
		web.NewUserHandler,
		jwt2.NewRedisJWTHandler,
		web.NewOAuth2DingDingHandler,
		web.NewArticleHandler,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}

func InitArticleHandler(dao dao.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		repository.NewCachedArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}
