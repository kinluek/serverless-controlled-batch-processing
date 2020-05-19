.PHONY: build clean deploy remove
STAGE?=dev

build:
	env GOOS=linux go build -o bin/manage-pipeline cmd/functions/manage-pipeline/main.go

clean:
	rm -rf ./bin

# deploy the serverless application
deploy: clean build
	sls deploy --verbose -s $(STAGE)

# empty the bucket, this must be done before we remove the deployed stack
empty_bucket:
	aws s3 rm s3://kinluek-serverless-processing-code-$(STAGE) --recursive

remove: empty_bucket
	sls remove -s $(STAGE)

# upload the consume source code to s3.
upload_consumer:
	env GOOS=linux go build -o consume cmd/functions/consume/main.go
	zip consume.zip consume
	aws s3 cp ./consume.zip s3://kinluek-serverless-processing-code-$(STAGE)/consume.zip
	rm consume consume.zip
