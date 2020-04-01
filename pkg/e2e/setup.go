package e2e

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/events"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/metadata"
	"github.com/openshift/osde2e/pkg/common/osd"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/state"
)

// Check if the test should run
var _ = ginkgo.BeforeEach(func() {
	testText := ginkgo.CurrentGinkgoTestDescription().TestText
	testContext := strings.TrimSpace(strings.TrimSuffix(ginkgo.CurrentGinkgoTestDescription().FullTestText, testText))

	shouldRun := false
	for _, testToRun := range config.Instance.Tests.TestsToRun {
		if strings.HasPrefix(testContext, testToRun) {
			shouldRun = true
			break
		}
	}

	if !shouldRun {
		ginkgo.Skip(fmt.Sprintf("test %s will not be run as its context (%s) is not specified as part of the tests to run", ginkgo.CurrentGinkgoTestDescription().FullTestText, testContext))
	}
})

// Setup cluster before testing begins.
var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	defer ginkgo.GinkgoRecover()
	cfg := config.Instance
	state := state.Instance

	err := setupCluster()
	events.HandleErrorWithEvents(err, events.InstallSuccessful, events.InstallFailed).ShouldNot(HaveOccurred(), "failed to setup cluster for testing")
	if err != nil {
		return []byte{}
	}

	if len(cfg.Addons.IDs) > 0 {
		err = installAddons()
		events.HandleErrorWithEvents(err, events.InstallAddonsSuccessful, events.InstallAddonsFailed).ShouldNot(HaveOccurred(), "failed while installing addons")
		if err != nil {
			return []byte{}
		}
	}

	if len(state.Kubeconfig.Contents) == 0 {
		// Give the cluster some breathing room.
		log.Println("OSD cluster installed. Sleeping for 600s.")
		time.Sleep(600 * time.Second)
	} else {
		log.Printf("No kubeconfig contents found, but there should be some by now.")
	}

	return []byte{}
}, func(data []byte) {
	// only needs to run once
})

// After all tests run, clean up within the Ginkgo context
var _ = ginkgo.SynchronizedAfterSuite(func() {
	h := helper.NewOutsideGinkgo()
	// Run Must Gather
	func() {
		defer ginkgo.GinkgoRecover()

		log.Print("Running Must Gather...")
		mustGatherTimeoutInSeconds := 900
		h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")
		r := h.Runner(fmt.Sprintf("oc adm must-gather --dest-dir=%v", runner.DefaultRunner.OutputDir))
		r.Name = "must-gather"
		r.Tarball = true
		stopCh := make(chan struct{})
		err := r.Run(mustGatherTimeoutInSeconds, stopCh)
		Expect(err).NotTo(HaveOccurred())
		gatherResults, err := r.RetrieveResults()
		Expect(err).NotTo(HaveOccurred())
		h.WriteResults(gatherResults)

	}()

	// Run Cluster State
	func() {
		defer ginkgo.GinkgoRecover()

		log.Print("Gathering Cluster State...")
		clusterState := h.GetClusterState()
		stateResults := make(map[string][]byte, len(clusterState))
		for resource, list := range clusterState {
			data, err := json.MarshalIndent(list, "", "    ")
			Expect(err).NotTo(HaveOccurred())

			var gbuf bytes.Buffer
			zw := gzip.NewWriter(&gbuf)
			_, err = zw.Write(data)
			Expect(err).NotTo(HaveOccurred())

			err = zw.Close()
			Expect(err).NotTo(HaveOccurred())

			// include gzip in filename to mark compressed data
			filename := fmt.Sprintf("%s-%s-%s.json.gzip", resource.Group, resource.Version, resource.Resource)
			stateResults[filename] = gbuf.Bytes()
		}

		// write results to disk
		h.WriteResults(stateResults)
	}()

	// Get state from OCM
	func() {
		var OSD *osd.OSD
		var err error
		cfg := config.Instance
		if len(state.Instance.Cluster.ID) > 0 {
			if OSD, err = osd.New(cfg.OCM.Token, cfg.OCM.Env, cfg.OCM.Debug); err != nil {
				log.Printf("Could not setup OSD: %v", err)
			}

			cluster, err := OSD.GetCluster(state.Instance.Cluster.ID)
			if err != nil {
				log.Printf("Could not query OCM for cluster %s: %s", state.Instance.Cluster.ID, err.Error())
			} else {
				flavorName, _ := cluster.Flavour().GetName()
				log.Printf("Cluster addons: %v", cluster.Addons().Slice())
				log.Printf("Cluster cloud provider: %v", cluster.CloudProvider().DisplayName())
				log.Printf("Cluster expiration: %v", cluster.ExpirationTimestamp())
				log.Printf("Cluster flavor: %s", flavorName)
				log.Printf("Cluster state: %v", cluster.State())
			}

		} else {
			log.Print("No cluster ID set. Skipping OCM Queries.")
		}
	}()

	log.Printf("Getting logs for cluster '%s'...", state.Instance.Cluster.ID)
	getLogs()
}, func() {})

