package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"nodejs-express-ddb-eks/config"
	"nodejs-express-ddb-eks/imports/k8s"

	"github.com/cdk8s-team/cdk8s-core-go/cdk8s"
	"github.com/cdk8s-team/cdk8s-plus-go/cdk8splus21"

	"github.com/aws/constructs-go/constructs/v3"
	"github.com/aws/jsii-runtime-go"

	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/awseks"
	"github.com/aws/aws-cdk-go/awscdk/awsiam"
)

type ClusterInfo struct {
	ClusterName              string `json:"clusterName"`
	ApiServerEndpoint        string `json:"apiServerEndpoint"`
	KubectlRoleArn           string `json:"kubectlRoleArn"`
	OidcIdpArn               string `json:"oidcIdpArn"`
	ClusterSecurityGroupId   string `json:"clusterSecurityGroupId"`
	Region                   string `json:"region"`
	VpcId                    string `json:"vpcId"`
	CertificateAuthorityData string `json:"certificateAuthorityData"`
}

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
		TableName:   jsii.String(config.TableName(stack)),
		BillingMode: awsdynamodb.BillingMode_PAY_PER_REQUEST,
		// ReadCapacity:  jsii.Number(1),
		// WriteCapacity: jsii.Number(1),
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
		// ReadCapacity:  jsii.Number(1),
		// WriteCapacity: jsii.Number(1),
	})

	// Deploy app to k8s cluster.
	// Uns11n cluster info.
	clusterInfoFile, _ := ioutil.ReadFile("./cluster-info.json")
	clusterInfo := ClusterInfo{}
	json.Unmarshal(clusterInfoFile, &clusterInfo)

	// Import cluster.
	cluster := awseks.Cluster_FromClusterAttributes(stack, jsii.String("mycluster"), &awseks.ClusterAttributes{
		ClusterName:                     jsii.String(clusterInfo.ClusterName),
		ClusterCertificateAuthorityData: jsii.String(clusterInfo.CertificateAuthorityData),
		ClusterEndpoint:                 jsii.String(clusterInfo.ApiServerEndpoint),
		ClusterSecurityGroupId:          jsii.String(clusterInfo.ClusterSecurityGroupId),
		OpenIdConnectProvider:           awsiam.OpenIdConnectProvider_FromOpenIdConnectProviderArn(stack, jsii.String("idp"), jsii.String(clusterInfo.OidcIdpArn)),
		Vpc: awsec2.Vpc_FromLookup(stack, jsii.String("vpc"), &awsec2.VpcLookupOptions{
			IsDefault: jsii.Bool(false),
			Region:    jsii.String(clusterInfo.Region),
			VpcId:     jsii.String(clusterInfo.VpcId),
		}),
		KubectlRoleArn: jsii.String(clusterInfo.KubectlRoleArn),
	})

	// Construct CDK8s app.
	app := cdk8s.NewApp(nil)
	chart := cdk8s.NewChart(app, jsii.String("CDK8s-Chart"), nil)

	ns := k8s.NewKubeNamespace(chart, jsii.String("K8s-Namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String("test2"),
		},
	})

	cfgmap := cdk8splus21.NewConfigMap(chart, jsii.String("K8s-ConfigMap"), &cdk8splus21.ConfigMapProps{
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String("test2-cfgmap"),
			Namespace: jsii.String("test2"),
			Labels: &map[string]*string{
				"app": jsii.String("test2"),
			},
		},
		Data: &map[string]*string{
			"deploymentRegion": jsii.String("ap-northeast-1"),
			"tableName":        jsii.String("RESTBenchmark-ChatTable"),
			"gsiName":          jsii.String("ChatTableGSI"),
			"testName":         jsii.String("Test"),
			"myZZZ":            jsii.String("Hahaha"),
			"finally_finished": jsii.String("wahaha"),
		},
	})

	cfgmap.ApiObject().AddDependency(ns)

	cluster.AddCdk8sChart(jsii.String("k8s-add-chart"), chart, &awseks.KubernetesManifestOptions{})

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
