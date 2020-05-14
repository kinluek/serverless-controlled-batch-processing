.PHONY: build clean deploy

build:
	env GOOS=linux go build -o bin/initiate-task-queues cmd/functions/initiate-task-queues/main.go
	#env GOOS=linux go build -o bin/initiate-task-processor cmd/functions/initiate-task-processor/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
