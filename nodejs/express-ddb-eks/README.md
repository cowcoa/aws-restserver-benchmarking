## Simple REST server PoC (Node.js + Express + DynamoDB)
Deploy and run this PoC on your x86_64 or arm64 EKS cluster.<br />

## Prerequisites
1. [Init](https://github.com/cowcoa/aws-restserver-benchmarking) - Follow this to initialize the development environment.
2. Install CDK8s Toolkit:
    ```sh
    npm install -g cdk8s-cli
    ```

## Build Docker image and Push to ECR repository
Run the following command to build docker image:<br />
  ```sh
  cd app/
  ./deploy_image.sh
  ```
This script will help you build an image of your Node App and push it to the ECR repository created automatically.

## Deployment
Generate low-level (L1) constructs for Kubernetes API objects and Custom Resources (CRDs):<br />
  ```sh
  cdk8s import k8s@1.21.0 -l go
  ```
Export the following information from your EKS cluster and fill in the cluster-info.json file:<br />
| Name | Example Value |
| ------ | ------ |
| clusterSecurityGroupId | sg-0cb7ee5b03a23bb74 |
| apiServerEndpoint | https://AB123D8E12345CD123AA92855957B4F8.gr7.ap-northeast-1.eks.amazonaws.com |
| vpcId | vpc-0445143cc39ee48f6 |
| clusterName | CDKGoExample-EKSCluster |
| certificateAuthorityData | LS0tLS1CRUdJTi...BDRVJUSU0tCg== |
| kubectlRoleArn | arn:aws:iam::123456789012:role/CDKGoExample-EKSCluster-EksClusterCreationRole75AA-1UKOP8JQ8R9DN |
| region | ap-northeast-1 |
| oidcIdpArn | arn:aws:iam::123456789012:oidc-provider/oidc.eks.ap-northeast-1.amazonaws.com/id/AB123D8E12345CD123AA92855957B4F8 |

If you don't know how to do it, you can refer to [this].

Run the following command to deploy AWS resources by CDK Toolkit:<br />
  ```sh
  cdk-cli-wrapper-dev.sh deploy
  ```
If all goes well, you will see the following output:<br />
  ```sh
  ✅  RESTBenchmark-NodeExpressDDB-EKS
  
  ✨  Deployment time: 44.65s
  
  Stack ARN:
  arn:aws:cloudformation:ap-northeast-1:123456789012:stack/RESTBenchmark-NodeExpressDDB-EKS/7794aa30-bb42-11ec-ac27-06e90b0496e1
  
  ✨  Total time: 48.96s
  ```
You can also clean up the deployment by running command:<br />
  ```sh
  cdk-cli-wrapper-dev.sh destroy
  ```

## Testing
Run the following command to launch rest server in docker container:<br />
  ```sh
  cd app/
  ./run.sh
  Server running on port 8387
  ```
You can POST user comment by command:
  ```sh
  curl -v -X POST http://127.0.0.1:8387/put -H 'Content-Type: application/json' -d '{"name": "Cow","comment": "sample comment!","chatRoom": "123"}'
  Status Code: 201 Created
  ```
Or you can QUERY user comments by command:
  ```sh
  curl -v http://127.0.0.1:8387/get?chatroom=123
  Status Code: 200 OK
  [
    {
      "name"   :"Cow",
      "comment":"sample comment!",
      "time"   :"2022-03-23T22:59:04+08:00"
    },
    ...
  ]
  ```

[this]: <https://github.com/cowcoa/aws-cdk-go-examples/tree/master/eks/simple-cluster/>
  
