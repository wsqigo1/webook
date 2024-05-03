.PHONY: mock
mock:
	@mockgen -destination=./webook/internal/service/mocks/user.mock.go -package=svcmocks -source=./webook/internal/service/user.go
	@mockgen -destination=./webook/internal/service/mocks/code.mock.go -package=svcmocks -source=./webook/internal/service/code.go
	@mockgen -destination=./webook/internal/repository/mocks/user.mock.go -package=repomocks -source=./webook/internal/repository/user.go
	@mockgen -destination=./webook/internal/repository/mocks/code.mock.go -package=repomocks -source=./webook/internal/repository/code.go
	@mockgen -destination=./webook/internal/repository/dao/mocks/user.mock.go -package=daomocks -source=./webook/internal/repository/dao/user.go
	@mockgen -destination=./webook/internal/repository/cache/mocks/user.mock.go -package=cachemocks -source=./webook/internal/repository/cache/user.go
	@mockgen -destination=./webook/internal/repository/cache/mocks/code.mock.go -package=cachemocks -source=./webook/internal/repository/cache/code.go
	@mockgen -destination=./webook/internal/repository/cache/redismocks/cmd.mock.go -package=redismocks github.com/redis/go-redis/v9 Cmdable
	@go mod tidy