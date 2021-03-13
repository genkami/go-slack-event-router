package reaction_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestReaction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Reaction Suite")
}
