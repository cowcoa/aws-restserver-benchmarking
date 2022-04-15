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

	// Import eks cluster.
	cluster := awseks.Cluster_FromClusterAttributes(stack, jsii.String("EKSCluster"), &awseks.ClusterAttributes{
		ClusterName:                     jsii.String(clusterInfo.ClusterName),
		ClusterCertificateAuthorityData: jsii.String(clusterInfo.CertificateAuthorityData),
		ClusterEndpoint:                 jsii.String(clusterInfo.ApiServerEndpoint),
		ClusterSecurityGroupId:          jsii.String(clusterInfo.ClusterSecurityGroupId),
		OpenIdConnectProvider:           awsiam.OpenIdConnectProvider_FromOpenIdConnectProviderArn(stack, jsii.String("idp"), jsii.String(clusterInfo.OidcIdpArn)),
		Vpc: awsec2.Vpc_FromLookup(stack, jsii.String("VPC"), &awsec2.VpcLookupOptions{
			IsDefault: jsii.Bool(false),
			Region:    jsii.String(clusterInfo.Region),
			VpcId:     jsii.String(clusterInfo.VpcId),
		}),
		KubectlRoleArn: jsii.String(clusterInfo.KubectlRoleArn),
	})

	// Construct CDK8s app.
	app := cdk8s.NewApp(nil)
	chart := cdk8s.NewChart(app, jsii.String("CDK8s-Chart"), nil)

	nsName := "restbenchmark"
	appName := "nodejs-express-ddb"
	saName := appName
	imageUri := *stack.Account() + ".dkr.ecr." + *stack.Region() + ".amazonaws.com/" + config.EcrRepoName(stack) + ":latest"
	appLabel := map[string]*string{
		"app": jsii.String(appName),
	}
	const servicePort = 8783
	const containerPort = 8387
	// const nodePort = 30037

	// Create app service account role.
	saRole := awsiam.NewRole(stack, jsii.String("SARole"), &awsiam.RoleProps{
		RoleName: jsii.String(*stack.StackName() + "-SARole"),
		AssumedBy: awsiam.NewWebIdentityPrincipal(cluster.OpenIdConnectProvider().OpenIdConnectProviderArn(), &map[string]interface{}{
			"StringEquals": awscdk.NewCfnJson(stack, jsii.String("CfnJson-SARole"), &awscdk.CfnJsonProps{
				Value: map[string]string{
					*cluster.OpenIdConnectProvider().OpenIdConnectProviderIssuer() + ":aud": "sts.amazonaws.com",
					*cluster.OpenIdConnectProvider().OpenIdConnectProviderIssuer() + ":sub": "system:serviceaccount:" + nsName + ":" + saName,
				},
			}),
		}),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonDynamoDBFullAccess")),
		},
	})

	// Create k8s namespace
	ns := k8s.NewKubeNamespace(chart, jsii.String("K8s-Namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String(nsName),
		},
	})

	// Create k8s service account
	sa := cdk8splus21.NewServiceAccount(chart, jsii.String("K8s-ServiceAccount"), &cdk8splus21.ServiceAccountProps{
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String(saName),
			Namespace: jsii.String(nsName),
			Labels:    &appLabel,
			Annotations: &map[string]*string{
				"eks.amazonaws.com/role-arn": saRole.RoleArn(),
			},
		},
	})
	sa.ApiObject().AddDependency(ns)

	// Create k8s config map
	cfgMap := cdk8splus21.NewConfigMap(chart, jsii.String("K8s-ConfigMap"), &cdk8splus21.ConfigMapProps{
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String(appName + "-cfgmap"),
			Namespace: jsii.String(nsName),
			Labels:    &appLabel,
		},
		Data: &map[string]*string{
			"deploymentRegion": jsii.String(clusterInfo.Region),
			"tableName":        jsii.String(config.TableName(stack)),
			"gsiName":          jsii.String(config.GsiName(stack)),
		},
	})
	cfgMap.ApiObject().AddDependency(sa)

	// Create k8s deployment
	deploy := cdk8splus21.NewDeployment(chart, jsii.String("K8s-Deployment"), &cdk8splus21.DeploymentProps{
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String(appName + "-deployment"),
			Namespace: jsii.String(nsName),
			Labels:    &appLabel,
		},
		Containers: &[]*cdk8splus21.ContainerProps{
			{
				Name:  jsii.String(appName),
				Image: jsii.String(imageUri),
				Port:  jsii.Number(containerPort),
				Env: &map[string]cdk8splus21.EnvValue{
					"AWS_REGION":     cdk8splus21.EnvValue_FromConfigMap(cfgMap, jsii.String("deploymentRegion"), nil),
					"DDB_TABLE_NAME": cdk8splus21.EnvValue_FromConfigMap(cfgMap, jsii.String("tableName"), nil),
					"DDB_GSI_NAME":   cdk8splus21.EnvValue_FromConfigMap(cfgMap, jsii.String("gsiName"), nil),
				},
				/*
					Liveness: cdk8splus21.Probe_FromCommand(&[]*string{
						jsii.String("/bin/sh"),
						jsii.String("-c"),
						jsii.String("nc -z localhost " + strconv.Itoa(containerPort)),
					}, &cdk8splus21.CommandProbeOptions{
						InitialDelaySeconds: cdk8s.Duration_Millis(jsii.Number(3000)),
						PeriodSeconds:       cdk8s.Duration_Millis(jsii.Number(5000)),
					}),
				*/
				Resources: &cdk8splus21.Resources{
					Cpu: &cdk8splus21.CpuResources{
						Limit:   cdk8splus21.Cpu_Millis(jsii.Number(1000)),
						Request: cdk8splus21.Cpu_Millis(jsii.Number(500)),
					},
					Memory: &cdk8splus21.MemoryResources{
						Limit:   cdk8s.Size_Mebibytes(jsii.Number(500)),
						Request: cdk8s.Size_Mebibytes(jsii.Number(250)),
					},
				},
			},
		},
		ServiceAccount: cdk8splus21.ServiceAccount_FromServiceAccountName(jsii.String(*sa.Name())),
		PodMetadata: &cdk8s.ApiObjectMetadata{
			Namespace: jsii.String(nsName),
			Labels:    &appLabel,
		},
		DefaultSelector: jsii.Bool(true),
		Replicas:        jsii.Number(3),
	})
	deploy.ApiObject().AddDependency(cfgMap)

	// Create k8s service
	svc := cdk8splus21.NewService(chart, jsii.String("K8s-Service"), &cdk8splus21.ServiceProps{
		Metadata: &cdk8s.ApiObjectMetadata{
			Name:      jsii.String(appName + "-service"),
			Namespace: jsii.String(nsName),
			Labels:    &appLabel,
		},
		Ports: &[]*cdk8splus21.ServicePort{
			{
				Port:       jsii.Number(servicePort),
				TargetPort: jsii.Number(containerPort),
			},
		},
		Type: cdk8splus21.ServiceType_NODE_PORT,
	})
	svc.AddSelector(jsii.String("app"), jsii.String(appName))
	svc.ApiObject().AddDependency(deploy)

	// Create k8s ingress
	ingress := k8s.NewKubeIngress(chart, jsii.String("K8s-Ingress"), &k8s.KubeIngressProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String(appName + "-ingress"),
			Namespace: &nsName,
			Labels:    &appLabel,
			Annotations: &map[string]*string{
				// All lowercase, no longer than 32 characters.
				"alb.ingress.kubernetes.io/load-balancer-name": jsii.String("eks-alb-ingress"),
				// Ingress Core Settings
				"alb.ingress.kubernetes.io/scheme": jsii.String("internet-facing"),
				// Health Check Settings
				"alb.ingress.kubernetes.io/healthcheck-protocol":         jsii.String("HTTP"),
				"alb.ingress.kubernetes.io/healthcheck-port":             jsii.String("traffic-port"),
				"alb.ingress.kubernetes.io/healthcheck-path":             jsii.String("/health"),
				"alb.ingress.kubernetes.io/healthcheck-interval-seconds": jsii.String("15"),
				"alb.ingress.kubernetes.io/healthcheck-timeout-seconds":  jsii.String("5"),
				"alb.ingress.kubernetes.io/success-codes":                jsii.String("200"),
				"alb.ingress.kubernetes.io/healthy-threshold-count":      jsii.String("2"),
				"alb.ingress.kubernetes.io/unhealthy-threshold-count":    jsii.String("2"),
			},
		},
		Spec: &k8s.IngressSpec{
			IngressClassName: jsii.String("alb"),
			DefaultBackend: &k8s.IngressBackend{
				Service: &k8s.IngressServiceBackend{
					Name: svc.Name(),
					Port: &k8s.ServiceBackendPort{
						Number: jsii.Number(servicePort),
					},
				},
			},
		},
	})
	ingress.AddDependency(svc)

	// Add chart to cluster.
	cluster.AddCdk8sChart(jsii.String("EKSCDK8sChart"), chart, nil)

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
