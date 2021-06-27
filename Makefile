.PHONY: build clean deploy

CMD_DIR = ./cmd

build:
	@for f in $(shell ls ${CMD_DIR}); do echo Building $${f} && env GOOS=linux go build -ldflags="-s -w" -o bin/$${f} cmd/$${f}/*.go; done

clean:
	rm -rf ./bin

deploy-staging: clean build
	sls deploy --stage=staging --verbose

deploy-production: clean build
	sls deploy --stage=production --verbose
