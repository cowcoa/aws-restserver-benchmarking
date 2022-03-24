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

echo "Build docker image..."
docker build -t $nodejs_express_ddb_repo $SHELL_PATH

image_tag="$(echo $(date '+%Y.%m.%d.%H%M%S' -d '+8 hours'))"

echo "nodejs_express_ddb_repo: $nodejs_express_ddb_repo"
echo "nodejs_express_ddb_repo_uri: $nodejs_express_ddb_repo_uri"

echo "Upload docker image to ECR..."
DOCKER_LOGIN_CMD=$(aws ecr get-login --no-include-email --region $deployment_region)
eval "${DOCKER_LOGIN_CMD}"
docker tag $nodejs_express_ddb_repo:latest $nodejs_express_ddb_repo_uri:$image_tag
docker push $nodejs_express_ddb_repo_uri:$image_tag
docker tag $nodejs_express_ddb_repo:latest $nodejs_express_ddb_repo_uri:latest
docker push $nodejs_express_ddb_repo_uri:latest

echo
echo "Done"
