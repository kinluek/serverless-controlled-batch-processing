.PHONY: build clean deploy

build:
	env GOOS=linux go build -o bin/initiate-task-queues cmd/functions/initiate-task-queues/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
