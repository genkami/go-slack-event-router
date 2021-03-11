package eventrouter_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGoSlackEventRouter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoSlackEventRouter Suite")
}
