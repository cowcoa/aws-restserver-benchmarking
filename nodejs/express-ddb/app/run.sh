#!/bin/bash

# Get script location.
SHELL_PATH=$(cd "$(dirname "$0")";pwd)

REGION_CODE="$(jq -r .context.deploymentRegion ${SHELL_PATH}/../cdk.json)"
if [ -z "$REGION_CODE" ]; then
    REGION_CODE="$(aws configure get region)"
fi

TABLE_NAME="$(jq -r .context.tableName ${SHELL_PATH}/../cdk.json)"
GSI_NAME="$(jq -r .context.gsiName ${SHELL_PATH}/../cdk.json)"

pushd ${SHELL_PATH} &> /dev/null
AWS_REGION=${REGION_CODE} DDB_TABLE_NAME=${TABLE_NAME} DDB_GSI_NAME=${GSI_NAME} npm start
popd &> /dev/null
