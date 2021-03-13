package urlverification_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUrlverification(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Urlverification Suite")
}
