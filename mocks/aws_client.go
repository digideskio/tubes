package mocks

import "github.com/rosenhouse/tubes/lib/awsclient"

type AWSClient struct {
	GetLatestNATBoxAMIIDCall struct {
		Returns struct {
			AMIID string
			Error error
		}
	}
	UpsertStackCall struct {
		Receives struct {
			StackName  string
			Template   string
			Parameters map[string]string
		}
		Returns struct {
			Error error
		}
	}
	DeleteStackCall struct {
		Receives struct {
			StackName string
		}
		Returns struct {
			Error error
		}
	}
	WaitForStackCall struct {
		Receives struct {
			StackName string
			Pundit    awsclient.CloudFormationStatusPundit
		}
		Returns struct {
			Error error
		}
	}
}

func (c *AWSClient) GetLatestNATBoxAMIID() (string, error) {
	return c.GetLatestNATBoxAMIIDCall.Returns.AMIID, c.GetLatestNATBoxAMIIDCall.Returns.Error
}

func (c *AWSClient) UpsertStack(stackName string, template string, parameters map[string]string) error {
	c.UpsertStackCall.Receives.StackName = stackName
	c.UpsertStackCall.Receives.Template = template
	c.UpsertStackCall.Receives.Parameters = parameters
	return c.UpsertStackCall.Returns.Error
}

func (c *AWSClient) WaitForStack(stackName string, pundit awsclient.CloudFormationStatusPundit) error {
	c.WaitForStackCall.Receives.StackName = stackName
	c.WaitForStackCall.Receives.Pundit = pundit
	return c.WaitForStackCall.Returns.Error
}

func (c *AWSClient) DeleteStack(stackName string) error {
	c.DeleteStackCall.Receives.StackName = stackName
	return c.DeleteStackCall.Returns.Error
}
