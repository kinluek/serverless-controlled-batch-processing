.PHONY: build clean deploy remove

build:
	env GOOS=linux go build -o bin/manage-pipeline cmd/functions/manage-pipeline/main.go

clean:
	rm -rf ./bin

# deploy the serverless application
deploy: clean build
	sls deploy --verbose -s dev

# empty the bucket, this must be done before we remove the deployed stack
empty_bucket:
	aws s3 rm s3://kinluek-serverless-processing-code --recursive

remove: empty_bucket
	sls remove -s dev

# upload the consume source code to s3.
upload_consumer:
	env GOOS=linux go build -o consumer cmd/functions/consumer/main.go
	zip consumer.zip consumer
	aws s3 cp ./consumer.zip s3://kinluek-serverless-processing-code/consumer.zip
	rm consumer consumer.zip
