package itest

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestItest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration test for sabakan")
}
