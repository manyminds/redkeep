package redkeep_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var testConfiguration = "example-configuration.json"

func TestRedkeep(t *testing.T) {
	if env := os.Getenv("TEST_CONFIGURATION"); env != "" {
		testConfiguration = env
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Redkeep Suite")
}
