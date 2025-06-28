package executioner_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestExecutioner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Executioner Suite")
}
