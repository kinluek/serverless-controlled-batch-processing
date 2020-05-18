# Serverless Processing - Controlling Concurrency (JUST FOR FUN! WIP)

An experimental project started out of boredom and curiosity, the project sets out to see how we can use AWS Lambda with it's managed concurrency, and event driven architectures to manage thousands of work queue pipelines to provide optimal throughput for rate limited workloads.

Before you delve into this, it is important to know that this project is just for fun and somewhat of an AWS anti-pattern, anything you decide to take from this, you take at your own risk.

## The Problem

### Sharing Queues

In queue based architectures, tasks are added to queues for them to be processed asynchronously by what ever is consuming the queue on the other side.
The problem arises when you have multiple groups of tasks all being processed from the same queue, and you want each task group to have their own throughput rate. 

For example, given a queue: 

    -> [A, B, C, C, C, C, A, A, B] -> (4 * Consumer)
    
     - consumer concurrency of 4 
     - process time of 1 second
     - rate limit of 2/sec for A tasks
     - rate limit of 1/sec for B tasks
     - rate limit of 2/sec for C tasks

When the consumer starts to pull tasks off the queue, the first 4 will be pulled off and processed at once, 1 B, 2 A's and 1 C,
all falling within their rate limits, this is fine. Once the function has finished processing those tasks, the next 4 will be pulled off.
This time, 3 C's and 1 B, there are now 3 C tasks being processed at once, which exceeds its rate limit.

A scenario in which this may happen, might be where you handle similar tasks using different API keys to access another service as part of the process,
each customer may have access tokens which have different rate limits. 

Another scenario might be that each task group fetches data from a different website, each website will have their own capacity to handle requests,
and you may end up DDoSing some websites that can't handle your throughput. 

There are a number of ways that you can limit the number of concurrent processes for each task group, here are a few:
 - Let the rate limit be hit, make a note of it in some data store and handle the task later
    - This we will need as part of any solution, as there is still a chance tasks will fail even with controlled concurrency.  
 - In a distributed system you could try implement distributed atomic counters and locks to keep track of the number of tasks you are processing.
    - The problem here is that this would be extremely difficult to implement, the atomic counting could be achieved, but with the locking part on top, I'm not sure about that.
 - Don't use a distributed system, process all your tasks through one server, where mutex locks and atomic counters are much easier to implement.
    - here we would have scalability issues.
 - Same as above, but have the tasks partitioned based on task group ID, so that we can scale the system out horizontally.

In all the ways described above, you may be able to stop the task groups from being processed over their rate limit, but you can't guarantee that they will
be processed at their maximum throughput. This is because the tasks are sharing resources (queues and consumers). 

For task groups to be processed with constant throughput, they would each need their own queue and consumer. Here the concurrency can be set on the consumer rather than using counters to track 
how many tasks of the same group are being processed at once. If all the tasks coming in are from the same group, then you know that they will be processed at the rate set by the concurrency of the platform.

In a non-serverless world, this may be very expensive to implement if you have thousands of task groups, as you would have to pay for the running costs for each queue and consumer.

In the serverless world, it would cost you next to nothing, 1 million tasks being sent through 1 AWS SQS queue and consumed by 1 AWS Lambda function, would cost you the same as the 1 million tasks
going through 1000 SQS queues each with their own Lambda consumer. Lambda also allows you to easily configure the concurrency limit for each function.

## My Experimental Solution - (WIP)

So as you can probably tell from above, my solution is to have a separate task pipeline made up of an SQS queue and Lambda handler function for each
task group. 

If we rejig our example from above, it would now look like this:

    -> [A, A, A, A, A, A, A, A, A] -> (2 * Lambda)
    -> [B, B, B, B, B, B, B, B, B] -> (1 * Lambda)
    -> [C, C, C, C, C, C, C, C, C] -> (2 * Lambda)
   
    - process time of 1 second
    - rate limit of 2/sec for A tasks
    - rate limit of 1/sec for B tasks
    - rate limit of 2/sec for C tasks
   
Now each task group can be processed at their maximum throughput limit at a constant rate without having to keep track of the count.

The problem here is that you now have to manage potentially 1000s of task pipelines, and this could get out of hand really quickly. 

What I'm going to attempt in this project, is set up an event driven system where these pipelines are automatically created, updated and removed etc when I add, update or remove a
task config object from a DynamoDB table.

I also want to be able to update my consumer code in one place and then for it to update all my active consumers. The consumers should all be running the same code
but with different concurrency settings.

Here's what we'll need:
 - S3 bucket to hold the source code that Lambda functions will be created from. 
 - SNS topic that the S3 bucket will publish events to on S3 object updates.
 - Source code for your Lambda function consumer in the S3 bucket - we will create an instance of the Lambda for each task group
 - Source code for a Lambda function which gets subscribed to the SNS topic that delivers object update events from the S3 bucket when the consumer code is updated.
   We will create an instance of this function along with every consumer function, so that we can update lambda source codes easily,
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

Given these limits, the best kind of use case for this architecture would be for workloads that come in batches of tasks, that have a maximum rate in which they can be processed.
For example, lets say at most you have to process 50 batches of work at once, some batches could be configured to work through their tasks 100 at a time, while others may have to be 
limited to 10. 

Pipelines could also be removed and replaced with others if they are not being used, since we are just working with SQS and Lambda it would take seconds to swap them out, meaning we could 
have a lot more than the 3800 job configurations, if we are not having to use them all at once, you could set one directly before you use it, then remove it after. 

Finally, since this is all serverless, even with all these pipelines set up, you still won't have to pay a penny
if they don't get used, although you will still have to pay for the Lambda storage costs which is $0.03 GB/month, so that would equate to $2.25 a month if we hit our soft limit storage for Lambda.



### NOTES

- Add Lucid Chart Diagram
