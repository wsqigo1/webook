//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/wsqigo/basic-go/webook/internal/events/article"
	"github.com/wsqigo/basic-go/webook/internal/job"
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
	InitSaramaClient,
	InitSyncProducer,
	InitLogger)

var jobProviderSet = wire.NewSet(
	service.NewCronJobService,
	repository.NewPreemptJobRepository,
	dao.NewGORMJobDAO)

var userSvcProvider = wire.NewSet(
	dao.NewUserDao,
	cache.NewUserCache,
	repository.NewUserRepository,
	service.NewUserService)

var articleSvcProvider = wire.NewSet(
	repository.NewCachedArticleRepository,
	cache.NewArticleRedisCache,
	dao.NewArticleGORMDAO,
	service.NewArticleService)

var interactiveSvcSet = wire.NewSet(
	dao.NewGORMInteractiveDAO,
	cache.NewInteractiveRedisCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		articleSvcProvider,
		interactiveSvcSet,

		// cache 部分
		cache.NewCodeCache,

		// repository 部分
		repository.NewCodeRepository,

		article.NewSaramaSyncProducer,

		// Service 部分
		ioc.InitSMSService,
		service.NewCodeService,
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
		userSvcProvider,
		interactiveSvcSet,
		repository.NewCachedArticleRepository,
		cache.NewArticleRedisCache,
		service.NewArticleService,
		article.NewSaramaSyncProducer,
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}

func InitInteractiveService() service.InteractiveService {
	wire.Build(thirdPartySet, interactiveSvcSet)
	return service.NewInteractiveService(nil)
}

func InitJobScheduler() *job.Scheduler {
	wire.Build(jobProviderSet, thirdPartySet, job.NewScheduler)
	return &job.Scheduler{}
}
