## Simple REST server PoC (Node.js + Express + DynamoDB)
Deploy and run this PoC on your x86_64 or arm64 EC2 instance.<br />

## Deployment
Run the following command to deploy AWS resources by CDK Toolkit:<br />
  ```sh
  cdk-cli-wrapper-dev.sh deploy
  ```
If all goes well, you will see the following output:<br />
  ```sh
  Outputs:
  RESTBenchmark-NodeExpressDDB.EcrRepositoryName = restbenchmark-nodeexpressddb
  RESTBenchmark-NodeExpressDDB.EcrRepositoryUri = 123456789012.dkr.ecr.ap-northeast-2.amazonaws.com/restbenchmark-nodeexpressddb
  Stack ARN:
  arn:aws:cloudformation:ap-northeast-2:123456789012:stack/RESTBenchmark-NodeExpressDDB/52be8470-aab6-11ec-a9b8-0656cd628778
  
  ✨  Total time: 5.24s
  ```
You can also clean up the deployment by running command:<br />
  ```sh
  cdk-cli-wrapper-dev.sh destroy
  ```

## Build Docker image and Push to ECR repository
Run the following command to build docker image:<br />
  ```sh
  cd app/
  ./deploy_image.sh
  ```
This script will help you build an image of your Node App and automatically push it to the ECR repository created in the previous step.

## Testing
Run the following command to launch rest server in docker container:<br />
  ```sh
  cd app/
  ./run.sh
  
  > nodejs-express-ddb@1.0.0 start
  > node server.js
  
  Server running on port 8387
  ```
You can POST user comment by command:
  ```sh
  curl -v -X POST http://127.0.0.1:8387/put -H 'Content-Type: application/json' -d '{"name": "Cow","comment": "sample comment!","chatRoom": "123"}'
  ```
Or you can QUERY user comments by command:
  ```sh
  curl -v http://127.0.0.1:8387/get?chatroom=123
  ```
  
  
