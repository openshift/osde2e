package operators

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	operatorv1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func checkClusterServiceVersion(h *helper.H, namespace, name string) {
	// Check that the operator clusterServiceVersion exists
	ginkgo.Context(fmt.Sprintf("clusterServiceVersion %s/%s", namespace, name), func() {
		ginkgo.It("should be present and in succeeded state", func(ctx context.Context) {
			Eventually(func() bool {
				csvList, err := h.Operator().OperatorsV1alpha1().ClusterServiceVersions(namespace).List(ctx, metav1.ListOptions{})
				if err != nil {
					log.Printf("failed to get CSVs in namespace %s: %v", namespace, err)
					return false
				}
				for _, csv := range csvList.Items {
					if csv.Spec.DisplayName == name && csv.Status.Phase == operatorv1.CSVPhaseSucceeded {
						return true
					}
				}
				return false
			}, "30m", "15s").Should(BeTrue())
		})
	})
}

func checkConfigMapLockfile(h *helper.H, namespace, operatorLockFile string) {
	// Check that the operator configmap has been deployed
	ginkgo.Context("configmaps", func() {
		ginkgo.It("should exist", func(ctx context.Context) {
			// Wait for lockfile to signal operator is active
			err := pollLockFile(ctx, h, namespace, operatorLockFile)
			Expect(err).ToNot(HaveOccurred(), "failed fetching the configMap lockfile")
		})
	})
}

func checkDeployment(h *helper.H, namespace string, name string, defaultDesiredReplicas int32) {
	// Check that the operator deployment exists in the operator namespace
	ginkgo.Context("deployment", func() {
		ginkgo.It("should exist and be available", func(ctx context.Context) {
			deployment, err := pollDeployment(ctx, h, namespace, name)
			Expect(err).ToNot(HaveOccurred(), "failed fetching deployment")
			Expect(deployment).NotTo(BeNil(), "deployment is nil")

			readyReplicas := deployment.Status.ReadyReplicas
			desiredReplicas := deployment.Status.Replicas

			// The desired replicas should match the default installed replica count
			Expect(desiredReplicas).To(BeNumerically("==", defaultDesiredReplicas), "The deployment desired replicas should not drift from the default 1.")

			// Desired replica count should match ready replica count
			Expect(readyReplicas).To(BeNumerically("==", desiredReplicas), "All desired replicas should be ready.")
		})
	})
}

func checkPod(h *helper.H, namespace string, name string, gracePeriod int, maxAcceptedRestart int) {
	// Checks that deployed pods have less than maxAcceptedRestart restarts

	ginkgo.Context("pods", func() {
		ginkgo.It(fmt.Sprintf("should have %v or less restart(s)", maxAcceptedRestart), func(ctx context.Context) {
			// wait for graceperiod
			time.Sleep(time.Duration(gracePeriod) * time.Second)
			// retrieve pods
			pods, err := h.Kube().CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: "name=" + name})
			Expect(err).ToNot(HaveOccurred(), "failed fetching pods")

			var restartSum int32 = 0
			for _, pod := range pods.Items {
				for _, status := range pod.Status.ContainerStatuses {
					restartSum += status.RestartCount
				}
			}
			Expect(restartSum).To(BeNumerically("<=", maxAcceptedRestart))
		})
	})
}

func checkClusterRoles(h *helper.H, clusterRoles []string, matchPrefix bool) {
	// Check that the clusterRoles exist
	ginkgo.Context("clusterRoles", func() {
		ginkgo.It("should exist", func(ctx context.Context) {
			allClusterRoles, err := h.Kube().RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
			Expect(err).ToNot(HaveOccurred(), "failed to list clusterRoles\n")

			for _, clusterRoleToFind := range clusterRoles {
				found := false
				for _, clusterRole := range allClusterRoles.Items {
					if (matchPrefix && strings.HasPrefix(clusterRole.Name, clusterRoleToFind)) ||
						(!matchPrefix && clusterRole.Name == clusterRoleToFind) {
						found = true
						break
					}
				}
				Expect(found).To(BeTrue(), "failed to find ClusterRole %s\n", clusterRoleToFind)
			}
		})
	})
}

func checkClusterRoleBindings(h *helper.H, clusterRoleBindings []string, matchPrefix bool) {
	// Check that the clusterRoles exist
	ginkgo.Context("clusterRoleBindings", func() {
		ginkgo.It("should exist", func(ctx context.Context) {
			allClusterRoleBindings, err := h.Kube().RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
			Expect(err).ToNot(HaveOccurred(), "failed to list clusterRoles\n")

			for _, clusterRoleBindingToFind := range clusterRoleBindings {
				found := false
				for _, clusterRole := range allClusterRoleBindings.Items {
					if (matchPrefix && strings.HasPrefix(clusterRole.Name, clusterRoleBindingToFind)) ||
						(!matchPrefix && clusterRole.Name == clusterRoleBindingToFind) {
						found = true
						break
					}
				}
				Expect(found).To(BeTrue(), "failed to find ClusterRoleBinding %s\n", clusterRoleBindingToFind)
			}
		})
	})
}

func checkRole(h *helper.H, namespace string, roles []string) {
	// Check that deployed roles exist
	ginkgo.Context("roles", func() {
		ginkgo.It("should exist", func(ctx context.Context) {
			for _, roleName := range roles {
				_, err := h.Kube().RbacV1().Roles(namespace).Get(ctx, roleName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "failed to get role %v\n", roleName)
			}
		})
	})
}

