package application_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/rosenhouse/tubes/application"
	"github.com/rosenhouse/tubes/lib/awsclient"
)

var _ = Describe("Up", func() {
	BeforeEach(func() {
		awsClient.GetLatestNATBoxAMIIDCall.Returns.AMIID = "some-nat-box-ami-id"
	})

	It("should boot the base stack using the latest NAT ID", func() {
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(logBuffer).To(gbytes.Say("Creating keypair"))
		Expect(logBuffer).To(gbytes.Say("Looking for latest AWS NAT box AMI..."))
		Expect(logBuffer).To(gbytes.Say("Latest NAT box AMI is \"some-nat-box-ami-id\""))
		Expect(logBuffer).To(gbytes.Say("Upserting stack..."))
		Expect(logBuffer).To(gbytes.Say("Stack update complete"))
		Expect(logBuffer).To(gbytes.Say("Finished"))

		Expect(awsClient.UpsertStackCall.Receives.StackName).To(Equal(stackName))
		Expect(awsClient.UpsertStackCall.Receives.Template).To(Equal(awsclient.BaseStackTemplate.String()))
		Expect(awsClient.UpsertStackCall.Receives.Parameters).To(Equal(map[string]string{
			"NATInstanceAMI": "some-nat-box-ami-id",
			"KeyName":        stackName,
		}))
	})

	It("should wait for the stack to boot", func() {
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(awsClient.WaitForStackCall.Receives.StackName).To(Equal(stackName))
		Expect(awsClient.WaitForStackCall.Receives.Pundit).To(Equal(awsclient.CloudFormationUpsertPundit{}))
	})

	It("should create a new ssh keypair", func() {
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(awsClient.CreateKeyPairCall.Receives.StackName).To(Equal(stackName))
	})

	It("should store the ssh keypair in the config store", func() {
		awsClient.CreateKeyPairCall.Returns.KeyPair = "some pem bytes"
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(configStore.Values).To(HaveKeyWithValue(
			"ssh-key",
			[]byte("some pem bytes")))
	})

	It("should get the base stack resources", func() {
		Expect(app.Boot(stackName)).To(Succeed())
		Expect(awsClient.GetBaseStackResourcesCall.Receives.StackName).To(Equal(stackName))
	})

	It("should provide the stack resources to the manifest builder", func() {
		resources := awsclient.BaseStackResources{
			AccountID: "ping pong",
		}
		awsClient.GetBaseStackResourcesCall.Returns.Resources = resources
		Expect(app.Boot(stackName)).To(Succeed())

		Expect(manifestBuilder.BuildCall.Receives.StackName).To(Equal(stackName))
		Expect(manifestBuilder.BuildCall.Receives.Resources).To(Equal(resources))
	})

	It("should store the manifest", func() {
		manifestBuilder.BuildCall.Returns.ManifestYAML = []byte("some-manifest-bytes")

		Expect(app.Boot(stackName)).To(Succeed())

		Expect(configStore.Values).To(HaveKeyWithValue(
			"director.yml",
			[]byte("some-manifest-bytes"),
		))
	})

	Context("when the stackName contains invalid characters", func() {
		It("should immediately error", func() {
			Expect(app.Boot("invalid_name")).To(MatchError(fmt.Sprintf("invalid name: must match pattern %s", application.StackNamePattern)))
			Expect(logBuffer.Contents()).To(BeEmpty())
		})
	})

	Context("when getting the latest NAT AMI errors", func() {
		It("should immediately return the error", func() {
			awsClient.GetLatestNATBoxAMIIDCall.Returns.Error = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))
			Expect(awsClient.UpsertStackCall.Receives.StackName).To(BeEmpty())
			Expect(awsClient.WaitForStackCall.Receives.StackName).To(BeEmpty())
		})
	})

	Context("when creating a keypair fails", func() {
		It("should immediately return the error", func() {
			awsClient.CreateKeyPairCall.Returns.Error = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))
			Expect(awsClient.UpsertStackCall.Receives.StackName).To(BeEmpty())
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Looking for latest AWS NAT box AMI"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Finished"))
		})
	})

	Context("when storing the ssh key fails", func() {
		It("should return an error", func() {
			configStore.Errors["ssh-key"] = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Upserting stack..."))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Finished"))
		})
	})

	Context("when upserting the stack errors", func() {
		It("should immediately return the error", func() {
			awsClient.UpsertStackCall.Returns.Error = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))
			Expect(awsClient.WaitForStackCall.Receives.StackName).To(BeEmpty())
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Stack update complete"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Finished"))
		})
	})

	Context("when waiting for the stack errors", func() {
		It("should return the error", func() {
			awsClient.WaitForStackCall.Returns.Error = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))

			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Stack update complete"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Finished"))
		})
	})

	Context("when getting the base stack resources fails", func() {
		It("should return the error", func() {
			awsClient.GetBaseStackResourcesCall.Returns.Error = errors.New("boom")

			Expect(app.Boot(stackName)).To(MatchError("boom"))
		})
	})

	Context("when building the manifest yaml errors", func() {
		It("should return the error", func() {
			manifestBuilder.BuildCall.Returns.Error = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))
		})
	})

	Context("when storing the manifest yaml fails", func() {
		It("should return an error", func() {
			configStore.Errors["director.yml"] = errors.New("some error")

			Expect(app.Boot(stackName)).To(MatchError("some error"))
			Expect(logBuffer.Contents()).NotTo(ContainSubstring("Finished"))
		})
	})

})
