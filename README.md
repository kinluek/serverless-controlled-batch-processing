# Serverless Processing - Controlling Concurrency (JUST FOR FUN! WIP)

An experimental project started out of boredom and curiosity, the project sets out to see how we can use Lambda, with it's managed concurrency and event driven architectures in AWS, to manage thousands of work queue pipelines to provide optimal throughput for rate limited workloads.

Before you delve into this, it is important to know, that this project is just for fun and somewhat of an AWS anti-pattern, anything you decide to take from this, you take at your own risk.

## The Problem

Imagine in your company, you have a Lambda function that processes simple tasks for all your customers, which seems fairly standard I guess?
Now lets say the producer of these tasks just puts them into the same queue along with other customer's tasks and are consumed by the Lambda function. 
For each customer you also want to configure the rate at which these batches are processed for
each customer individually, how would you do that with Lambda? 

The thing is with Lambda, you can only set the concurrency limit per function, so if you set the limit to be 10,
that means 10 tasks can be processed at a time, and each customer will get an undetermined fraction of that.

For example, given a queue: ->[A, B, C, C, C, C, A, A, B]->
 - Lambda concurrency of 4 
 - process time of 1 second
 - rate limit of 1/sec for B tasks
 - rate limit of 2/sec for A tasks
 - rate limit of 2/sec for C tasks

When Lambda starts to pull tasks off this queue, the first 4 will be pulled off and processed at once, 1 B, 2 A's and 1 C,
all falling within their rate limits, this is fine. Once the function has finished processing those tasks, the next 4 will be pulled off.
This time, 3 C's and 1 B, there are now 3 C tasks being processed at once, which exceeds its rate limit.

A scenario in which this may happen, might be where you handle similar tasks using different API keys to access another service as part of the process,
each customer may have access tokens which have different rate limits. 

Another scenario might be that each task group fetches data from a different website, each website will have their own capacity to handle requests,
and you may end up DDoSing some websites that can't handle your throughput. 

There are a number of ways this can be solved, here are a few of them:
 1. Let the rate limit be hit, make a note of it in some data store and handle the task later
    - This we will need as part of any solution, as there is still a chance tasks will fail even with controlled concurrency. 
 2. Use distributed counters and locks for each task group, to make sure none of them can go over the limit.
    - The problem here is that distributed locks and counters add complexity and are not easily achieved.
    - Also, Lambdas would still have to pull the task off the queue to find out whether they are at the limit, and therefore using up resources.
 3. Don't use Lambda, process all your tasks through one server, where mutex locks and atomic counters are much easier to implement.
    - Here we would have scalability issues, as we can only scale vertically so far.
 4. Partition your tasks based on rate/concurrency limits and send them to different consumers that have that use the same amount of concurrency.
    - Here you have to manage more infrastructure.
    - although this will stop the tasks going over their concurrency limit, task groups still have to share processing power with other groups on the same partition, meaning they still
    don't get the full amount of concurrency they can handle.
 5. MY EXPERIMENTAL SOLUTION: use a separate SQS queue and Lambda function for each task group.
    - Here there is even more infracture to manage, but with the right setup, this can be easily managed (in theory).

## My Experimental Solution - (WIP)
Use a separate queue and Lambda function for each task group, that are created through triggers and events when new process configurations are register.
This way each set of tasks can be processed at their maximum through put consistently, without being blocked by tasks from other groups in the queue.

Here's what we'll need:
 - S3 bucket to hold the source code that Lambda functions will be created from. 
 - SNS topic that the S3 bucket will publish events to on S3 object updates.
 - Source code for your task processor Lambda function in the S3 bucket - we will create an instance of the Lambda for each task group
 - Source code for a Lambda function which gets subscribed to an SNS topic that delivers object update events from the S3 bucket when the task processor source code is updated.
   We will create an instance of this function along with every task processor function, so that we can update lambda source codes easily,
   as by default, your Lambda function does not automatically update when you update the source code for it in S3.
 - A DynamoDB table to hold process configurations for each task group, with streams enabled.
 - A stream listener Lambda function, which listens to the DyamoDB streams, and creates new instances of the task processor function along with SQS queues for it to pull from when new configurations are added.
   The configuration item will hold the SQS and Lambda configuration options (concurrency, visibility timeouts, environment variables etc)
 - The stream listener should also be able to update the Lambda and SQS configuration when the configuration object is updated.
 - The stream listener should also be able to remove all the resources for a process coniguration when that object is deleted.
 
Once we have all this set up, we should just be able to add new configuration object to DynamoDB and have the triggers do the rest of the work
to set up a whole new task pipe line for us, each with their own concurrency limits. We should be able to update the configuration pipelines by just updating 
the configuration object for it, in DynamoDB. We should be to update the source code for our task processor in one place and then allow the update
to propagate to all our Lambda function instances of it. 

Resource Limitations:
  - No limit on the number of SQS queues that can be created.
  - 10 million subscribers per SNS topic, and we can have 100,000 topics per AWS account per region. This means we could update the source code for 10 million of our Lambda instances at once with our one topic. If you ever reach the limit, 
    then you could just create a new topic and publish the event to two topics.
  - The is no limit on the number of Lambda functions you can have, but there is a limit on the amount of source code storage you can have, stored within the Lambda service. 
    The limit is 75GB per account per region. Even though we will only have one source of truth for our source code, Lambda actually copies this code into the Lambda service for each function that is created from it.
    This then becomes our limiting factor. Lets for each process configuration we need 20MB of Lambda storage, that means we could have around 3800 different pipelines at most, if we used a dedicated account for these pipelines. 
    This is a soft limit though, so this limit can be increased if you give AWS a good reason to... For this, they would probably tell you, you're using the service incorrectly and decline.
  - 1000 concurrency limit across all Lambda functions per account per region, this is also a soft limit which can be increased.
    So, if we stayed at this limit, even if we did set up 3800 pipelines, each with their own concurrency, we still wouldn't even be able to run all of them at once. 

Given these limits, the best use case for this architecture would be for work loads that come in batches of tasks, that have a maximum rate in which they can be processed.
For example, lets say at most you have to process 50 batches of work at once, some batches could be configured to work through their tasks 100 at a time, while others may have to be 
limited to 10. Pipelines could also be removed and replaced with others if they are not being used, since we are just working with SQS and Lambda it would take seconds to swap them out, meaning we could 
have a lot more than the 3800 job configurations, if we are not having to use them all at once. Finally, since this is all serverless, even with all these pipelines set up, you still won't have to pay a penny
if they don't get used, although you will still have to pay for the Lambda storage costs which is $0.03 GB/month, so that would equate to $2.25 a month if we hit our soft limit storage for Lambda.

### NOTES

- Add Lucid Chart Diagram