func checkRoleBindings(h *helper.H, namespace string, roleBindings []string) {
	// Check that deployed rolebindings exist
	ginkgo.Context("roleBindings", func() {
		ginkgo.It("should exist", func(ctx context.Context) {
			for _, roleBindingName := range roleBindings {
				err := pollRoleBinding(ctx, h, namespace, roleBindingName)
				Expect(err).NotTo(HaveOccurred(), "failed to get roleBinding %v\n", roleBindingName)
			}
		})
	})
}

func PerformUpgrade(ctx context.Context, h *helper.H, subNamespace string, subName string) error {
	return h.GetClient().UpgradeOperator(ctx, subName, subNamespace)
}

func checkUpgrade(h *helper.H, subNamespace string, subName string, packageName string, regServiceName string) {
	ginkgo.Context("Operator Upgrade", func() {
		ginkgo.It("should upgrade from the replaced version", func(ctx context.Context) {
			err := PerformUpgrade(ctx, h, subNamespace, subName)
			Expect(err).NotTo(HaveOccurred(), "unable to upgrade operator %s", subName)
		})
	})
}

func checkService(h *helper.H, namespace string, name string, port int) {
	ginkgo.Context("service", func() {
		ginkgo.It(
			"should exist",
			func(ctx context.Context) {
				Eventually(func() bool {
					_, err := h.Kube().CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
					if err != nil {
						return false
					}
					return true
				}, "30m", "1m").Should(BeTrue())
			},
		)
	})
}

func pollRoleBinding(ctx context.Context, h *helper.H, projectName string, roleBindingName string) error {
	// pollRoleBinding will check for the existence of a roleBinding
	// in the specified project, and wait for it to exist, until a timeout

	var err error
	// interval is the duration in seconds between polls
	// values here for humans

	interval := 5

	// convert time.Duration type
	timeoutDuration := time.Duration(viper.GetFloat64(config.Tests.PollingTimeout)) * time.Second
	intervalDuration := time.Duration(interval) * time.Second

	start := time.Now()

Loop:
	for {
		_, err = h.Kube().RbacV1().RoleBindings(projectName).Get(ctx, roleBindingName, metav1.GetOptions{})
		elapsed := time.Since(start)

		switch {
		case err == nil:
			// Success
			break Loop
		case strings.Contains(err.Error(), "forbidden"):
			return err
		default:
			if elapsed < timeoutDuration {
				log.Printf("Waiting %v for %s roleBinding to exist", (timeoutDuration - elapsed), roleBindingName)
				time.Sleep(intervalDuration)
			} else {
				err = fmt.Errorf("failed to get rolebinding %s before timeout", roleBindingName)
				break Loop
			}
		}
	}

	return err
}

func pollLockFile(ctx context.Context, h *helper.H, namespace, operatorLockFile string) error {
	// GetConfigMap polls for a configMap with a timeout
	// to handle the case when a new cluster is up but the OLM has not yet
	// finished deploying the operator

	var err error

	// interval is the duration in seconds between polls
	// values here for humans
	interval := 30

	// convert time.Duration type
	timeoutDuration := time.Duration(viper.GetFloat64(config.Tests.PollingTimeout)) * time.Second
	intervalDuration := time.Duration(interval) * time.Second

	start := time.Now()

Loop:
	for {
		_, err = h.Kube().CoreV1().ConfigMaps(namespace).Get(ctx, operatorLockFile, metav1.GetOptions{})
		elapsed := time.Since(start)

		switch {
		case err == nil:
			// Success
			break Loop
		case strings.Contains(err.Error(), "forbidden"):
			return err
		default:
			if elapsed < timeoutDuration {
				log.Printf("Waiting %v for %s configMap to exist", (timeoutDuration - elapsed), operatorLockFile)
				time.Sleep(intervalDuration)
			} else {
				err = fmt.Errorf("failed to get configMap %s before timeout", operatorLockFile)
				break Loop
			}
		}
	}

	return err
}

func pollDeployment(ctx context.Context, h *helper.H, namespace, deploymentName string) (*appsv1.Deployment, error) {
	// pollDeployment polls for a deployment with a timeout
	// to handle the case when a new cluster is up but the OLM has not yet
	// finished deploying the operator

	var err error
	var deployment *appsv1.Deployment

	// interval is the duration in seconds between polls
	// values here for humans
	interval := 5

	// convert time.Duration type
	timeoutDuration := time.Duration(viper.GetFloat64(config.Tests.PollingTimeout)) * time.Second
	intervalDuration := time.Duration(interval) * time.Second

	start := time.Now()

Loop:
	for {
		deployment, err = h.Kube().AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
		elapsed := time.Since(start)

		switch {
		case err == nil:
			// Success
			break Loop
		case strings.Contains(err.Error(), "forbidden"):
			return nil, err
		default:
			if elapsed < timeoutDuration {
				log.Printf("Waiting %v for %s deployment to exist", (timeoutDuration - elapsed), deploymentName)
				time.Sleep(intervalDuration)
			} else {
				deployment = nil
				err = fmt.Errorf("failed to get %s Deployment before timeout", deploymentName)
				break Loop
			}
		}
	}

	return deployment, err
}

func CheckUpgrade(h *helper.H, subNamespace string, subName string, packageName string, regServiceName string) {
	checkUpgrade(h, subNamespace, subName, packageName, regServiceName)
}

func CheckPod(h *helper.H, namespace string, name string, gracePeriod int, maxAcceptedRestart int) {
	checkPod(h, namespace, name, gracePeriod, maxAcceptedRestart)
}

func PollDeployment(ctx context.Context, h *helper.H, namespace, deploymentName string) (*appsv1.Deployment, error) {
	return pollDeployment(ctx, h, namespace, deploymentName)
}
