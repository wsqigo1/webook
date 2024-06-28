// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

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
	"github.com/wsqigo/basic-go/webook/internal/web/jwt"
	"github.com/wsqigo/basic-go/webook/ioc"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	loggerV1 := InitLogger()
	v := ioc.InitGinMiddlewares(cmdable, handler, loggerV1)
	db := InitDB()
	userDAO := dao.NewUserDao(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, handler, codeService)
	articleDAO := dao.NewArticleGORMDAO(db)
	articleCache := cache.NewArticleRedisCache(cmdable)
	articleRepository := repository.NewCachedArticleRepository(articleDAO, userRepository, articleCache, loggerV1)
	client := InitSaramaClient()
	syncProducer := InitSyncProducer(client)
	producer := article.NewSaramaSyncProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer)
	interactiveDAO := dao.NewGORMInteractiveDAO(db)
	interactiveCache := cache.NewInteractiveRedisCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDAO, loggerV1, interactiveCache)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	articleHandler := web.NewArticleHandler(loggerV1, articleService, interactiveService)
	dingdingService := InitDingDingService(loggerV1)
	oAuth2DingDingHandler := web.NewOAuth2DingDingHandler(dingdingService, handler, userService)
	engine := ioc.InitWebServer(v, userHandler, articleHandler, oAuth2DingDingHandler)
	return engine
}

func InitArticleHandler(dao2 dao.ArticleDAO) *web.ArticleHandler {
	loggerV1 := InitLogger()
	db := InitDB()
	userDAO := dao.NewUserDao(db)
	cmdable := InitRedis()
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	articleCache := cache.NewArticleRedisCache(cmdable)
	articleRepository := repository.NewCachedArticleRepository(dao2, userRepository, articleCache, loggerV1)
	client := InitSaramaClient()
	syncProducer := InitSyncProducer(client)
	producer := article.NewSaramaSyncProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer)
	interactiveDAO := dao.NewGORMInteractiveDAO(db)
	interactiveCache := cache.NewInteractiveRedisCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDAO, loggerV1, interactiveCache)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	articleHandler := web.NewArticleHandler(loggerV1, articleService, interactiveService)
	return articleHandler
}

func InitInteractiveService() service.InteractiveService {
	db := InitDB()
	interactiveDAO := dao.NewGORMInteractiveDAO(db)
	loggerV1 := InitLogger()
	cmdable := InitRedis()
	interactiveCache := cache.NewInteractiveRedisCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDAO, loggerV1, interactiveCache)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	return interactiveService
}

func InitJobScheduler() *job.Scheduler {
	db := InitDB()
	jobDAO := dao.NewGORMJobDAO(db)
	cronJobRepository := repository.NewPreemptJobRepository(jobDAO)
	loggerV1 := InitLogger()
	cronJobService := service.NewCronJobService(cronJobRepository, loggerV1)
	scheduler := job.NewScheduler(cronJobService, loggerV1)
	return scheduler
}

// wire.go:

var thirdPartySet = wire.NewSet(

	InitRedis, InitDB,
	InitSaramaClient,
	InitSyncProducer,
	InitLogger)

var jobProviderSet = wire.NewSet(service.NewCronJobService, repository.NewPreemptJobRepository, dao.NewGORMJobDAO)

var userSvcProvider = wire.NewSet(dao.NewUserDao, cache.NewUserCache, repository.NewUserRepository, service.NewUserService)

var articleSvcProvider = wire.NewSet(repository.NewCachedArticleRepository, cache.NewArticleRedisCache, dao.NewArticleGORMDAO, service.NewArticleService)

var interactiveSvcSet = wire.NewSet(dao.NewGORMInteractiveDAO, cache.NewInteractiveRedisCache, repository.NewCachedInteractiveRepository, service.NewInteractiveService)
