package application_test

import (
	"fmt"
	"log"
	"math/rand"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/rosenhouse/tubes/application"
	"github.com/rosenhouse/tubes/mocks"

	"testing"
)

func TestApplication(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Application Suite")
}

var _ = BeforeSuite(func() {
	rand.Seed(config.GinkgoConfig.RandomSeed)
})

var (
	awsClient *mocks.AWSClient

	app *application.Application

	stackName    string
	logBuffer    *gbytes.Buffer
	resultBuffer *gbytes.Buffer
	configStore  *mocks.ConfigStore
)

var _ = BeforeEach(func() {
	awsClient = &mocks.AWSClient{}
	configStore = mocks.NewConfigStore()

	logBuffer = gbytes.NewBuffer()
	resultBuffer = gbytes.NewBuffer()

	app = &application.Application{
		AWSClient:    awsClient,
		Logger:       log.New(logBuffer, "", 0),
		ResultWriter: resultBuffer,
		ConfigStore:  configStore,
	}

	stackName = fmt.Sprintf("some-stack-name-%x", rand.Int31())
})
