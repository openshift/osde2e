package sdn_migration_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSdnMigration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SdnMigration Suite")
}
