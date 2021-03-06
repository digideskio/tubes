package acceptance_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Creating and destroying real environments", func() {
	var (
		stackName  string
		envVars    map[string]string
		workingDir string
	)

	var start = func(args ...string) *gexec.Session {
		command := exec.Command(pathToCLI, args...)
		command.Env = []string{}
		if envVars != nil {
			for k, v := range envVars {
				command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
			}
		}
		command.Dir = workingDir
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		return session
	}

	BeforeEach(func() {
		stackName = fmt.Sprintf("tubes-acceptance-test-%x", rand.Int())
		var err error
		workingDir, err = ioutil.TempDir("", "tubes-acceptance-test")
		Expect(err).NotTo(HaveOccurred())

		envVars = getEnvironment()
	})

	It("should support basic environment manipulation", func() { // slow happy path
		const NormalTimeout = "10s"
		const StackChangeTimeout = "6m"

		By("booting a fresh environment", func() {
			session := start("-n", stackName, "up")

			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Creating keypair"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Looking for latest AWS NAT box AMI"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("ami-[a-f0-9]*"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Upserting base stack"))
			Eventually(session.Err, StackChangeTimeout).Should(gbytes.Say("Stack update complete"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Upserting Concourse stack"))
			Eventually(session.Err, StackChangeTimeout).Should(gbytes.Say("Stack update complete"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Finished"))
			Eventually(session, NormalTimeout).Should(gexec.Exit(0))
		})

		By("tearing down the environment", func() {
			session := start("-n", stackName, "down")

			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Deleting Concourse stack"))
			Eventually(session.Err, StackChangeTimeout).Should(gbytes.Say("Delete complete"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Deleting base stack"))
			Eventually(session.Err, StackChangeTimeout).Should(gbytes.Say("Delete complete"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Deleting keypair"))
			Eventually(session.Err, NormalTimeout).Should(gbytes.Say("Finished"))
			Eventually(session, NormalTimeout).Should(gexec.Exit(0))
		})
	})
})
