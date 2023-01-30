package operators

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/util"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"

	routev1 "github.com/openshift/api/route/v1"

	operatorv1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	clientreg "github.com/operator-framework/operator-registry/pkg/client"

	"github.com/adamliesko/retry"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
)

func checkClusterServiceVersion(h *helper.H, namespace, name string) {
	// Check that the operator clusterServiceVersion exists
	ginkgo.Context(fmt.Sprintf("clusterServiceVersion %s/%s", namespace, name), func() {
		util.GinkgoIt("should be present and in succeeded state", func(ctx context.Context) {
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
		}, viper.GetFloat64(config.Tests.PollingTimeout))
	})
}

func checkConfigMapLockfile(h *helper.H, namespace, operatorLockFile string) {
	// Check that the operator configmap has been deployed
	ginkgo.Context("configmaps", func() {
		util.GinkgoIt("should exist", func(ctx context.Context) {
			// Wait for lockfile to signal operator is active
			err := pollLockFile(ctx, h, namespace, operatorLockFile)
			Expect(err).ToNot(HaveOccurred(), "failed fetching the configMap lockfile")
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func checkDeployment(h *helper.H, namespace string, name string, defaultDesiredReplicas int32) {
	// Check that the operator deployment exists in the operator namespace
	ginkgo.Context("deployment", func() {
		util.GinkgoIt("should exist", func(ctx context.Context) {
			deployment, err := pollDeployment(ctx, h, namespace, name)
			Expect(err).ToNot(HaveOccurred(), "failed fetching deployment")
			Expect(deployment).NotTo(BeNil(), "deployment is nil")
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
		util.GinkgoIt("should have all desired replicas ready", func(ctx context.Context) {
			deployment, err := pollDeployment(ctx, h, namespace, name)
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
		util.GinkgoIt(fmt.Sprintf("should have %v or less restart(s)", maxAcceptedRestart), func(ctx context.Context) {
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
		}, float64(gracePeriod)+viper.GetFloat64(config.Tests.PollingTimeout))
	})
}

func checkServiceAccounts(h *helper.H, operatorNamespace string, serviceAccounts []string) {
	// Check that deployed serviceAccounts exist
	ginkgo.Context("serviceAccounts", func() {
		util.GinkgoIt("should exist", func(ctx context.Context) {
			for _, serviceAccountName := range serviceAccounts {
				_, err := h.Kube().CoreV1().ServiceAccounts(operatorNamespace).Get(ctx, serviceAccountName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "failed to get serviceAccount %v\n", serviceAccountName)
			}
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func checkClusterRoles(h *helper.H, clusterRoles []string, matchPrefix bool) {
	// Check that the clusterRoles exist
	ginkgo.Context("clusterRoles", func() {
		util.GinkgoIt("should exist", func(ctx context.Context) {
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
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func checkClusterRoleBindings(h *helper.H, clusterRoleBindings []string, matchPrefix bool) {
	// Check that the clusterRoles exist
	ginkgo.Context("clusterRoleBindings", func() {
		util.GinkgoIt("should exist", func(ctx context.Context) {
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
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func checkRole(h *helper.H, namespace string, roles []string) {
	// Check that deployed roles exist
	ginkgo.Context("roles", func() {
		util.GinkgoIt("should exist", func(ctx context.Context) {
			for _, roleName := range roles {
				_, err := h.Kube().RbacV1().Roles(namespace).Get(ctx, roleName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "failed to get role %v\n", roleName)
			}
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func checkRolesWithNamePrefix(h *helper.H, namespace string, prefix string, count int) {
	ginkgo.Context("roles with prefix", func() {
		util.GinkgoIt("should exist", func(ctx context.Context) {
			Eventually(func() int {
				rolesList, err := h.Kube().RbacV1().Roles(namespace).List(ctx, metav1.ListOptions{})
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
		util.GinkgoIt("should exist", func(ctx context.Context) {
			Eventually(func() int {
				roleBindings, err := h.Kube().RbacV1().RoleBindings(namespace).List(ctx, metav1.ListOptions{})
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
		util.GinkgoIt("should exist", func(ctx context.Context) {
			for _, roleBindingName := range roleBindings {
				err := pollRoleBinding(ctx, h, namespace, roleBindingName)
				Expect(err).NotTo(HaveOccurred(), "failed to get roleBinding %v\n", roleBindingName)
			}
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

// nolint
func checkSecrets(h *helper.H, namespace string, secrets []string) {
	// Check that deployed secrets exist
	ginkgo.Context("secrets", func() {
		util.GinkgoIt("should exist", func(ctx context.Context) {
			for _, secretName := range secrets {
				_, err := h.Kube().CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "failed to get secret %v\n", secretName)
			}
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func PerformUpgrade(ctx context.Context, h *helper.H, subNamespace string, subName string, packageName string, regServiceName string,
	installPlanPollCount time.Duration, upgradePollCount time.Duration,
) (string, error) {
	installPlanPollingDuration := installPlanPollCount * time.Minute
	upgradePollingDuration := upgradePollCount * time.Minute

	var latestCSV string
	var sub *operatorv1.Subscription
	var err error

	// The subscription must first exist on the cluster
	sub, err = h.Operator().OperatorsV1alpha1().Subscriptions(subNamespace).Get(ctx, subName, metav1.GetOptions{})
	if err != nil {
		return fmt.Sprintf("subscription %s not found", subName), err
	}

	// Get the CSV we're currently installed with
	installedCSVs, err := h.Operator().OperatorsV1alpha1().ClusterServiceVersions(subNamespace).List(ctx, metav1.ListOptions{})
	for _, csv := range installedCSVs.Items {
		if csv.Spec.DisplayName == packageName && csv.Status.Phase == operatorv1.CSVPhaseSucceeded {
			latestCSV = csv.Name
		}
	}

	// If we couldn't find a Succeeded CSV, then the operator is likely not even installed
	if len(latestCSV) == 0 {
		return fmt.Sprintf("no successfully installed CSV found for subscription %s", subName), err
	}

	// Get the N-1 version of the CSV to test an upgrade from
	previousCSV, err := getReplacesCSV(ctx, h, subNamespace, packageName, regServiceName)
	if err != nil {
		return fmt.Sprintf("failed trying to get previous CSV for Subscription %s in %s namespace", subName, subNamespace), err
	}

	log.Printf("Reverting to package %v from %v to test upgrade of %v", previousCSV, latestCSV, subName)

	// Delete current Operator installation
	err = h.Operator().OperatorsV1alpha1().Subscriptions(subNamespace).Delete(ctx, subName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Sprintf("failed trying to delete Subscription %s", subName), err
	}
	log.Printf("Removed subscription %s", subName)

	err = h.Operator().OperatorsV1alpha1().ClusterServiceVersions(subNamespace).Delete(ctx, latestCSV, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Sprintf("failed trying to delete ClusterServiceVersion %s", latestCSV), err
	}
	log.Printf("Removed csv %s", latestCSV)

	err = wait.PollImmediate(10*time.Second, installPlanPollingDuration, func() (bool, error) {
		ips, err := h.Operator().OperatorsV1alpha1().InstallPlans(subNamespace).List(ctx, metav1.ListOptions{})
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
	if err != nil {
		return "installplan never garbage collected", err
	}
	log.Printf("Verified installplan removal")

	// Create subscription to the previous version
	sub, err = h.Operator().OperatorsV1alpha1().Subscriptions(subNamespace).Create(ctx, &operatorv1.Subscription{
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
	if err != nil {
		return fmt.Sprintf("failed trying to create Subscription %s", subName), err
	}

	log.Printf("Created replacement subscription %s with starting CSV %s", subName, previousCSV)

	// Wait for the operator to arrive back on its latest CSV
	err = wait.PollImmediate(5*time.Second, upgradePollingDuration, func() (bool, error) {
		csv, err := h.Operator().OperatorsV1alpha1().ClusterServiceVersions(sub.Namespace).Get(ctx, latestCSV, metav1.GetOptions{})
		if err != nil && !kerror.IsNotFound(err) {
			log.Printf("Returning err %v", err)
			return false, err
		}
		if csv.Status.Phase == operatorv1.CSVPhaseSucceeded {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return fmt.Sprintf("CSV %s did not eventually install successfully", latestCSV), err
	}

	// Lastly, verify that the Subscription correctly reflects that the CSV is installed.
	err = wait.PollImmediate(5*time.Second, 2*time.Minute, func() (bool, error) {
		sub, err = h.Operator().OperatorsV1alpha1().Subscriptions(subNamespace).Get(ctx, subName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		currentCSV := sub.Status.CurrentCSV
		if currentCSV == latestCSV {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return fmt.Sprintf("subscription %s status is not reflecting that csv %s is installed", subName, latestCSV), err
	}

	return "", nil
}

func checkUpgrade(h *helper.H, subNamespace string, subName string, packageName string, regServiceName string) {
	ginkgo.Context("Operator Upgrade", func() {
		ginkgo.It("should upgrade from the replaced version", func(ctx context.Context) {
			errorMsg, err := PerformUpgrade(ctx, h, subNamespace, subName, packageName, regServiceName, 5, 15)
			Expect(err).NotTo(HaveOccurred(), errorMsg)
		})
	})
}

func checkService(h *helper.H, namespace string, name string, port int) {
	pollTimeout := viper.GetFloat64(config.Tests.PollingTimeout)
	ginkgo.Context("service", func() {
		util.GinkgoIt(
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
			pollTimeout,
		)
	})
}

func pollClusterRoleBinding(ctx context.Context, h *helper.H, clusterRoleBindingName string) error {
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
		_, err = h.Kube().RbacV1().ClusterRoleBindings().Get(ctx, clusterRoleBindingName, metav1.GetOptions{})
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

func pollRoleBinding(ctx context.Context, h *helper.H, projectName string, roleBindingName string) error {
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
				err = fmt.Errorf("Failed to get rolebinding %s before timeout", roleBindingName)
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
	timeoutDuration := time.Duration(viper.GetFloat64(config.Tests.PollingTimeout)) * time.Minute
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
				err = fmt.Errorf("Failed to get configMap %s before timeout", operatorLockFile)
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
	timeoutDuration := time.Duration(viper.GetFloat64(config.Tests.PollingTimeout)) * time.Minute
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
				err = fmt.Errorf("Failed to get %s Deployment before timeout", deploymentName)
				break Loop
			}
		}
	}

	return deployment, err
}

func getReplacesCSV(ctx context.Context, h *helper.H, subscriptionNS string, csvDisplayName string, catalogSvcName string) (string, error) {
	// Build the service that we'll use to GRPC-query the catalog
	svc, err := buildQueryableCatalogService(ctx, h, catalogSvcName, subscriptionNS)
	if err != nil {
		return "", err
	}
	createdSvc, err := h.Kube().CoreV1().Services(subscriptionNS).Create(ctx, svc, metav1.CreateOptions{})
	if err != nil && kerror.IsAlreadyExists(err) {
		return "", err
	}
	defer func() {
		// Clean up the service afterwards
		h.Kube().CoreV1().Services(subscriptionNS).Delete(ctx, createdSvc.GetName(), metav1.DeleteOptions{})
	}()
	log.Printf("Created service to grpc query for catalog data: %v/%v", createdSvc.Namespace, createdSvc.Name)

	// Build the route that we'll use to access the service by
	route := buildGRPCRoute(createdSvc.Name, subscriptionNS)
	createdRoute, err := h.Route().RouteV1().Routes(subscriptionNS).Create(ctx, route, metav1.CreateOptions{})
	if err != nil && kerror.IsAlreadyExists(err) {
		return "", err
	}
	defer func() {
		// Clean up the route afterwards
		h.Route().RouteV1().Routes(subscriptionNS).Delete(ctx, createdRoute.GetName(), metav1.DeleteOptions{})
	}()
	log.Printf("Created route to grpc query for catalog data: %v/%v", createdRoute.Namespace, createdRoute.Name)

	// We need to wait a little bit for the route to settle, this is a bit of an arbitrary sleep time..
	time.Sleep(5 * time.Second)

	// Build up the GRPC client
	var tlsConf tls.Config
	tlsConf.InsecureSkipVerify = true
	creds := credentials.NewTLS(&tlsConf)
	// We need to allow http/2 in order to GRPC query the catalog
	tlsConf.NextProtos = []string{"h2"}
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	grpcConn, err := grpc.Dial(fmt.Sprintf("%v:%v", createdRoute.Spec.Host, 443), dialOpts...)
	if err != nil {
		return "", err
	}
	rc := clientreg.NewClientFromConn(grpcConn)

	// Map the cluster environment to the catalog environment (prod->production, stage&int both use staging)
	clusterProvider, err := providers.ClusterProvider()
	if err != nil {
		return "", err
	}
	environment := clusterProvider.Environment()
	var packageChannel string
	if strings.HasPrefix(environment, "prod") {
		packageChannel = "production"
	} else {
		packageChannel = "staging"
	}

	csv := &operatorv1.ClusterServiceVersion{}

	r := retry.New(retry.Sleep(5), retry.Tries(6), retry.Recover())
	err = r.Do(func() error {
		bundle, err := rc.GetBundleInPackageChannel(ctx, csvDisplayName, packageChannel)
		if err != nil {
			return err
		}

		err = json.Unmarshal([]byte(bundle.GetCsvJson()), &csv)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// Return the 'replaces'
	return csv.Spec.Replaces, nil
}

// buildQueryableCatalogService constructs a gRPC over HTTP/2 clone of the specified
// service/namespace, specifically used for issuing gRPC queries of operator catalogs
func buildQueryableCatalogService(ctx context.Context, h *helper.H, serviceName string, serviceNamespace string) (*corev1.Service, error) {
	origService, err := h.Kube().CoreV1().Services(serviceNamespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	newSpec := origService.Spec.DeepCopy()
	newSpec.ClusterIP = ""
	newSpec.ClusterIPs = make([]string, 0)
	appProtocol := "h2c"
	newSpec.Ports = []corev1.ServicePort{
		{
			Name:        "grpc",
			Protocol:    "TCP",
			AppProtocol: &appProtocol,
			Port:        50051,
			TargetPort:  intstr.FromInt(50051),
		},
	}
	newService := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      origService.Name + "-e2e",
			Namespace: origService.Namespace,
		},
		Spec: *newSpec,
	}

	return &newService, nil
}

// buildGRPCRoute exposes the specified service/namespace via a gRPC route,
// specifically used to be able to issue gRPC queries to operator catalogs.
func buildGRPCRoute(service string, namespace string) *routev1.Route {
	route := routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "grpc",
			Namespace: namespace,
		},
		Spec: routev1.RouteSpec{
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("grpc"),
			},
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: service,
			},
			TLS: &routev1.TLSConfig{
				Termination: "edge",
			},
		},
	}

	return &route
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
