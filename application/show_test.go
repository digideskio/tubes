package application_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type erroringWriter struct{}

func (w *erroringWriter) Write(data []byte) (int, error) {
	return -3, errors.New("write failed")
}

var _ = Describe("Show", func() {
	It("should print the SSH key to the result writer", func() {
		configStore.Values[stackName+"/ssh-key"] = []byte("some pem block")

		Expect(app.Show(stackName)).To(Succeed())

		Expect(resultBuffer.Contents()).To(Equal([]byte("some pem block")))
	})

	Context("when the config store get errors", func() {
		It("should return the error", func() {
			configStore.Errors[stackName+"/ssh-key"] = errors.New("some error")
			Expect(app.Show(stackName)).To(MatchError("some error"))
		})
	})

	Context("when the writing the result", func() {
		It("should return the error", func() {
			app.ResultWriter = &erroringWriter{}
			Expect(app.Show(stackName)).To(MatchError("write failed"))
		})
	})
})
