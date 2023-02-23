package common

import (
	"os"
	"testing"

	"github.com/onsi/gomega"
)

func TestGetEnv(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	err := os.Setenv("a", "1")
	g.Expect(err).To(gomega.BeNil())
	err = os.Setenv("b", "2")
	g.Expect(err).To(gomega.BeNil())

	g.Expect(GetEnv("a", "_")).To(gomega.Equal("1"))
	g.Expect(GetEnv("b", "_")).To(gomega.Equal("2"))
	g.Expect(GetEnv("c", "_")).To(gomega.Equal("_"))
}
