.PHONY: mock
mock:
	@go generate ./...
	@go mod tidy