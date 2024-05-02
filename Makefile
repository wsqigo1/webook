.PHONY: mock
mock:
	@mockgen -destination=./webook/internal/service/mocks/user.mock.go -package=svcmocks -source=./webook/internal/service/user.go
	@mockgen -destination=./webook/internal/service/mocks/code.mock.go -package=svcmocks -source=./webook/internal/service/code.go
	@go mod tidy