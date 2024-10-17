package fedramp

import (
	"context"
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ocmclient "github.com/openshift/osde2e-common/pkg/clients/ocm"
	awscloud "github.com/openshift/osde2e-common/pkg/clouds/aws"
	"github.com/openshift/osde2e-common/pkg/openshift/rosa"
)

var _ = Describe("Fedramp Rosa Cluster Creation", Ordered, Label("Fedramp"), func() {
	var (
		logger         = GinkgoLogr
		provider       *rosa.Provider
		clusterName    string
		clusterChannel string
		clusterVersion string
		privatelink    bool
		sts            bool
		clusterID      string
		reportDir      string
	)

	BeforeAll(func(ctx context.Context) {
		var err error

		// Initialize ROSA provider
		logger.Info("Initializing ROSA provider")
		provider, err = rosa.New(
			ctx,
			os.Getenv("OCM_TOKEN"),
			os.Getenv("FEDRAMP_CLIENT_ID"),
			os.Getenv("FEDRAMP_CLIENT_SECRET"),
			ocmclient.FedRampIntegration,
			logger,
			&awscloud.AWSCredentials{Profile: "", Region: ""},
		)
		Expect(err).NotTo(HaveOccurred(), "Failed to create ROSA provider")
		logger.Info("ROSA provider initialized successfully")

		// Create report directory in the pod. This is used by the junit reporter.
		reportDir = os.Getenv("REPORT_DIR")
		Expect(os.MkdirAll(reportDir, os.ModePerm)).ShouldNot(HaveOccurred(), "failed to create report directory")
		logger.Info(fmt.Sprintf("Report directory created successfully: %s", reportDir))

	})

	AfterAll(func(ctx context.Context) {
		deleteOptions := &rosa.DeleteClusterOptions{
			ClusterName:     clusterName,
			PrivateLink:     privatelink,
			STS:             sts,
			DeleteHostedVPC: true,
		}
		err := provider.DeleteCluster(ctx, deleteOptions)
		Expect(err).NotTo(HaveOccurred(), "Failed to delete ROSA cluster")
	})

	It("should successfully create a ROSA FedRamp Cluster", func(ctx context.Context) {
		var err error

		// Cluster setup variables
		// TODO: Change this to a random name
		clusterName = "osde2efr"
		logger.Info(fmt.Sprintf("Cluster name set to: %s", clusterName))

		clusterChannel = os.Getenv("CHANNEL")
		if clusterChannel == "" {
			clusterChannel = "stable"
		}
		logger.Info(fmt.Sprintf("Cluster channel set to: %s", clusterChannel))

		// Get the cluster version
		logger.Info("Fetching cluster version")
		clusterVersion = os.Getenv("CLUSTER_VERSION")
		if clusterVersion == "" {
			// Use provider.Versions to get the list of versions
			versions, err := provider.Versions(ctx, clusterChannel, false)
			Expect(err).NotTo(HaveOccurred(), "Failed to get ROSA versions")
			// Find the default version
			for _, version := range versions {
				if version.Default {
					clusterVersion = strings.TrimSpace(version.RawID)
					break
				}
			}
			Expect(clusterVersion).NotTo(BeEmpty(), "No default ROSA version found in channel group")
		}
		logger.Info(fmt.Sprintf("Cluster version fetched: %s", clusterVersion))

		privatelink = true
		sts = true
		logger.Info("Creating ROSA cluster with the following options:")
		logger.Info(fmt.Sprintf("PrivateLink: %t, STS: %t, Version: %s", privatelink, sts, clusterVersion))

		// Privatelink Cluster Option enables the VPC creation in the customer account.
		// Create ROSA cluster
		clusterID, err = provider.CreateCluster(
			ctx,
			&rosa.CreateClusterOptions{
				ClusterName:     clusterName,
				ChannelGroup:    clusterChannel,
				PrivateLink:     privatelink,
				STS:             sts,
				Version:         clusterVersion,
				InstallTimeout:  180,  // 3 hours since this includes the time to create the VPC and subnets
				SkipHealthCheck: true, // TODO: Remove this once we have a way to access a privatelink cluster
			},
		)
		Expect(err).NotTo(HaveOccurred(), "Failed to create ROSA cluster")
		logger.Info(fmt.Sprintf("Cluster created successfully! ClusterID: %s", clusterID))
	})
})
