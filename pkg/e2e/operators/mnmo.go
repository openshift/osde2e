package operators

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ocmTypes "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	machinev1 "github.com/openshift/machine-api-operator/pkg/apis/machine/v1beta1"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var mnmoOperatorTestName string = "[Suite: e2e] [OSD] Managed Node Metadata Operator"

const (
	ocmMachinePoolName                       = "osde2e-test"
	MaxIndividualTestTimeout   time.Duration = 30 * time.Second
	MaxMachinePoolInitWaitTime time.Duration = 5 * time.Minute
)

func init() {
	alert.RegisterGinkgoAlert(mnmoOperatorTestName, "SD-SREP", "@sre-platform-team-v1alpha1", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

func ensureMachinePoolExists(clusterClient *ocmTypes.ClusterClient, h *helper.H) string {
	// Check to see if the machine pool exists before creating a new one.
	machinePoolListResp, err := clusterClient.MachinePools().List().Send()
	Expect(err).NotTo(HaveOccurred())

	if machinePoolListResp.Size() == 0 {
		// We're creating a whole new MachinePoolBuilder here because you can't pass
		// instanceType during an UPDATE call and if we build this in the BeforeEach
		// because it's a pointer it will update the stored Builder and the first run
		// of each suite of tests where it's not building a machinepool will fail
		machinePool, err := ocmTypes.NewMachinePool().
			ID(ocmMachinePoolName).
			InstanceType("m5.xlarge").
			Replicas(3).
			Build()
		Expect(err).NotTo(HaveOccurred())

		createPlainMachinePool(clusterClient, machinePool)
	}
	machineSet, err := waitForMachinePoolNodesReady(MaxMachinePoolInitWaitTime, h)
	Expect(err).NotTo(HaveOccurred())

	return machineSet.Name
}

func createPlainMachinePool(clusterClient *ocmTypes.ClusterClient, machinePool *ocmTypes.MachinePool) {
	_, err := clusterClient.MachinePools().Add().Body(machinePool).Send()
	Expect(err).NotTo(HaveOccurred())
}

func deleteMachinePool(clusterClient *ocmTypes.ClusterClient) {
	_, err := clusterClient.MachinePools().MachinePool(ocmMachinePoolName).Delete().Send()
	Expect(err).NotTo(HaveOccurred())
}

// This function ensures that the MachinePool Nodes are Ready and then unblocks.  Returns an error if it times out.
func waitForMachinePoolNodesReady(timeout time.Duration, h *helper.H) (*machinev1.MachineSet, error) {
	for t := 0 * time.Second; t < timeout; t = t + 2*time.Second {
		time.Sleep(2 * time.Second)
		machineSet, err := getMachineSet(h)
		if err != nil {
			// we don't really care what the error is here, we specifically expect some 404's here the first few calls to this.
			continue
		}
		if machineSet.Status.ReadyReplicas < 1 {
			// we don't have at least one ready node
			continue
		}
		// if we reach this point, we have at least one
		// ready node.
		return machineSet, nil
	}
	return &machinev1.MachineSet{}, fmt.Errorf("Timed Out waiting for MachinePool Nodes to become Ready")
}

func getMachineSet(h *helper.H) (*machinev1.MachineSet, error) {
	labelSelector := fmt.Sprintf("%s=%s", "hive.openshift.io/machine-pool", ocmMachinePoolName)
	machineSetList, err := h.Machine().MachineV1beta1().MachineSets("openshift-machine-api").List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}
	if len(machineSetList.Items) < 1 {
		return nil, fmt.Errorf("Items list is empty")
	}
	return &machineSetList.Items[0], err
}

func getNodesForMachineSet(h *helper.H, machineSetName string) ([]*corev1.Node, error) {
	nodeList := []*corev1.Node{}
	labelSelector := fmt.Sprintf("%s=%s", "machine.openshift.io/cluster-api-machineset", machineSetName)

	retry := true

	for retry {
		machineList, err := h.Machine().MachineV1beta1().Machines("openshift-machine-api").List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			return nodeList, err
		}

		for i := range machineList.Items {
			// prevent nil pointer panics
			if machineList.Items[i].Status.NodeRef == nil {
				continue
			}
			nodeName := machineList.Items[i].Status.NodeRef.Name
			node, err := h.Kube().CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
			if err != nil {
				return nodeList, err
			}
			nodeList = append(nodeList, node)
		}
		return nodeList, nil
	}
	return nodeList, nil
}