// Collect logs after each test
var _ = ginkgo.JustAfterEach(getLogs)

func getLogs() {
	defer ginkgo.GinkgoRecover()
	state := state.Instance

	if OSD == nil {
		log.Println("OSD was not configured. Skipping log collection...")
	} else if state.Cluster.ID == "" {
		log.Println("CLUSTER_ID is not set, likely due to a setup failure. Skipping log collection...")
	} else {
		logs, err := OSD.FullLogs(state.Cluster.ID)
		Expect(err).NotTo(HaveOccurred(), "failed to collect cluster logs")
		writeLogs(logs)
	}
}

// setupCluster brings up a cluster, waits for it to be ready, then returns it's name.
func setupCluster() (err error) {
	cfg := config.Instance
	state := state.Instance

	// if TEST_KUBECONFIG has been set, skip configuring OCM
	if len(state.Kubeconfig.Contents) > 0 || len(cfg.Kubeconfig.Path) > 0 {
		return useKubeconfig()
	}

	// create a new cluster if no ID is specified
	if state.Cluster.ID == "" {
		if state.Cluster.Name == "" {
			state.Cluster.Name = clusterName()
		}

		if state.Cluster.ID, err = OSD.LaunchCluster(); err != nil {
			return fmt.Errorf("could not launch cluster: %v", err)
		}
	} else {
		log.Printf("CLUSTER_ID of '%s' was provided, skipping cluster creation and using it instead", state.Cluster.ID)

		cluster, err := OSD.GetCluster(state.Cluster.ID)
		if err != nil {
			return fmt.Errorf("could not retrieve cluster information from OCM: %v", err)
		}

		state.Cluster.Name = cluster.Name()
		log.Printf("CLUSTER_NAME set to %s from OCM.", state.Cluster.Name)

		state.Cluster.Version = cluster.Version().ID()
		log.Printf("CLUSTER_VERSION set to %s from OCM.", state.Cluster.Version)
	}

	metadata.Instance.SetClusterName(state.Cluster.Name)
	metadata.Instance.SetClusterID(state.Cluster.ID)

	if err = OSD.WaitForClusterReady(); err != nil {
		return fmt.Errorf("failed waiting for cluster ready: %v", err)
	}

	if state.Kubeconfig.Contents, err = OSD.ClusterKubeconfig(state.Cluster.ID); err != nil {
		return fmt.Errorf("could not get kubeconfig for cluster: %v", err)
	}

	return nil
}

// installAddons installs addons onto the cluster
func installAddons() (err error) {
	num, err := OSD.InstallAddons(config.Instance.Addons.IDs)
	if err != nil {
		return fmt.Errorf("could not install addons: %s", err.Error())
	}
	if num > 0 {
		if err = OSD.WaitForClusterReady(); err != nil {
			return fmt.Errorf("failed waiting for cluster ready: %v", err)
		}
	}

	return nil
}

// useKubeconfig reads the path provided for a TEST_KUBECONFIG and uses it for testing.
func useKubeconfig() (err error) {
	cfg := config.Instance
	state := state.Instance

	_, err = clientcmd.RESTConfigFromKubeConfig(state.Kubeconfig.Contents)
	if err != nil {
		log.Println("Not an existing Kubeconfig, attempting to read file instead...")
	} else {
		log.Println("Existing valid kubeconfig!")
		return nil
	}

	state.Kubeconfig.Contents, err = ioutil.ReadFile(cfg.Kubeconfig.Path)
	if err != nil {
		return fmt.Errorf("failed reading '%s' which has been set as the TEST_KUBECONFIG: %v", cfg.Kubeconfig.Path, err)
	}
	log.Printf("Using a set TEST_KUBECONFIG of '%s' for Origin API calls.", cfg.Kubeconfig.Path)
	return nil
}

// cluster name format must be short enough to support all versions
func clusterName() string {
	vers := strings.TrimPrefix(state.Instance.Cluster.Version, osd.VersionPrefix)
	safeVersion := strings.Replace(vers, ".", "-", -1)
	return "ci-cluster-" + safeVersion + "-" + config.Instance.Suffix
}

func writeLogs(m map[string][]byte) {
	for k, v := range m {
		name := k + "-log.txt"
		filePath := filepath.Join(config.Instance.ReportDir, name)
		err := ioutil.WriteFile(filePath, v, os.ModePerm)
		Expect(err).NotTo(HaveOccurred(), "failed to write log '%s'", filePath)
	}
}
