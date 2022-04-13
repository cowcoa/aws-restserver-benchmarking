package config

import (
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/jsii-runtime-go"
)

// DO NOT modify this function, change stack name by 'cdk.json/context/stackName'.
func StackName(scope awscdk.App) string {
	stackName := "MyStack"

	ctxValue := scope.Node().TryGetContext(jsii.String("stackName"))
	if v, ok := ctxValue.(string); ok {
		stackName = v
	}

	return stackName
}

// DO NOT modify this function, change DDB table name by 'cdk.json/context/tableName'.
func TableName(scope awscdk.Stack) string {
	tableName := "MyTable"

	ctxValue := scope.Node().TryGetContext(jsii.String("tableName"))
	if v, ok := ctxValue.(string); ok {
		tableName = v
	}

	return tableName
}

// DO NOT modify this function, change DDB gsi name by 'cdk.json/context/gsiName'.
func GsiName(scope awscdk.Stack) string {
	gsiName := "MyGSI"

	ctxValue := scope.Node().TryGetContext(jsii.String("gsiName"))
	if v, ok := ctxValue.(string); ok {
		gsiName = v
	}

	return gsiName
}

// DO NOT modify this function, change ECR repository name by 'cdk.json/context/imageRepoName'.
func EcrRepoName(scope awscdk.Stack) string {
	ecrRepoName := "MyRepository"

	ctxValue := scope.Node().TryGetContext(jsii.String("imageRepoName"))
	if v, ok := ctxValue.(string); ok {
		ecrRepoName = v
	}

	return ecrRepoName
}