// MNMOTest is a struct of a test definition
type MNMOTest struct {
	interval time.Duration
	timeout  time.Duration

	// test is a function that runs a test condition.
	// this test will run in a loop in the `waitFor` function
	// returning a non-nil error will quickly exit the test function failing the upstream test.
	// returning a false, nil combination will retry the test
	// returning a true, nil combination will exit the test successfully
	test func([]*corev1.Node) (bool, error)
}

func waitFor(test MNMOTest, h *helper.H, machineSetName string) error {
	var nodeListError error
	for t := 0 * time.Second; t < test.timeout; t = t + test.interval {
		nodeList, err := getNodesForMachineSet(h, machineSetName)
		if nodeListError = err; err != nil {
			// If there's some error getting the node list, persist it
			// outside the loop in case it repeats.  We'll retry until the
			// timeout, though. We reset nodeListError every time so in the
			// case there's an error the first time but every other time it
			// works we don't create a red herring.
			time.Sleep(test.interval)
			continue
		}

		if len(nodeList) < 1 {
			// Make sure we're getting a list of nodes, otherwise
			// just sleep and retry (otherwise we get a nil pointer panic)
			time.Sleep(test.interval)
			continue
		}

		pass, err := test.test(nodeList)
		if err != nil {
			// Test explicitly errored.  Exit out.
			return err
		}

		if !pass {
			// Test did not pass, retry after interval
			time.Sleep(test.interval)
			continue
		}

		// test has passed, exit out
		return nil
	}

	// if listing the nodes returns an error consistently, let's return that,
	// otherwise let's just say the test timed out, which means MNMO isn't working.
	if nodeListError != nil {
		return nodeListError
	}
	return fmt.Errorf("Test has timed out, something is wrong with MNMO or there is a network issue with OCM")
}

// Note these are ORDERED tests.  While ordering tests is generally a code smell, in this case
// we don't want to do the expensive setup and teardown of the MachinePool, both in the cost of
// $$ for the machinepool but also for the time it takes for it to come online.  So these tests
// are Ordered to gain the most value, as they mimic a potential customer workflow over time

