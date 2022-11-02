package operators

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/templates"
	"github.com/openshift/osde2e/pkg/common/util"

	appsv1 "k8s.io/api/apps/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	operatorv1 "github.com/operator-framework/api/pkg/operators/v1alpha1"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
)

func checkClusterServiceVersion(h *helper.H, namespace, name string) {
	// Check that the operator clusterServiceVersion exists
	ginkgo.Context(fmt.Sprintf("clusterServiceVersion %s/%s", namespace, name), func() {
		util.GinkgoIt("should be present and in succeeded state", func() {
			Eventually(func() bool {
				csvList, err := h.Operator().OperatorsV1alpha1().ClusterServiceVersions(namespace).List(context.TODO(), metav1.ListOptions{})
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
		}, viper.GetFloat64(config.Tests.PollingTimeout))
	})
}

func checkConfigMapLockfile(h *helper.H, namespace, operatorLockFile string) {
	// Check that the operator configmap has been deployed
	ginkgo.Context("configmaps", func() {
		util.GinkgoIt("should exist", func() {
			// Wait for lockfile to signal operator is active
			err := pollLockFile(h, namespace, operatorLockFile)
			Expect(err).ToNot(HaveOccurred(), "failed fetching the configMap lockfile")
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func checkDeployment(h *helper.H, namespace string, name string, defaultDesiredReplicas int32) {
	// Check that the operator deployment exists in the operator namespace
	ginkgo.Context("deployment", func() {
		util.GinkgoIt("should exist", func() {
			deployment, err := pollDeployment(h, namespace, name)
			Expect(err).ToNot(HaveOccurred(), "failed fetching deployment")
			Expect(deployment).NotTo(BeNil(), "deployment is nil")
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
		util.GinkgoIt("should have all desired replicas ready", func() {
			deployment, err := pollDeployment(h, namespace, name)
			Expect(err).ToNot(HaveOccurred(), "failed fetching deployment")

			readyReplicas := deployment.Status.ReadyReplicas
			desiredReplicas := deployment.Status.Replicas

			// The desired replicas should match the default installed replica count
			Expect(desiredReplicas).To(BeNumerically("==", defaultDesiredReplicas), "The deployment desired replicas should not drift from the default 1.")

			// Desired replica count should match ready replica count
			Expect(readyReplicas).To(BeNumerically("==", desiredReplicas), "All desired replicas should be ready.")
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func checkPod(h *helper.H, namespace string, name string, gracePeriod int, maxAcceptedRestart int) {
	// Checks that deployed pods have less than maxAcceptedRestart restarts

	ginkgo.Context("pods", func() {
		util.GinkgoIt(fmt.Sprintf("should have %v or less restart(s)", maxAcceptedRestart), func() {
			// wait for graceperiod
			time.Sleep(time.Duration(gracePeriod) * time.Second)
			//retrieve pods
			pods, err := h.Kube().CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "name=" + name})
			Expect(err).ToNot(HaveOccurred(), "failed fetching pods")

			var restartSum int32 = 0
			for _, pod := range pods.Items {
				for _, status := range pod.Status.ContainerStatuses {
					restartSum += status.RestartCount
				}
			}
			Expect(restartSum).To(BeNumerically("<=", maxAcceptedRestart))
		}, float64(gracePeriod)+viper.GetFloat64(config.Tests.PollingTimeout))
	})
}

func checkServiceAccounts(h *helper.H, operatorNamespace string, serviceAccounts []string) {
	// Check that deployed serviceAccounts exist
	ginkgo.Context("serviceAccounts", func() {
		util.GinkgoIt("should exist", func() {
			for _, serviceAccountName := range serviceAccounts {
				_, err := h.Kube().CoreV1().ServiceAccounts(operatorNamespace).Get(context.TODO(), serviceAccountName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "failed to get serviceAccount %v\n", serviceAccountName)
			}
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func checkClusterRoles(h *helper.H, clusterRoles []string, matchPrefix bool) {
	// Check that the clusterRoles exist
	ginkgo.Context("clusterRoles", func() {
		util.GinkgoIt("should exist", func() {
			allClusterRoles, err := h.Kube().RbacV1().ClusterRoles().List(context.TODO(), metav1.ListOptions{})
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
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func checkClusterRoleBindings(h *helper.H, clusterRoleBindings []string, matchPrefix bool) {
	// Check that the clusterRoles exist
	ginkgo.Context("clusterRoleBindings", func() {
		util.GinkgoIt("should exist", func() {
			allClusterRoleBindings, err := h.Kube().RbacV1().ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})
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
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func checkRole(h *helper.H, namespace string, roles []string) {
	// Check that deployed roles exist
	ginkgo.Context("roles", func() {
		util.GinkgoIt("should exist", func() {
			for _, roleName := range roles {
				_, err := h.Kube().RbacV1().Roles(namespace).Get(context.TODO(), roleName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "failed to get role %v\n", roleName)
			}
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})

}

func checkRolesWithNamePrefix(h *helper.H, namespace string, prefix string, count int) {
	ginkgo.Context("roles with prefix", func() {
		util.GinkgoIt("should exist", func() {
			Eventually(func() int {
				rolesList, err := h.Kube().RbacV1().Roles(namespace).List(context.TODO(), metav1.ListOptions{})
				Expect(err).NotTo(HaveOccurred(), "failed to get roles in namespace %s", namespace)
				var roleCount int
				for _, r := range rolesList.Items {
					if strings.HasPrefix(r.Name, prefix) {
						roleCount++
					}
				}
				return roleCount
			}, "10m", "30s").Should(BeNumerically(">=", count))
		}, viper.GetFloat64(config.Tests.PollingTimeout))
	})
}

func checkRoleBindingsWithNamePrefix(h *helper.H, namespace string, prefix string, count int) {
	ginkgo.Context("roles with prefix", func() {
		util.GinkgoIt("should exist", func() {
			Eventually(func() int {
				roleBindings, err := h.Kube().RbacV1().RoleBindings(namespace).List(context.TODO(), metav1.ListOptions{})
				Expect(err).NotTo(HaveOccurred(), "failed to get roles in namespace %s", namespace)
				var roleCount int
				for _, r := range roleBindings.Items {
					if strings.HasPrefix(r.Name, prefix) {
						roleCount++
					}
				}
				return roleCount
			}, "10m", "30s").Should(BeNumerically(">=", count))
		}, viper.GetFloat64(config.Tests.PollingTimeout))
	})
}
func checkRoleBindings(h *helper.H, namespace string, roleBindings []string) {
	// Check that deployed rolebindings exist
	ginkgo.Context("roleBindings", func() {
		util.GinkgoIt("should exist", func() {
			for _, roleBindingName := range roleBindings {
				err := pollRoleBinding(h, namespace, roleBindingName)
				Expect(err).NotTo(HaveOccurred(), "failed to get roleBinding %v\n", roleBindingName)
			}
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

//nolint
func checkSecrets(h *helper.H, namespace string, secrets []string) {
	// Check that deployed secrets exist
	ginkgo.Context("secrets", func() {
		util.GinkgoIt("should exist", func() {
			for _, secretName := range secrets {
				_, err := h.Kube().CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "failed to get secret %v\n", secretName)
			}
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func checkUpgrade(h *helper.H, subNamespace string, subName string, packageName string, regServiceName string) {

	ginkgo.Context("Operator Upgrade", func() {

		installPlanPollingDuration := 5 * time.Minute
		upgradePollingDuration := 15 * time.Minute

		util.GinkgoIt("should upgrade from the replaced version", func() {

			var latestCSV string
			var sub *operatorv1.Subscription
			var err error

			// The subscription must first exist on the cluster
			sub, err = h.Operator().OperatorsV1alpha1().Subscriptions(subNamespace).Get(context.TODO(), subName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("subscription %s not found", subName))

			// Get the CSV we're currently installed with
			installedCSVs, err := h.Operator().OperatorsV1alpha1().ClusterServiceVersions(subNamespace).List(context.TODO(), metav1.ListOptions{})
			for _, csv := range installedCSVs.Items {
				if csv.Spec.DisplayName == packageName && csv.Status.Phase == operatorv1.CSVPhaseSucceeded {
					latestCSV = csv.Name
				}
			}

			// If we couldn't find a Succeeded CSV, then the operator is likely not even installed
			Expect(latestCSV).NotTo(BeEmpty(), fmt.Sprintf("no successfully installed CSV found for subscription %s", subName))

			// Get the N-1 version of the CSV to test an upgrade from
			previousCSV, err := getReplacesCSV(h, subNamespace, packageName, regServiceName)
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed trying to get previous CSV for Subscription %s in %s namespace", subName, subNamespace))

			log.Printf("Reverting to package %v from %v to test upgrade of %v", previousCSV, latestCSV, subName)

			// Delete current Operator installation
			err = h.Operator().OperatorsV1alpha1().Subscriptions(subNamespace).Delete(context.TODO(), subName, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed trying to delete Subscription %s", subName))
			log.Printf("Removed subscription %s", subName)

			err = h.Operator().OperatorsV1alpha1().ClusterServiceVersions(subNamespace).Delete(context.TODO(), latestCSV, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed trying to delete ClusterServiceVersion %s", latestCSV))
			log.Printf("Removed csv %s", latestCSV)

			err = wait.PollImmediate(10*time.Second, installPlanPollingDuration, func() (bool, error) {
				ips, err := h.Operator().OperatorsV1alpha1().InstallPlans(subNamespace).List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					return false, err
				}
				// If an InstallPlan exists that is associated with the subscription, then the oeprator hasn't been fully removed
				for _, ip := range ips.Items {
					for _, csvName := range ip.Spec.ClusterServiceVersionNames {
						if latestCSV == csvName {
							return false, nil
						}
					}
				}
				return true, nil
			})
			Expect(err).NotTo(HaveOccurred(), "installplan never garbage collected")
			log.Printf("Verified installplan removal")

			// Create subscription to the previous version
			sub, err = h.Operator().OperatorsV1alpha1().Subscriptions(subNamespace).Create(context.TODO(), &operatorv1.Subscription{
				ObjectMeta: metav1.ObjectMeta{
					Name:      subName,
					Namespace: subNamespace,
				},
				Spec: &operatorv1.SubscriptionSpec{
					Package:                sub.Spec.Package,
					Channel:                sub.Spec.Channel,
					CatalogSourceNamespace: sub.Spec.CatalogSourceNamespace,
					CatalogSource:          sub.Spec.CatalogSource,
					InstallPlanApproval:    operatorv1.ApprovalAutomatic,
					StartingCSV:            previousCSV,
				},
			}, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed trying to create Subscription %s", subName))

			log.Printf("Created replacement subscription %s with starting CSV %s", subName, previousCSV)

			// Wait for the operator to arrive back on its latest CSV
			err = wait.PollImmediate(5*time.Second, upgradePollingDuration, func() (bool, error) {
				csv, err := h.Operator().OperatorsV1alpha1().ClusterServiceVersions(sub.Namespace).Get(context.TODO(), latestCSV, metav1.GetOptions{})
				if err != nil && !kerror.IsNotFound(err) {
					log.Printf("Returning err %v", err)
					return false, err
				}
				if csv.Status.Phase == operatorv1.CSVPhaseSucceeded {
					return true, nil
				}
				return false, nil
			})
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("CSV %s did not eventually install successfully", latestCSV))

			// Lastly, verify that the Subscription correctly reflects that the CSV is installed.
			err = wait.PollImmediate(5*time.Second, 2*time.Minute, func() (bool, error) {
				sub, err = h.Operator().OperatorsV1alpha1().Subscriptions(subNamespace).Get(context.TODO(), subName, metav1.GetOptions{})
				if err != nil {
					return false, err
				}
				currentCSV := sub.Status.CurrentCSV
				if currentCSV == latestCSV {
					return true, nil
				}
				return false, nil
			})
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("subscription %s status is not reflecting that csv %s is installed", subName, latestCSV))

		}, upgradePollingDuration.Seconds()+installPlanPollingDuration.Seconds()+
			float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func checkService(h *helper.H, namespace string, name string, port int) {
	pollTimeout := viper.GetFloat64(config.Tests.PollingTimeout)
	ginkgo.Context("service", func() {
		util.GinkgoIt(
			"should exist",
			func() {
				Eventually(func() bool {
					_, err := h.Kube().CoreV1().Services(namespace).Get(context.Background(), name, metav1.GetOptions{})
					if err != nil {
						return false
					}
					return true
				}, "30m", "1m").Should(BeTrue())
			},
			pollTimeout,
		)
	})
}

func pollClusterRoleBinding(h *helper.H, clusterRoleBindingName string) error {
	// pollRoleBinding will check for the existence of a clusterRole
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
		_, err = h.Kube().RbacV1().ClusterRoleBindings().Get(context.TODO(), clusterRoleBindingName, metav1.GetOptions{})
		elapsed := time.Since(start)

		switch {
		case err == nil:
			// Success
			break Loop
		case strings.Contains(err.Error(), "forbidden"):
			return err
		default:
			if elapsed < timeoutDuration {
				log.Printf("Waiting %v for %s clusterRoleBinding to exist", (timeoutDuration - elapsed), clusterRoleBindingName)
				time.Sleep(intervalDuration)
			} else {
				err = fmt.Errorf("Failed to get clusterRolebinding %s before timeout", clusterRoleBindingName)
				break Loop
			}
		}
	}

	return err
}

func pollRoleBinding(h *helper.H, projectName string, roleBindingName string) error {
	// pollRoleBinding will check for the existence of a roleBinding
	// in the specified project, and wait for it to exist, until a timeout

	var err error
	// interval is the duration in seconds between polls
	// values here for humans

	interval := 5

	// convert time.Duration type
	timeoutDuration := time.Duration(viper.GetFloat64(config.Tests.PollingTimeout)) * time.Minute
	intervalDuration := time.Duration(interval) * time.Second

	start := time.Now()

Loop:
	for {
		_, err = h.Kube().RbacV1().RoleBindings(projectName).Get(context.TODO(), roleBindingName, metav1.GetOptions{})
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
				err = fmt.Errorf("Failed to get rolebinding %s before timeout", roleBindingName)
				break Loop
			}
		}
	}

	return err
}

func pollLockFile(h *helper.H, namespace, operatorLockFile string) error {
	// GetConfigMap polls for a configMap with a timeout
	// to handle the case when a new cluster is up but the OLM has not yet
	// finished deploying the operator

	var err error

	// interval is the duration in seconds between polls
	// values here for humans
	interval := 30

	// convert time.Duration type
	timeoutDuration := time.Duration(viper.GetFloat64(config.Tests.PollingTimeout)) * time.Minute
	intervalDuration := time.Duration(interval) * time.Second

	start := time.Now()

Loop:
	for {
		_, err = h.Kube().CoreV1().ConfigMaps(namespace).Get(context.TODO(), operatorLockFile, metav1.GetOptions{})
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
				err = fmt.Errorf("Failed to get configMap %s before timeout", operatorLockFile)
				break Loop
			}
		}
	}

	return err
}

func pollDeployment(h *helper.H, namespace, deploymentName string) (*appsv1.Deployment, error) {
	// pollDeployment polls for a deployment with a timeout
	// to handle the case when a new cluster is up but the OLM has not yet
	// finished deploying the operator

	var err error
	var deployment *appsv1.Deployment

	// interval is the duration in seconds between polls
	// values here for humans
	interval := 5

	// convert time.Duration type
	timeoutDuration := time.Duration(viper.GetFloat64(config.Tests.PollingTimeout)) * time.Minute
	intervalDuration := time.Duration(interval) * time.Second

	start := time.Now()

Loop:
	for {
		deployment, err = h.Kube().AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
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
				err = fmt.Errorf("Failed to get %s Deployment before timeout", deploymentName)
				break Loop
			}
		}
	}

	return deployment, err
}

func getReplacesCSV(h *helper.H, subscriptionNS string, csvDisplayName string, catalogSvcName string) (string, error) {
	cmdTimeoutInSeconds := 60
	cmdTestTemplate, err := templates.LoadTemplate("registry/replaces.template")

	if err != nil {
		panic(fmt.Sprintf("error while loading registry-replaces addon: %v", err))
	}

	// This is extremely crude, but saves making multiple grpcurl queries to know what
	// channel the package is using
	clusterProvider, err := providers.ClusterProvider()
	environment := clusterProvider.Environment()
	var packageChannel string
	if strings.HasPrefix(environment, "prod") {
		packageChannel = "production"
	} else {
		packageChannel = "staging"
	}
	registrySvcPort := "50051"

	h.SetServiceAccount("system:serviceaccount:%s:cluster-admin")
	r := h.RunnerWithNoCommand()
	r.Name = fmt.Sprintf("csvq-%s", csvDisplayName)

	Expect(err).NotTo(HaveOccurred())
	values := struct {
		Name           string
		OutputDir      string
		Namespace      string
		PackageName    string
		PackageChannel string
		ServiceName    string
		ServicePort    string
		CA             string
		TokenFile      string
		Server         string
	}{
		Name:           r.Name,
		OutputDir:      runner.DefaultRunner.OutputDir,
		Namespace:      subscriptionNS,
		PackageName:    csvDisplayName,
		PackageChannel: packageChannel,
		ServiceName:    catalogSvcName,
		ServicePort:    registrySvcPort,
		Server:         "https://kubernetes.default",
		CA:             "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
		TokenFile:      "/var/run/secrets/kubernetes.io/serviceaccount/token",
	}

	registryQueryCmd, err := h.ConvertTemplateToString(cmdTestTemplate, values)
	Expect(err).NotTo(HaveOccurred())

	r.Cmd = registryQueryCmd

	// run tests
	stopCh := make(chan struct{})
	err = r.Run(cmdTimeoutInSeconds, stopCh)
	Expect(err).NotTo(HaveOccurred())

	// get results
	results, err := r.RetrieveResults()
	Expect(err).NotTo(HaveOccurred())

	var result map[string]interface{}
	err = json.Unmarshal(results["registry.json"], &result)
	Expect(err).NotTo(HaveOccurred(), "error unmarshalling json from registry gatherer")

	var csvresult map[string]interface{}
	err = json.Unmarshal([]byte(fmt.Sprintf("%v", result["csvJson"])), &csvresult)
	Expect(err).NotTo(HaveOccurred(), "error unmarshalling csv json from registry gatherer")

	replacesCsv, ok := csvresult["spec"].(map[string]interface{})["replaces"]
	Expect(ok).NotTo(BeFalse(), "cannot find 'replaces' clusterversion from registry gatherer")

	return fmt.Sprintf("%v", replacesCsv), nil
}

func CheckUpgrade(h *helper.H, subNamespace string, subName string, packageName string, regServiceName string) {
	checkUpgrade(h, subNamespace, subName, packageName, regServiceName)
}

func CheckPod(h *helper.H, namespace string, name string, gracePeriod int, maxAcceptedRestart int) {
	checkPod(h, namespace, name, gracePeriod, maxAcceptedRestart)
}

func PollDeployment(h *helper.H, namespace, deploymentName string) (*appsv1.Deployment, error) {
	return pollDeployment(h, namespace, deploymentName)
}
