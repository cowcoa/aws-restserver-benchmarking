package config

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// DO NOT modify this function, change stack name by 'cdk.json/context/stackName'.
func StackName(scope constructs.Construct) string {
	stackName := "MyStack"

	ctxValue := scope.Node().TryGetContext(jsii.String("stackName"))
	if v, ok := ctxValue.(string); ok {
		stackName = v
	}

	return stackName
}

// DO NOT modify this function, change DDB table name by 'cdk.json/context/tableName'.
func TableName(scope constructs.Construct) string {
	tableName := "MyTable"

	ctxValue := scope.Node().TryGetContext(jsii.String("tableName"))
	if v, ok := ctxValue.(string); ok {
		tableName = v
	}

	return tableName
}

// DO NOT modify this function, change DDB gsi name by 'cdk.json/context/gsiName'.
func GsiName(scope constructs.Construct) string {
	gsiName := "MyGSI"

	ctxValue := scope.Node().TryGetContext(jsii.String("gsiName"))
	if v, ok := ctxValue.(string); ok {
		gsiName = v
	}

	return gsiName
}
