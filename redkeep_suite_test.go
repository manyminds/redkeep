package redkeep_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRedkeep(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Redkeep Suite")
}
