package operators

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/openshift/osde2e/pkg/common/providers"
	"github.com/openshift/osde2e/pkg/common/runner"
	"github.com/openshift/osde2e/pkg/common/templates"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"

	operatorv1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

func checkClusterServiceVersion(h *helper.H, namespace, name string) {
	// Check that the operator clusterServiceVersion exists
	ginkgo.Context("clusterServiceVersion", func() {
		ginkgo.It("should exist", func() {
			csvs, err := pollCsvList(h, namespace, name)
			Expect(err).ToNot(HaveOccurred(), "failed fetching the clusterServiceVersions")
			Expect(csvs).NotTo(BeNil())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func checkConfigMapLockfile(h *helper.H, namespace, operatorLockFile string) {
	// Check that the operator configmap has been deployed
	ginkgo.Context("configmaps", func() {
		ginkgo.It("should exist", func() {
			// Wait for lockfile to signal operator is active
			err := pollLockFile(h, namespace, operatorLockFile)
			Expect(err).ToNot(HaveOccurred(), "failed fetching the configMap lockfile")
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func checkDeployment(h *helper.H, namespace string, name string, defaultDesiredReplicas int32) {
	// Check that the operator deployment exists in the operator namespace
	ginkgo.Context("deployment", func() {
		ginkgo.It("should exist", func() {
			deployment, err := pollDeployment(h, namespace, name)
			Expect(err).ToNot(HaveOccurred(), "failed fetching deployment")
			Expect(deployment).NotTo(BeNil(), "deployment is nil")
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
		ginkgo.It("should have all desired replicas ready", func() {
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

func checkClusterRoles(h *helper.H, clusterRoles []string, matchPrefix bool) {
	// Check that the clusterRoles exist
	ginkgo.Context("clusterRoles", func() {
		ginkgo.It("should exist", func() {
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
		ginkgo.It("should exist", func() {
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
		ginkgo.It("should exist", func() {
			for _, roleName := range roles {
				_, err := h.Kube().RbacV1().Roles(namespace).Get(context.TODO(), roleName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "failed to get role %v\n", roleName)
			}
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})

}

func checkRoleBindings(h *helper.H, namespace string, roleBindings []string) {
	// Check that deployed rolebindings exist
	ginkgo.Context("roleBindings", func() {
		ginkgo.It("should exist", func() {
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
		ginkgo.It("should exist", func() {
			for _, secretName := range secrets {
				_, err := h.Kube().CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred(), "failed to get secret %v\n", secretName)
			}
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
}

func approveInstallPlan(h *helper.H, csv string, sub *operatorv1.Subscription) error {

	// find the install plan associated with the CSV
	var ip operatorv1.InstallPlan
	err := wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		ips, err := h.Operator().OperatorsV1alpha1().InstallPlans(sub.Namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		for _, i := range ips.Items {
			for _, c := range i.Spec.ClusterServiceVersionNames {
				if c == csv {
					ip = i
					return true, nil
				}
			}
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("no install plan found for csv %s", csv)
	}

	log.Printf("Found install plan %s for csv %s", ip.Name, csv)
	if ip.Spec.Approved == true {
		log.Printf("Install plan %s already approved", ip.Name)
		return nil
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		ipToUpdate, err := h.Operator().OperatorsV1alpha1().InstallPlans(sub.Namespace).Get(context.TODO(), ip.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		// Flag the install plan as approved
		ipToUpdate.Spec.Approved = true
		_, err = h.Operator().OperatorsV1alpha1().InstallPlans(ipToUpdate.Namespace).Update(context.TODO(), ipToUpdate, metav1.UpdateOptions{})
		if err != nil {
			log.Printf("Could not approve install plan %s", ipToUpdate.Name)
			return err
		}
		log.Printf("Approved install plan %s", ipToUpdate.Name)
		return nil
	})
	return err
}

func ensureCSVIsInstalled(h *helper.H, csvName string, sub *operatorv1.Subscription) error {
	err := wait.PollImmediate(5*time.Second, 15*time.Minute, func() (bool, error) {
		// Ensure the install plan for the CSV is approved
		err := approveInstallPlan(h, csvName, sub)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Failed approving InstallPlan for CSV %s", csvName))
		log.Printf("Approved install plan for sub %s, now checking CSV %s is installed", sub.Name, csvName)

		csv, err := h.Operator().OperatorsV1alpha1().ClusterServiceVersions(sub.Namespace).Get(context.TODO(), csvName, metav1.GetOptions{})
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
		return err
	}

	return nil
}

func checkUpgrade(h *helper.H, subNamespace string, subName string, packageName string, regServiceName string) {

	ginkgo.Context("Operator Upgrade", func() {
		ginkgo.It("should upgrade from the replaced version", func() {

			// Get the CSV we're currently installed with
			sub, err := h.Operator().OperatorsV1alpha1().Subscriptions(subNamespace).Get(context.TODO(), subName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed trying to get Subscription %s in %s namespace", subName, subNamespace))
			startingCSV := sub.Status.CurrentCSV

			// Get the N-1 version of the CSV to test an upgrade from
			previousCSV, err := getReplacesCSV(h, subNamespace, packageName, regServiceName)
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed trying to get previous CSV for Subscription %s in %s namespace", subName, subNamespace))

			log.Printf("Reverting to package %v from %v to test upgrade of %v", previousCSV, startingCSV, subName)

			// Delete current Operator installation
			err = h.Operator().OperatorsV1alpha1().Subscriptions(subNamespace).Delete(context.TODO(), subName, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed trying to delete Subscription %s", subName))
			log.Printf("Removed subscription %s", subName)

			err = h.Operator().OperatorsV1alpha1().ClusterServiceVersions(subNamespace).Delete(context.TODO(), startingCSV, metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed trying to delete ClusterServiceVersion %s", startingCSV))
			log.Printf("Removed csv %s", startingCSV)

			Eventually(func() bool {
				_, err := h.Operator().OperatorsV1alpha1().InstallPlans(subNamespace).Get(context.TODO(), sub.Status.Install.Name, metav1.GetOptions{})
				return apierrors.IsNotFound(err)
			}, 5*time.Minute, 10*time.Second).Should(BeTrue(), "installplan never garbage collected")
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
					InstallPlanApproval:    operatorv1.ApprovalManual,
					StartingCSV:            previousCSV,
				},
			}, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("failed trying to create Subscription %s", subName))

			log.Printf("Created replacement subscription %s", subName)

			// Approve and manually verify the first installation to previousCSV
			err = ensureCSVIsInstalled(h, previousCSV, sub)
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("CSV %s did not install successfully", previousCSV))

			log.Printf("Verified installation of previous CSV %s", previousCSV)

			// Update the Subscription to apply Automatic updates from now on in order to reach currentCSV
			err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				sub, err = h.Operator().OperatorsV1alpha1().Subscriptions(subNamespace).Get(context.TODO(), subName, metav1.GetOptions{})
				if err != nil {
					return err
				}
				sub.Spec.InstallPlanApproval = operatorv1.ApprovalAutomatic
				sub, err = h.Operator().OperatorsV1alpha1().Subscriptions(subNamespace).Update(context.TODO(), sub, metav1.UpdateOptions{})
				if err != nil {
					return err
				}
				return nil
			})
			Expect(err).To(BeNil(), fmt.Sprintf("unable to set Subscription to Automatic"))

			log.Printf("Set subscription %s to Automatic approval", subName)

			// The previous CSV is now installed and a new InstallPlan is also created ready for approval to upgrade to startingCSV
			// Approve and verify install to startingCSV
			err = ensureCSVIsInstalled(h, startingCSV, sub)
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("CSV %s did not install successfully", startingCSV))
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
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
	timeoutDuration := time.Duration(viper.GetFloat64(config.Tests.PollingTimeout)) * time.Minute
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

func pollCsvList(h *helper.H, namespace, csvDisplayName string) (*operatorv1.ClusterServiceVersionList, error) {
	// pollCsvList polls for clusterServiceVersions with a timeout
	// to handle the case when a new cluster is up but the OLM has not yet
	// finished deploying the operator

	var err error
	var csvList *operatorv1.ClusterServiceVersionList

	// interval is the duration in seconds between polls
	// values here for humans
	interval := 5

	// convert time.Duration type
	timeoutDuration := time.Duration(viper.GetFloat64(config.Tests.PollingTimeout)) * time.Minute
	intervalDuration := time.Duration(interval) * time.Second

	start := time.Now()

Loop:
	for {
		csvList, err = h.Operator().OperatorsV1alpha1().ClusterServiceVersions(namespace).List(context.TODO(), metav1.ListOptions{})
		for _, csv := range csvList.Items {
			switch {
			case csvDisplayName == csv.Spec.DisplayName:
				// Success
				err = nil
			default:
				err = fmt.Errorf("No matching clusterServiceVersion in CSV List")
			}
		}
		elapsed := time.Since(start)

		switch {
		case err == nil:
			// Success
			break Loop
		case strings.Contains(err.Error(), "forbidden"):
			return nil, err
		default:
			if elapsed < timeoutDuration {
				log.Printf("Waiting %v for %s clusterServiceVersion to exist", (timeoutDuration - elapsed), csvDisplayName)
				time.Sleep(intervalDuration)
			} else {
				csvList = nil
				err = fmt.Errorf("Failed to get %s clusterServiceVersion before timeout", csvDisplayName)
				break Loop
			}
		}
	}

	return csvList, err
}

func getReplacesCSV(h *helper.H, subscriptionNS string, csvDisplayName string, catalogSvcName string) (string, error) {
	cmdTimeoutInSeconds := 60
	cmdTestTemplate, err := templates.LoadTemplate("/assets/registry/replaces.template")

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
