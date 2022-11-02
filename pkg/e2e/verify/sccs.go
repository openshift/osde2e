package verify

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "github.com/openshift/api/security/v1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

var dedicatedAdminSccTestName = "[Suite: e2e] [OSD] RBAC Dedicated Admins SCC permissions"

func init() {
	alert.RegisterGinkgoAlert(dedicatedAdminSccTestName, "SD-CICD", "Matt Bargenquast", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(dedicatedAdminSccTestName, func() {
	h := helper.New()

	workloadDir := "workloads/e2e/scc"
	// How long to wait for prometheus pods to restart
	prometheusRestartPollingDuration := 4 * time.Minute

	ginkgo.Context("Dedicated Admin permissions", func() {

		util.GinkgoIt("should include anyuid", func() {
			checkSccPermissions(h, "dedicated-admins-cluster", "anyuid")
		})

		util.GinkgoIt("should include nonroot", func() {
			checkSccPermissions(h, "dedicated-admins-cluster", "nonroot")
		})

		util.GinkgoIt("can create pods with SCCs", func() {
			_, err := helper.ApplyYamlInFolder(workloadDir, h.CurrentProject(), h.Kube())
			Expect(err).NotTo(HaveOccurred(), "couldn't apply workload yaml")
		})
	})
	ginkgo.Context("scc-test", func() {
		util.GinkgoIt("new SCC does not break pods", func() {
			//Test to verify that creation of a permissive scc does not disrupt ability to run pods https://bugzilla.redhat.com/show_bug.cgi?id=1868976
			newScc := makeMinimalSCC("scc-test")
			log.Printf("SCC:(%v)", newScc)
			err := createScc(newScc, h)
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				err = deleteScc("scc-test", h)
				Expect(err).NotTo(HaveOccurred())
			}()
			log.Printf("Error:(%v)", err)

			// Reestarting the prometheus operator
			err = restartOperator(h, "prometheus-operator", "openshift-monitoring")
			Expect(err).NotTo(HaveOccurred())
			log.Printf("Error:(%v)", err)
			//Deleting all prometheus pods
			list, _ := FilterPods("openshift-monitoring", "app.kubernetes.io/name=prometheus", h)
			names, _ := GetPodNames(list, h)
			log.Printf("Names of pods:(%v)", names)
			numPrometheusPods := deletePods(names, "openshift-monitoring", h)
			//Verifying the same number of running prometheus pods has come up
			err = wait.PollImmediate(2*time.Second, prometheusRestartPollingDuration, func() (bool, error) {
				pollList, _ := FilterPods("openshift-monitoring", "app.kubernetes.io/name=prometheus", h)
				if !AllDifferentPods(list, pollList) {
					return false, nil
				}
				_, newNamesNum := GetPodNames(pollList, h)
				if numPrometheusPods == newNamesNum {
					return true, nil
				}
				return false, nil
			})
			Expect(err).NotTo(HaveOccurred())
		}, (prometheusRestartPollingDuration + 30*time.Second).Seconds())
	})
})

// AllDifferentPods returns whether or not the newPods contains a pod with
// the same UID as a pod in originalPods. It is useful for ensureing that
// all pods in a given deployment have been restarted (the new ones will
// have different UIDs).
func AllDifferentPods(originalPods, newPods *apiv1.PodList) bool {
	orig := make(map[types.UID]struct{})
	for _, p := range originalPods.Items {
		orig[p.ObjectMeta.UID] = struct{}{}
	}
	for _, p := range newPods.Items {
		if _, ok := orig[p.ObjectMeta.UID]; ok {
			return false
		}
	}
	return true
}

func checkSccPermissions(h *helper.H, clusterRole string, scc string) {

	// Get the cluster role containing the definition
	cr, err := h.Kube().RbacV1().ClusterRoles().Get(context.TODO(), clusterRole, metav1.GetOptions{})
	Expect(err).ToNot(HaveOccurred(), "failed to get clusterRole %s\n", clusterRole)

	foundRule := false
	for _, rule := range cr.Rules {

		// Find rules relating to SCCs
		isSccRule := false
		for _, resource := range rule.Resources {
			if resource == "securitycontextconstraints" {
				isSccRule = true
			}
		}
		if !isSccRule {
			continue
		}

		// check for 'use' verb
		for _, verb := range rule.Verbs {
			if verb == "use" {
				foundRule = true
				break
			}
		}
	}
	Expect(foundRule).To(BeTrue())
}

func makeMinimalSCC(name string) v1.SecurityContextConstraints {
	scc := v1.SecurityContextConstraints{

		TypeMeta: metav1.TypeMeta{

			Kind:       "SecurityContextConstraints",
			APIVersion: "security.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{

			Name: name,
		},
		Groups: []string{
			"system:authenticated",
		},

		SELinuxContext: v1.SELinuxContextStrategyOptions{
			Type: v1.SELinuxStrategyRunAsAny,
		},
		RunAsUser: v1.RunAsUserStrategyOptions{
			Type: v1.RunAsUserStrategyRunAsAny,
		},
		FSGroup: v1.FSGroupStrategyOptions{
			Type: v1.FSGroupStrategyRunAsAny,
		},
		SupplementalGroups: v1.SupplementalGroupsStrategyOptions{
			Type: v1.SupplementalGroupsStrategyRunAsAny,
		},
	}
	return scc
}

func createScc(scc v1.SecurityContextConstraints, h *helper.H) error {

	_, err := h.Security().SecurityV1().SecurityContextConstraints().Create(context.TODO(), &scc, metav1.CreateOptions{})
	return (err)
}

func deleteScc(scc string, h *helper.H) error {

	err := h.Security().SecurityV1().SecurityContextConstraints().Delete(context.TODO(), scc, metav1.DeleteOptions{})
	return (err)
}

//Filters pods based on namespace and label
func FilterPods(namespace string, label string, h *helper.H) (*apiv1.PodList, error) {

	list, err := h.Kube().CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: label})
	if err != nil {
		log.Printf("Could not issue create command")
		return list, err
	}

	return list, err

}

