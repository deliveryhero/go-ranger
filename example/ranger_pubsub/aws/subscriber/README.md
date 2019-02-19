## Reading messages from aws SQS  
This example demonstrate how you can use the `Subscriber` interface to read message from aws SQS. 

Usage  
-
Example Application
-
The `main.go` file demonstrate how to:

- Prepare configuration for `Subscriber` implementation
- Creating the `Subscriber` instance
- Start reading the messages from queue
- Informing the messages to be processed(ready for delete)
- Stop reading the messages from queue

`Note:`Use appropriate aws configuration to test the implementation.   