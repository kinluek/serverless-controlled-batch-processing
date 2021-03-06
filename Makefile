.PHONY: build clean deploy remove
STAGE?=dev
NAME_SPACE?=kinluek

build:
	env GOOS=linux go build -o bin/manage-pipeline cmd/functions/manage-pipeline/main.go

clean:
	rm -rf ./bin

# deploy the serverless application
deploy: clean build
	env NAME_SPACE=$(NAME_SPACE) sls deploy --verbose -s $(STAGE)

# empty the bucket, this must be done before we can remove the deployed stack
empty_bucket:
	aws s3 rm s3://$(NAME_SPACE)-serverless-processing-code-$(STAGE) --recursive

# remove the deployed service
remove: empty_bucket
	env NAME_SPACE=$(NAME_SPACE) sls remove -s $(STAGE)

# upload the consume source code to s3.
upload_consumer:
	env GOOS=linux go build -o consume cmd/functions/consume/main.go
	zip consume.zip consume
	aws s3 cp ./consume.zip s3://$(NAME_SPACE)-serverless-processing-code-$(STAGE)/consume.zip
	rm consume consume.zip
