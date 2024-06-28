docker:
	@rm demojob || true
	@GOOS=linux GOARCH=arm go build -o demojob .
	@docker rmi -f flycash/cron_job:v0.0.1
	@GOOS=linux GOARCH=arm go build -tags=k8s -o demojob .
	@docker build -t flycash/cron_job:v0.0.1 .