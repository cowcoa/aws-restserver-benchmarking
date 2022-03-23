package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"nodejs-express-ddb/config"
)

type NodejsExpressDdbStackProps struct {
	awscdk.StackProps
}

func NewNodejsExpressDdbStack(scope constructs.Construct, id string, props *NodejsExpressDdbStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create DynamoDB Base table.
	// Data Modeling
	// name(PK), time(SK),                  comment, chat_room
	// string    string(micro sec unixtime)	string   string
	chatTable := awsdynamodb.NewTable(stack, jsii.String(config.TableName(stack)), &awsdynamodb.TableProps{
		TableName:     jsii.String(config.TableName(stack)),
		BillingMode:   awsdynamodb.BillingMode_PROVISIONED,
		ReadCapacity:  jsii.Number(1),
		WriteCapacity: jsii.Number(1),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("Name"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("Time"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		PointInTimeRecovery: jsii.Bool(true),
	})

	// Create DynamoDB GSI table.
	// Data Modeling
	// chat_room(PK), time(SK),                  comment, name
	// string         string(micro sec unixtime) string   string
	chatTable.AddGlobalSecondaryIndex(&awsdynamodb.GlobalSecondaryIndexProps{
		IndexName: jsii.String(config.GsiName(stack)),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("ChatRoom"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("Time"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		ProjectionType: awsdynamodb.ProjectionType_ALL,
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewNodejsExpressDdbStack(app, config.StackName(app), &NodejsExpressDdbStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	account := os.Getenv("CDK_DEPLOY_ACCOUNT")
	region := os.Getenv("CDK_DEPLOY_REGION")

	if len(account) == 0 || len(region) == 0 {
		account = os.Getenv("CDK_DEFAULT_ACCOUNT")
		region = os.Getenv("CDK_DEFAULT_REGION")
	}

	return &awscdk.Environment{
		Account: jsii.String(account),
		Region:  jsii.String(region),
	}
}