// The only real drawback here to the ordered tests is that if the label tests fail, it won't
// run the tests on the taint.
var _ = ginkgo.Describe(mnmoOperatorTestName, ginkgo.Ordered, func() {
	h := helper.New()
	var (
		ocm                *ocmprovider.OCMProvider
		clusterClient      *ocmTypes.ClusterClient
		clusterId          string
		machinePoolBuilder *ocmTypes.MachinePoolBuilder
		machineSetName     string
	)

	ginkgo.BeforeAll(func() {
		var err error
		ocm, err = ocmprovider.New()
		Expect(err).NotTo(HaveOccurred())
		Expect(ocm).NotTo(BeNil())

		clusterId = viper.GetString(config.Cluster.ID)

		// build the ocm client for this specific cluster
		clusterClient = ocm.GetConnection().ClustersMgmt().V1().Clusters().Cluster(clusterId)

		// build the cluster client to interact with things on-cluster with
		h.Impersonate(rest.ImpersonationConfig{
			UserName: "test-user@redhat.com",
			Groups: []string{
				"cluster-admins",
			},
		})
		// Before everything runs, we need to create a new MachinePoolBuilder to use for OCM
		machinePoolBuilder = ocmTypes.NewMachinePool().
			ID(ocmMachinePoolName)

		machineSetName = ensureMachinePoolExists(clusterClient, h)
		ginkgo.DeferCleanup(deleteMachinePool, clusterClient)
	})

	ginkgo.Context("When adding a label to a MachineSet", ginkgo.Ordered, func() {

		var (
			TestLabelKeyOne          string = "test-label-one"
			TestLabelValueOne        string = "test-value-one"
			TestLabelValueOneUpdated string = "test-updated"
			TestLabelKeyTwo          string = "second-test-label"
			TestLabelValueTwo        string = "test-value-two"
		)

		util.GinkgoIt("Applies a label to the Nodes and Machines of the MachineSet", func() {
			labels := map[string]string{
				TestLabelKeyOne: TestLabelValueOne,
			}
			machinePool, err := machinePoolBuilder.Labels(labels).Build()
			Expect(err).NotTo(HaveOccurred())

			_, err = clusterClient.MachinePools().MachinePool(ocmMachinePoolName).Update().Body(machinePool).Send()
			Expect(err).NotTo(HaveOccurred())

			test := MNMOTest{
				interval: 2 * time.Second,
				timeout:  2 * time.Minute,
				test: func(nodeList []*corev1.Node) (bool, error) {
					nodeLabels := nodeList[0].ObjectMeta.Labels
					for label, value := range nodeLabels {
						if label == TestLabelKeyOne {
							if value == TestLabelValueOne {
								return true, nil
							}
						}
					}
					return false, nil
				},
			}
			err = waitFor(test, h, machineSetName)
			Expect(err).NotTo(HaveOccurred())
		}, 30)

		util.GinkgoIt("Updates the label on the Nodes and Machines with a new value", func() {
			labels := map[string]string{
				TestLabelKeyOne: TestLabelValueOneUpdated,
			}
			machinePool, err := machinePoolBuilder.Labels(labels).Build()
			Expect(err).NotTo(HaveOccurred())

			_, err = clusterClient.MachinePools().MachinePool(ocmMachinePoolName).Update().Body(machinePool).Send()
			Expect(err).NotTo(HaveOccurred())

			test := MNMOTest{
				interval: 2 * time.Second,
				timeout:  2 * time.Minute,
				test: func(nodeList []*corev1.Node) (bool, error) {
					nodeLabels := nodeList[0].ObjectMeta.Labels
					for label, value := range nodeLabels {
						if label == TestLabelKeyOne {
							if value == TestLabelValueOneUpdated {
								return true, nil
							}
							return false, nil
						}
					}
					return false, nil
				},
			}
			err = waitFor(test, h, machineSetName)
			Expect(err).NotTo(HaveOccurred())
		}, 30)

		util.GinkgoIt("Adds a second label to the nodes and machines", func() {
			labels := map[string]string{
				TestLabelKeyOne: TestLabelValueOneUpdated,
				TestLabelKeyTwo: TestLabelValueTwo,
			}
			machinePool, err := machinePoolBuilder.Labels(labels).Build()
			Expect(err).NotTo(HaveOccurred())

			_, err = clusterClient.MachinePools().MachinePool(ocmMachinePoolName).Update().Body(machinePool).Send()
			Expect(err).NotTo(HaveOccurred())

			test := MNMOTest{
				interval: 2 * time.Second,
				timeout:  2 * time.Minute,
				test: func(nodeList []*corev1.Node) (bool, error) {
					nodeLabels := nodeList[0].ObjectMeta.Labels
					foundOne := false
					foundTwo := false
					for label, value := range nodeLabels {
						if label == TestLabelKeyOne {
							if value == TestLabelValueOneUpdated {
								foundOne = true
							}
						}
						if label == TestLabelKeyTwo {
							if value == TestLabelValueTwo {
								foundTwo = true
							}
						}
					}
					if foundOne && foundTwo {
						return true, nil
					}
					return false, nil
				},
			}
			err = waitFor(test, h, machineSetName)
			Expect(err).NotTo(HaveOccurred())
		}, 30)

		util.GinkgoIt("Removes a single label from the nodes and machines", func() {
			labels := map[string]string{
				TestLabelKeyTwo: TestLabelValueTwo,
			}
			machinePool, err := machinePoolBuilder.Labels(labels).Build()
			Expect(err).NotTo(HaveOccurred())

			_, err = clusterClient.MachinePools().MachinePool(ocmMachinePoolName).Update().Body(machinePool).Send()
			Expect(err).NotTo(HaveOccurred())

			test := MNMOTest{
				interval: 2 * time.Second,
				timeout:  2 * time.Minute,
				test: func(nodeList []*corev1.Node) (bool, error) {
					nodeLabels := nodeList[0].ObjectMeta.Labels
					foundTwo := false
					for label, value := range nodeLabels {
						if label == TestLabelKeyOne {
							return false, nil
						}
						if label == TestLabelKeyTwo {
							if value == TestLabelValueTwo {
								foundTwo = true
							}
						}
					}
					if foundTwo {
						return true, nil
					}
					return false, nil
				},
			}
			err = waitFor(test, h, machineSetName)
			Expect(err).NotTo(HaveOccurred())
		}, 30)

		util.GinkgoIt("Removes all labels from nodes and machines", func() {
			labels := map[string]string{}
			machinePool, err := machinePoolBuilder.Labels(labels).Build()
			Expect(err).NotTo(HaveOccurred())

			_, err = clusterClient.MachinePools().MachinePool(ocmMachinePoolName).Update().Body(machinePool).Send()
			Expect(err).NotTo(HaveOccurred())

			test := MNMOTest{
				interval: 2 * time.Second,
				timeout:  2 * time.Minute,
				test: func(nodeList []*corev1.Node) (bool, error) {
					nodeLabels := nodeList[0].ObjectMeta.Labels
					for label := range nodeLabels {
						if label == TestLabelKeyOne {
							return false, nil
						}
						if label == TestLabelKeyTwo {
							return false, nil
						}
					}
					return true, nil
				},
			}
			err = waitFor(test, h, machineSetName)
			Expect(err).NotTo(HaveOccurred())
		}, 30)
	})

	ginkgo.Context("When adding a Taint to a MachinePool", ginkgo.Ordered, func() {
		var (
			TestTaintKeyOne          string = "test-label-one"
			TestTaintValueOne        string = "test-value-one"
			TestTaintEffectOne       string = "NoSchedule"
			TestTaintValueOneUpdated string = "test-updated"
			TestTaintKeyTwo          string = "second-test-label"
			TestTaintValueTwo        string = "test-value-two"
			TestTaintEffectTwo       string = "NoExecute"
		)

		util.GinkgoIt("Applies a taint to the Nodes and Machines of the MachineSet", func() {
			testTaint := ocmTypes.NewTaint().Key(TestTaintKeyOne).Value(TestTaintValueOne).Effect(TestTaintEffectOne)

			machinePool, err := machinePoolBuilder.Taints(testTaint).Build()
			Expect(err).NotTo(HaveOccurred())

			_, err = clusterClient.MachinePools().MachinePool(ocmMachinePoolName).Update().Body(machinePool).Send()
			Expect(err).NotTo(HaveOccurred())

			test := MNMOTest{
				interval: 2 * time.Second,
				timeout:  2 * time.Minute,
				test: func(nodeList []*corev1.Node) (bool, error) {
					nodeTaints := nodeList[0].Spec.Taints
					for i := range nodeTaints {
						taint := nodeTaints[i]
						if taint.Key == TestTaintKeyOne {
							if taint.Value == TestTaintValueOne {
								if string(taint.Effect) == TestTaintEffectOne {
									return true, nil
								}
							}
						}
					}
					return false, nil
				},
			}
			err = waitFor(test, h, machineSetName)
			Expect(err).NotTo(HaveOccurred())
		}, 30)

		util.GinkgoIt("Updates the taint on the Nodes and Machines with a new value", func() {
			testTaint := ocmTypes.NewTaint().Key(TestTaintKeyOne).Value(TestTaintValueOneUpdated).Effect(TestTaintEffectOne)

			machinePool, err := machinePoolBuilder.Taints(testTaint).Build()
			Expect(err).NotTo(HaveOccurred())

			_, err = clusterClient.MachinePools().MachinePool(ocmMachinePoolName).Update().Body(machinePool).Send()
			Expect(err).NotTo(HaveOccurred())

			test := MNMOTest{
				interval: 2 * time.Second,
				timeout:  2 * time.Minute,
				test: func(nodeList []*corev1.Node) (bool, error) {
					nodeTaints := nodeList[0].Spec.Taints
					for i := range nodeTaints {
						taint := nodeTaints[i]
						if taint.Key == TestTaintKeyOne {
							if taint.Value == TestTaintValueOne {
								return false, nil
							}
							if taint.Value == TestTaintValueOneUpdated {
								if string(taint.Effect) == TestTaintEffectOne {
									return true, nil
								}
							}
						}
					}
					return false, nil
				},
			}
			err = waitFor(test, h, machineSetName)
			Expect(err).NotTo(HaveOccurred())
		}, 30)

		util.GinkgoIt("Adds a second taint to the nodes and machines", func() {
			testTaint := ocmTypes.NewTaint().Key(TestTaintKeyOne).Value(TestTaintValueOneUpdated).Effect(TestTaintEffectOne)
			testTaintTwo := ocmTypes.NewTaint().Key(TestTaintKeyTwo).Value(TestTaintValueTwo).Effect(TestTaintEffectTwo)

			machinePool, err := machinePoolBuilder.Taints(testTaint, testTaintTwo).Build()
			Expect(err).NotTo(HaveOccurred())

			_, err = clusterClient.MachinePools().MachinePool(ocmMachinePoolName).Update().Body(machinePool).Send()
			Expect(err).NotTo(HaveOccurred())

			test := MNMOTest{
				interval: 2 * time.Second,
				timeout:  2 * time.Minute,
				test: func(nodeList []*corev1.Node) (bool, error) {
					nodeTaints := nodeList[0].Spec.Taints
					foundOne := false
					foundTwo := false
					for i := range nodeTaints {
						taint := nodeTaints[i]
						if taint.Key == TestTaintKeyOne {
							if taint.Value == TestTaintValueOneUpdated {
								if string(taint.Effect) == TestTaintEffectOne {
									foundOne = true
								}
							}
						}
						if taint.Key == TestTaintKeyTwo {
							if taint.Value == TestTaintValueTwo {
								if string(taint.Effect) == TestTaintEffectTwo {
									foundTwo = true
								}
							}
						}
					}
					if foundOne && foundTwo {
						return true, nil
					}
					return false, nil
				},
			}
			err = waitFor(test, h, machineSetName)
			Expect(err).NotTo(HaveOccurred())
		}, 30)

		util.GinkgoIt("Removes a single taint from the nodes and machines", func() {
			testTaintTwo := ocmTypes.NewTaint().Key(TestTaintKeyTwo).Value(TestTaintValueTwo).Effect(TestTaintEffectTwo)

			machinePool, err := machinePoolBuilder.Taints(testTaintTwo).Build()
			Expect(err).NotTo(HaveOccurred())

			_, err = clusterClient.MachinePools().MachinePool(ocmMachinePoolName).Update().Body(machinePool).Send()
			Expect(err).NotTo(HaveOccurred())

			test := MNMOTest{
				interval: 2 * time.Second,
				timeout:  2 * time.Minute,
				test: func(nodeList []*corev1.Node) (bool, error) {
					nodeTaints := nodeList[0].Spec.Taints
					foundTwo := false
					for i := range nodeTaints {
						taint := nodeTaints[i]
						if taint.Key == TestTaintKeyOne {
							return false, nil
						}
						if taint.Key == TestTaintKeyTwo {
							if taint.Value == TestTaintValueTwo {
								if string(taint.Effect) == TestTaintEffectTwo {
									foundTwo = true
								}
							}
						}
					}
					if foundTwo {
						return true, nil
					}
					return false, nil
				},
			}
			err = waitFor(test, h, machineSetName)
			Expect(err).NotTo(HaveOccurred())
		}, 30)

		util.GinkgoIt("Removes all taints from nodes and machines", func() {
			machinePool, err := machinePoolBuilder.Taints().Build()
			Expect(err).NotTo(HaveOccurred())

			_, err = clusterClient.MachinePools().MachinePool(ocmMachinePoolName).Update().Body(machinePool).Send()
			Expect(err).NotTo(HaveOccurred())

			test := MNMOTest{
				interval: 2 * time.Second,
				timeout:  2 * time.Minute,
				test: func(nodeList []*corev1.Node) (bool, error) {
					nodeTaints := nodeList[0].Spec.Taints
					foundOne := false
					foundTwo := false
					for i := range nodeTaints {
						taint := nodeTaints[i]
						if taint.Key == TestTaintKeyOne {
							foundOne = true
						}
						if taint.Key == TestTaintKeyTwo {
							foundTwo = true
						}
					}
					if foundOne || foundTwo {
						return false, nil
					}
					return true, nil
				},
			}
			err = waitFor(test, h, machineSetName)
			Expect(err).NotTo(HaveOccurred())
		}, 30)
	})
})
