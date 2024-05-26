.PHONY: mock
mock:
	@mockgen -destination=./webook/internal/service/mocks/user.mock.go -package=svcmocks -source=./webook/internal/service/user.go
	@mockgen -destination=./webook/internal/service/mocks/code.mock.go -package=svcmocks -source=./webook/internal/service/code.go
	@mockgen -destination=./webook/internal/service/mocks/article.mock.go -package=svcmocks -source=./webook/internal/service/article.go
	@mockgen -destination=./webook/internal/service/sms/mocks/sms.mock.go -package=smsmocks -source=./webook/internal/service/sms/types.go
	@mockgen -destination=./webook/internal/repository/mocks/code.mock.go -package=repomocks -source=./webook/internal/repository/code.go
	@mockgen -destination=./webook/internal/repository/mocks/user.mock.go -package=repomocks -source=./webook/internal/repository/user.go
	@mockgen -destination=./webook/internal/repository/mocks/article.mock.go -package=repomocks -source=./webook/internal/repository/article.go
	@mockgen -destination=./webook/internal/repository/mocks/article_author.mock.go -package=repomocks -source=./webook/internal/repository/article_author.go
	@mockgen -destination=./webook/internal/repository/mocks/article_reader.mock.go -package=repomocks -source=./webook/internal/repository/article_reader.go
	@mockgen -destination=./webook/internal/repository/dao/mocks/user.mock.go -package=daomocks -source=./webook/internal/repository/dao/user.go
	@mockgen -destination=./webook/internal/repository/dao/mocks/article_author.mock.go -package=daomocks -source=./webook/internal/repository/dao/article_author.go
	@mockgen -destination=./webook/internal/repository/dao/mocks/article_reader.mock.go -package=daomocks -source=./webook/internal/repository/dao/article_reader.go
	@mockgen -destination=./webook/internal/repository/cache/mocks/user.mock.go -package=cachemocks -source=./webook/internal/repository/cache/user.go
	@mockgen -destination=./webook/internal/repository/cache/mocks/code.mock.go -package=cachemocks -source=./webook/internal/repository/cache/code.go
	@mockgen -destination=./webook/internal/repository/cache/redismocks/cmd.mock.go -package=redismocks github.com/redis/go-redis/v9 Cmdable
	@mockgen -destination=./webook/pkg/limiter/mocks/limiter.mock.go -package=limitermocks -source=./webook/pkg/limiter/types.go
	@go mod tidy