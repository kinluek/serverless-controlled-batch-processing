# Serverless Processing - Controlled Batch Throughput (WIP)

An experimental project to see how we can gain controlled throughput/concurrency, with AWS Lambda, 
when processing long-running batched tasks where batches are in parallel with each other.

## TODOS

### Basic 
- [x] Trigger: Create SQS queues when job configs are added to the job config DyanamoDB table.
- [ ] Trigger: Create new Lambda function when SQS queue is created for the job config.
