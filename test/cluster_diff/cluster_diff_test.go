package cluster_diff_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/openshift/osde2e-common/pkg/clients/ocm"
	rosaprovider "github.com/openshift/osde2e-common/pkg/openshift/rosa"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

var _ = Describe("Gap Analysis - Cluster Diff", Ordered, func() {
	// provision two clusters: one Y-1, one Y
	// run gap analysis script
	// fail if there is a difference

	var (
		rosa        *rosaprovider.Provider
		name        = envconf.RandomName("osde2e-diff", 13)
		stableName  = name + "-1"
		nightlyName = name + "-2"

		nightlyVersion = os.Getenv("RELEASE_IMAGE_LATEST")
	)

	BeforeAll(func(ctx context.Context) {
		log.SetLogger(GinkgoLogr)
		klog.SetOutput(GinkgoWriter)

		var err error
		rosa, err = rosaprovider.New(ctx, os.Getenv("OCM_TOKEN"), ocm.Stage, klog.NewKlogr())
		Expect(err).ShouldNot(HaveOccurred(), "failed to create rosa provider")

		if nightlyVersion == "" {
			nightlyVersions, err := rosa.Versions(ctx, "nightly", false)
			Expect(err).ShouldNot(HaveOccurred(), "unable to get cluster versions")

			nightlyVersion = nightlyVersions[0].RawID
		}

		nightlySemver, err := semver.NewVersion(nightlyVersion)
		Expect(err).ShouldNot(HaveOccurred(), "unable to get cluster versions")

		yMinusOneVersion := fmt.Sprintf("%d.%d", nightlySemver.Major(), nightlySemver.Minor()-1)

		// >= 4.12 < 4.13
		constraint := fmt.Sprintf(">= %s < %d.%d", yMinusOneVersion, nightlySemver.Major(), nightlySemver.Minor())

		// find versions
		stableVersions, err := rosa.Versions(ctx, "stable", false, constraint)
		Expect(err).ShouldNot(HaveOccurred(), "unable to get cluster versions")

		stableVersion := stableVersions[0].RawID

		eg, ctx := errgroup.WithContext(ctx)

		provision := func(name, version, channelGroup string) error {
			_, err := rosa.CreateCluster(ctx, &rosaprovider.CreateClusterOptions{
				ClusterName:  name,
				Version:      version,
				ChannelGroup: channelGroup,
				STS:          true,
			})
			if err != nil {
				return fmt.Errorf("unable to provision cluster %s: %w", name, err)
			}
			return nil
		}

		eg.Go(func() error { return provision(stableName, stableVersion, "stable") })
		eg.Go(func() error { return provision(nightlyName, nightlyVersion, "nightly") })

		Expect(eg.Wait()).Should(Succeed(), "failed to provision clusters")
	})

	AfterAll(func(ctx context.Context) {
		eg, ctx := errgroup.WithContext(ctx)

		destroy := func(name string) error {
			return rosa.DeleteCluster(ctx, &rosaprovider.DeleteClusterOptions{
				ClusterName: name,
				STS:         true,
			})
		}

		eg.Go(func() error { return destroy(stableName) })
		eg.Go(func() error { return destroy(nightlyName) })

		Expect(eg.Wait()).Should(Succeed(), "failed to destroy clusters")
	})

	It("should succeed", func(ctx context.Context) {
		cmd := exec.Command("./cluster-diff.sh", stableName, nightlyName)
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred(), "unable to start gexec session")
		Eventually(session.Out).WithTimeout(3 * time.Minute).Should(gbytes.Say("Gap analysis done!"))
	})
})
