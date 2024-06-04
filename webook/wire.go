//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/wsqigo/basic-go/webook/internal/events/article"
	"github.com/wsqigo/basic-go/webook/internal/repository"
	"github.com/wsqigo/basic-go/webook/internal/repository/cache"
	"github.com/wsqigo/basic-go/webook/internal/repository/dao"
	"github.com/wsqigo/basic-go/webook/internal/service"
	"github.com/wsqigo/basic-go/webook/internal/web"
	jwt2 "github.com/wsqigo/basic-go/webook/internal/web/jwt"
	"github.com/wsqigo/basic-go/webook/ioc"
)

var interactiveSvcSet = wire.NewSet(dao.NewGORMInteractiveDAO,
	cache.NewInteractiveRedisCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService)

func InitWebServer() *App {
	wire.Build(
		// 第三方依赖
		ioc.InitRedis, ioc.InitDB,
		ioc.InitLogger,
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,
		// DAO 部分
		dao.NewUserDao,
		dao.NewArticleGORMDAO,

		interactiveSvcSet,

		article.NewSaramaSyncProducer,
		article.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,

		// cache 部分
		cache.NewCodeCache, cache.NewUserCache,
		cache.NewArticleRedisCache,

		// repository 部分
		repository.NewUserRepository,
		repository.NewCodeRepository,
		repository.NewCachedArticleRepository,

		// Service 部分
		ioc.InitSMSService,
		ioc.InitDingDingService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,

		// handler 部分
		web.NewUserHandler,
		web.NewArticleHandler,
		web.NewOAuth2DingDingHandler,
		jwt2.NewRedisJWTHandler,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
