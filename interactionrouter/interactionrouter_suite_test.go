package interactionrouter_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestInteractionrouter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Interactionrouter Suite")
}
