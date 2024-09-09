package fedramp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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
			clusterVersion, err = getClusterVersion(ctx, provider, clusterChannel)
			Expect(err).NotTo(HaveOccurred())
		}
		logger.Info(fmt.Sprintf("Cluster version fetched: %s", clusterVersion))

		privatelink = true
		sts = true
		logger.Info("Creating ROSA cluster with the following options:")
		logger.Info(fmt.Sprintf("PrivateLink: %t, STS: %t, Version: %s", privatelink, sts, clusterVersion))

		// Create ROSA cluster
		clusterID, err = provider.CreateCluster(
			ctx,
			&rosa.CreateClusterOptions{
				ClusterName:     clusterName,
				ChannelGroup:    clusterChannel,
				PrivateLink:     privatelink,
				STS:             sts,
				Version:         clusterVersion,
				SkipHealthCheck: true, // TODO: Remove this once we have a way to access a privatelink cluster
			},
		)
		Expect(err).NotTo(HaveOccurred(), "Failed to create ROSA cluster")
		logger.Info(fmt.Sprintf("Cluster created successfully! ClusterID: %s", clusterID))
	})
})

// getClusterVersion gets the default ROSA version for the given channel group
// This is a hacky way to get the default ROSA version, but it works for now
// TODO: Find a better way to get the default ROSA version
// Latest rosa version is not available in fedramp rosa as they lag behind commercial.
// Note the Provider isn't being used here, this function could be moved to osde2e-common after a review.
func getClusterVersion(ctx context.Context, provider *rosa.Provider, clusterChannel string) (string, error) {
	// Construct the command to list ROSA versions in JSON format
	cmd := exec.CommandContext(ctx, "rosa", "list", "versions", "--channel-group", clusterChannel, "--output", "json")

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to list ROSA versions: %w\n%s", err, stderrBuf.String())
	}

	type VersionInfo struct {
		RawID   string `json:"raw_id"`
		Default bool   `json:"default"`
	}

	var versions []VersionInfo
	err = json.Unmarshal(stdoutBuf.Bytes(), &versions)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON output: %w", err)
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no ROSA versions found")
	}

	// Find the version where Default is true
	for _, version := range versions {
		if version.Default {
			return strings.TrimSpace(version.RawID), nil
		}
	}

	return "", fmt.Errorf("no default ROSA version found")
}
