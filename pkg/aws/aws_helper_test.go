package aws

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestGetAccountNumber(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	stsClient := &STSMocker{}

	output := getAccountNumber(stsClient)
	g.Expect(output).ToNot(gomega.Equal(""))
}
