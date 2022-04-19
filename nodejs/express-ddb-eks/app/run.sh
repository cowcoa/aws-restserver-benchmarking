#!/bin/bash

# Get script location.
SHELL_PATH=$(cd "$(dirname "$0")";pwd)

aws_account_id="$(aws sts get-caller-identity --output text --query 'Account')"
REGION_CODE="$(jq -r .context.deploymentRegion ${SHELL_PATH}/../cdk.json)"
if [ -z "$REGION_CODE" ]; then
    REGION_CODE="$(aws configure get region)"
fi

TABLE_NAME="$(jq -r .context.tableName ${SHELL_PATH}/../cdk.json)"
GSI_NAME="$(jq -r .context.gsiName ${SHELL_PATH}/../cdk.json)"

# Run rest server by node
#pushd ${SHELL_PATH} &> /dev/null
#AWS_REGION=${REGION_CODE} DDB_TABLE_NAME=${TABLE_NAME} DDB_GSI_NAME=${GSI_NAME} npm start
#popd &> /dev/null

# Run rest server by container
AWS_ACCESS_KEY_ID=$(aws --profile default configure get aws_access_key_id)
AWS_SECRET_ACCESS_KEY=$(aws --profile default configure get aws_secret_access_key)

# app's image repository.
nodejs_express_ddb_repo="$(jq -r .context.imageRepoName ${SHELL_PATH}/../cdk.json)"
# app's image repository URI.
nodejs_express_ddb_repo_uri=$aws_account_id.dkr.ecr.$REGION_CODE.amazonaws.com/$nodejs_express_ddb_repo

docker run \
        -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
        -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
        -e AWS_REGION=$REGION_CODE \
        -e DDB_TABLE_NAME=$TABLE_NAME \
        -e DDB_GSI_NAME=$GSI_NAME \
        -p 8387:8387 ${nodejs_express_ddb_repo_uri}:latest
