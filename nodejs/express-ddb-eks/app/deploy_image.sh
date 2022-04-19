#!/bin/bash

# Get script location.
SHELL_PATH=$(cd "$(dirname "$0")";pwd)

aws_account_id="$(aws sts get-caller-identity --output text --query 'Account')"
deployment_region="$(jq -r .context.deploymentRegion ${SHELL_PATH}/../cdk.json)"
if [ -z "$deployment_region" ]; then
    deployment_region="$(aws configure get region)"
fi

# app's image repository.
nodejs_express_ddb_repo="$(jq -r .context.imageRepoName ${SHELL_PATH}/../cdk.json)"
# app's image repository URI.
nodejs_express_ddb_repo_uri=$aws_account_id.dkr.ecr.$deployment_region.amazonaws.com/$nodejs_express_ddb_repo

aws ecr describe-repositories --repository-names $nodejs_express_ddb_repo --region $deployment_region &> /dev/null
result=$?
if [ $result -ne 0 ]; then
    echo "Create image repository..."
    aws ecr create-repository --repository-name $nodejs_express_ddb_repo --region $deployment_region
fi

echo "Login to ECR..."
DOCKER_LOGIN_CMD=$(aws ecr get-login --no-include-email --region $deployment_region)
eval "${DOCKER_LOGIN_CMD}"

echo "Build docker image..."
image_tag="$(echo $(date '+%Y.%m.%d.%H%M%S' -d '+8 hours'))"
docker buildx build --push --platform linux/arm64,linux/amd64 --tag $nodejs_express_ddb_repo_uri:latest --tag $nodejs_express_ddb_repo_uri:$image_tag ${SHELL_PATH}/.

echo
echo "Done"
