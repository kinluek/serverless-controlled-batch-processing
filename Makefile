.PHONY: build clean deploy

build:
	env GOOS=linux go build -o bin/create-queue cmd/functions/create-queue/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose -s dev
