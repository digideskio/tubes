package awsclient

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Config struct {
	Region                    string
	AccessKey                 string
	SecretKey                 string
	EndpointOverride          string
	CloudFormationWaitTimeout time.Duration
}

type ec2Client interface {
	DescribeImages(*ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error)
	DescribeSubnets(*ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error)
}

type cloudformationClient interface {
	DescribeStackResources(*cloudformation.DescribeStackResourcesInput) (*cloudformation.DescribeStackResourcesOutput, error)
	DescribeStacks(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error)
	CreateStack(*cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error)
	UpdateStack(*cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error)
	DeleteStack(*cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error)
}

type cloudFormationStatusPundit interface {
	IsHealthy(statusString string) bool
	IsComplete(statusString string) bool
}

type clock interface {
	Sleep(time.Duration)
}

type Client struct {
	EC2                        ec2Client
	CloudFormation             cloudformationClient
	CloudFormationStatusPundit cloudFormationStatusPundit
	Clock                      clock
	CloudFormationWaitTimeout  time.Duration
}

func New(c Config) *Client {
	credentials := credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, "")
	sdkConfig := &aws.Config{
		Credentials: credentials,
		Region:      aws.String(c.Region),
		Endpoint:    aws.String(c.EndpointOverride),
	}

	session := session.New(sdkConfig)

	if c.CloudFormationWaitTimeout == 0 {
		c.CloudFormationWaitTimeout = 5 * time.Minute
	}
	return &Client{
		EC2:                        ec2.New(session),
		CloudFormation:             cloudformation.New(session),
		CloudFormationStatusPundit: CloudFormationStatusPundit{},
		Clock: clockImpl{},
		CloudFormationWaitTimeout: c.CloudFormationWaitTimeout,
	}
}

type clockImpl struct{}

func (c clockImpl) Sleep(d time.Duration) { time.Sleep(d) }

// ARN represents an Amazon Resource Name
// http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html
type ARN struct {
	Partition string
	Service   string
	Region    string
	AccountID string
	Resource  string
}

// ParseARN parses an ARN string into its component fields
func (c Client) ParseARN(arn string) (ARN, error) {
	const numExpectedParts = 6
	parts := strings.SplitN(arn, ":", numExpectedParts)
	if len(parts) < numExpectedParts {
		return ARN{}, fmt.Errorf("malformed ARN %q", arn)
	}
	return ARN{
		Partition: parts[1],
		Service:   parts[2],
		Region:    parts[3],
		AccountID: parts[4],
		Resource:  parts[5],
	}, nil
}
