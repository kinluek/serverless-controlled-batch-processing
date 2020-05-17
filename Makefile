.PHONY: build clean deploy

build:
	env GOOS=linux go build -o bin/manage-pipeline cmd/functions/manage-pipeline/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose -s dev
