package sdn_migration_test

import (
	"context"
	"os"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e-common/pkg/clients/ocm"
	"github.com/openshift/osde2e-common/pkg/clouds/aws"
	"github.com/openshift/osde2e-common/pkg/openshift/rosa"
)

/*

- before all create clients
- test cluster creation
- test cluster upgrade
- test cluster migration
	- test apply manifests

*/

var _ = Describe("SDN migration", ginkgo.Ordered, func() {
	const clusterName = "creed-sdn-ovn-1"

	var rp *rosa.Provider
	BeforeAll(func(ctx context.Context) {
		var err error
		rp, err = rosa.New(
			ctx,
			os.Getenv("OCM_TOKEN"),
			"",
			"",
			ocm.Stage,
			ginkgo.GinkgoLogr,
			&aws.AWSCredentials{
				Profile:         "",
				Region:          "us-east-1",
				SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
				AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
			},
		)
		Expect(err).Should(BeNil())
	})

	AfterAll(func(ctx context.Context) {
		err := rp.DeleteCluster(ctx, &rosa.DeleteClusterOptions{})
		Expect(err).Should(BeNil())
	})

	// rosa create cluster -y --sts --mode auto --cluster-name creed-sdn-ovn-141414-1 --region us-east-1 --version 4.14.14 --channel-group stable --compute-machine-type m5.xlarge --multi-az --enable-autoscaling --min-replicas 3 --max-replicas 24 --etcd-encryption --network-type OpenShiftSDN
	It("create cluster", func(ctx context.Context) {
		opts := &rosa.CreateClusterOptions{
			ClusterName:                  clusterName,
			Version:                      "4.14.14",
			UseDefaultAccountRolesPrefix: true,
			STS:                          true,
			Mode:                         "auto",
			ChannelGroup:                 "stable",
			ComputeMachineType:           "m5.xlarge",
			MinReplicas:                  3,
			MaxReplicas:                  24,
			MultiAZ:                      true,
			EnableAutoscaling:            true,
			ETCDEncryption:               true,
			NetworkType:                  "OpenShiftSDN",
		}

		if os.Getenv("CLUSTER_ID") != "" {
			ginkgo.Skip("")
		}

		id, err := rp.CreateCluster(ctx, opts)
		Expect(err).Should(BeNil())

		_ = id
	})

	It("upgrade cluster", func(ctx context.Context) {
	})
})