//Extracts pod names from a filtered list and counts how many are in running state
func GetPodNames(list *apiv1.PodList, h *helper.H) ([]string, int) {
	var notReady []apiv1.Pod
	var ready []apiv1.Pod
	var podNames []string
	var total int
	var numReady int

podLoop:
	for _, pod := range list.Items {
		name := pod.Name
		podNames = append(podNames, name)

		phase := pod.Status.Phase
		if phase != apiv1.PodRunning && phase != apiv1.PodSucceeded {
			notReady = append(notReady, pod)
		} else {
			for _, status := range pod.Status.ContainerStatuses {
				if !status.Ready {
					notReady = append(notReady, pod)
					continue podLoop
				}
			}
			ready = append(ready, pod)

		}
	}
	total = len(list.Items)
	numReady = (total - len(notReady))

	if total != numReady {
		log.Printf(" %v out of %v pods were/was in Ready state.", numReady, total)
	}
	return podNames, numReady
}

// Scales down and scales up the operator deployment to initiate a pod restart
func restartOperator(h *helper.H, operator string, ns string) error {

	log.Printf("restarting %s operator to force re-initialize pods", operator)

	err := wait.PollImmediate(5*time.Second, 2*time.Minute, func() (bool, error) {
		// scale down
		s, err := h.Kube().AppsV1().Deployments(ns).GetScale(context.TODO(), operator, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		sc := *s
		sc.Spec.Replicas = 0
		_, err = h.Kube().AppsV1().Deployments(ns).UpdateScale(context.TODO(), operator, &sc, metav1.UpdateOptions{})
		if err != nil {
			return false, nil
		}

		// scale up
		sc.Spec.Replicas = 1
		_, err = h.Kube().AppsV1().Deployments(ns).UpdateScale(context.TODO(), operator, &sc, metav1.UpdateOptions{})
		if err != nil {
			return false, nil
		}
		log.Printf(" %s operator restart complete..", operator)
		return true, nil
	})

	if err != nil {
		return fmt.Errorf("couldn't restart %s operator to re-initiate pods: %v", operator, err)
	}
	return nil
}

func deletePods(names []string, namespace string, h *helper.H) int {
	numPods := 0
	for _, name := range names {
		numPods = numPods + 1
		_, err := h.Kube().CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		log.Printf("Check before deleting pod %s, error: %v", name, err)
		if err == nil {
			err := h.Kube().CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
			log.Printf("Deleting pod %s, error: %v", name, err)
		}
	}
	return numPods
}
