package cloudingress

import (
	"context"
	"fmt"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/constants"
	"github.com/openshift/osde2e/pkg/common/helper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
	"k8s.io/apimachinery/pkg/util/wait"
)

var _ = ginkgo.Describe(constants.SuiteInforming+TestPrefix, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool("rosa.STS") {
			ginkgo.Skip("for now we skip this suite for STS")
		}
		if viper.GetString(config.CloudProvider.CloudProviderID) != "aws" {
			ginkgo.Skip("for now we only support aws provider")
		}
	})

	h := helper.New()
	testLBDeletion(h)
})

// getLBForService retrieves the loadbalancer name associated with a service of type LoadBalancer
func getLBForService(h *helper.H, namespace string, service string) (string, error) {
	svc, err := h.Kube().CoreV1().Services(namespace).Get(context.TODO(), service, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if svc.Spec.Type != "LoadBalancer" {
		return "", fmt.Errorf("service type is not LoadBalancer")
	}

	ingressList := svc.Status.LoadBalancer.Ingress
	if len(ingressList) == 0 {
		// the LB wasn't created yet
		return "", nil
	}
	return ingressList[0].Hostname[0:32], nil
}

// testLBDeletion deletes the loadbalancer of rh-api service and ensures that cloud-ingress-operator recreates it
func testLBDeletion(h *helper.H) {
	ginkgo.Context("rh-api-test", func() {
		ginkgo.It("Manually deleted LB should be recreated", func() {
			if viper.GetString(config.CloudProvider.CloudProviderID) == "aws" {
				awsAccessKey := viper.GetString("ocm.aws.accessKey")
				awsSecretKey := viper.GetString("ocm.aws.secretKey")
				awsRegion := viper.GetString(config.CloudProvider.Region)

				// getLoadBalancer name currently associated with rh-api service
				oldLBName, err := getLBForService(h, "openshift-kube-apiserver", "rh-api")
				Expect(err).NotTo(HaveOccurred())

				// delete the load balancer in aws
				awsSession, err := session.NewSession(aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, "")).WithRegion(awsRegion))
				Expect(err).NotTo(HaveOccurred())

				lb := elb.New(awsSession)
				input := &elb.DeleteLoadBalancerInput{
					LoadBalancerName: aws.String(oldLBName),
				}

				_, err = lb.DeleteLoadBalancer(input)
				Expect(err).NotTo(HaveOccurred())

				// wait for the new LB to be created
				err = wait.PollImmediate(15*time.Second, 5*time.Minute, func() (bool, error) {
					newLBName, err := getLBForService(h, "openshift-kube-apiserver", "rh-api")
					if err != nil || newLBName == "" {
						// either we couldn't retrieve the LB name, or it wasn't created yet
						return false, nil
					}
					if newLBName != oldLBName {
						// the LB was successfully recreated
						return true, nil
					}
					// the rh-api svc hasn't been deleted yet
					return false, nil
				})
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
}
