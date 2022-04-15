#!/bin/bash

# Get script location.
SHELL_PATH=$(cd "$(dirname "$0")";pwd)

CDK_CMD=$1
CDK_ACC="$(aws sts get-caller-identity --output text --query 'Account')"
CDK_REGION="$(jq -r .context.deploymentRegion ${SHELL_PATH}/cdk.json)"

# Check execution env.
if [ -z $CODEBUILD_BUILD_ID ]
then
    if [ -z "$CDK_REGION" ]; then
        CDK_REGION="$(aws configure get region)"
    fi

    echo "Run bootstrap..."
    export CDK_NEW_BOOTSTRAP=1 
    npx cdk bootstrap aws://${CDK_ACC}/${CDK_REGION} --cloudformation-execution-policies arn:aws:iam::aws:policy/AdministratorAccess
else
    CDK_REGION=$AWS_DEFAULT_REGION
fi

# CDK command pre-process.
#if [ ! -d "${SHELL_PATH}/imports/k8s" ]; then
#    npx cdk8s import k8s@1.21.0 -l go
#fi
#if [ "$CDK_CMD" == "deploy" ]; then
#    ${SHELL_PATH}/app/deploy_image.sh
#fi

# CDK command.
$SHELL_PATH/cdk-cli-wrapper.sh ${CDK_ACC} ${CDK_REGION} "$@"
cdk_exec_result=$?

# CDK command post-process.
if [ $cdk_exec_result -eq 0 ] && [ "$CDK_CMD" == "destroy" ]; then
    rm -rf $SHELL_PATH/cdk.out/

    nodejs_express_ddb_repo="$(jq -r .context.imageRepoName ${SHELL_PATH}/cdk.json)"
    aws ecr describe-repositories --repository-names $nodejs_express_ddb_repo --region $CDK_REGION &> /dev/null
    result=$?
    if [ $result -eq 0 ]; then
        echo "Delete image repository..."
        aws ecr delete-repository --repository-name $nodejs_express_ddb_repo --region $CDK_REGION --force
    fi
fi
