package cloudingress

import (
	compute "cloud.google.com/go/compute/apiv1"
	"context"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
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

	"golang.org/x/oauth2/google"
	computev1 "google.golang.org/api/compute/v1"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"

	"google.golang.org/api/option"
)

var _ = ginkgo.Describe(constants.SuiteInforming+TestPrefix, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool("rosa.STS") {
			ginkgo.Skip("for now we skip this suite for STS")
		}
	})

	h := helper.New()
	testLBDeletion(h)
})

// getLBForService retrieves the loadbalancer name associated with a service of type LoadBalancer
func getLBForService(h *helper.H, namespace string, service string, idtype string) (string, error) {
	svc, err := h.Kube().CoreV1().Services(namespace).Get(context.TODO(), service, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	//debugging
	fmt.Printf("%s", "service dump: ")
	spew.Dump(svc)
	if svc.Spec.Type != "LoadBalancer" {
		return "", fmt.Errorf("service type is not LoadBalancer")
	}

	ingressList := svc.Status.LoadBalancer.Ingress
	if len(ingressList) == 0 {
		// the LB wasn't created yet
		return "", nil
	}
	if idtype == "ip" {
		return ingressList[0].IP, nil
	}
	return ingressList[0].Hostname[0:32], nil

}

// testLBDeletion deletes the loadbalancer of rh-api service and ensures that cloud-ingress-operator recreates it
func testLBDeletion(h *helper.H) {
	ginkgo.Context("rh-api-lb-test", func() {
		ginkgo.It("Manually deleted LB should be recreated", func() {
			if viper.GetString(config.CloudProvider.CloudProviderID) == "aws" {
				awsAccessKey := viper.GetString("ocm.aws.accessKey")
				awsSecretKey := viper.GetString("ocm.aws.secretKey")
				awsRegion := viper.GetString(config.CloudProvider.Region)

				// getLoadBalancer name currently associated with rh-api service
				oldLBName, err := getLBForService(h, "openshift-kube-apiserver", "rh-api", "hostname")
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
					newLBName, err := getLBForService(h, "openshift-kube-apiserver", "rh-api", "hostname")
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
		if viper.GetString(config.CloudProvider.CloudProviderID) == "gcp" {
			ginkgo.It("LB should be recreated in GCP", func() {
				gcpCredsJson := viper.Get("ocm.gcp.credsJSON")
				project := viper.GetString("ocm.gcp.projectID")
				region := viper.GetString("cloudProvider.region")
				ctx := context.Background()
				c, _ := compute.NewForwardingRulesRESTClient(ctx)
				oldLBIP, err := getLBForService(h, "openshift-kube-apiserver", "rh-api", "ip")
				fmt.Printf("oldLBIP:  %s", oldLBIP)

				credsBytes, err := json.Marshal(gcpCredsJson)
				credentials, err := google.CredentialsFromJSON(
					ctx, credsBytes,
					computev1.ComputeScope)
				computeService, err := computev1.NewService(ctx, option.WithCredentials(credentials))

				// Delete LB in GCP
				oldLBName := "OldName"
				filtertext := "IPAddress = " + oldLBIP
				req := &computepb.AggregatedListForwardingRulesRequest{
					Filter:  &filtertext,
					Project: project,
				}
				it := c.AggregatedList(ctx, req)

				//debugging
				spew.Dump(it)
				//Find old LB Name in GCP and delete it
				//Use the first result
				resp, err := it.Next()
				//debugging
				spew.Dump(resp)

				oldLBName = *resp.Value.ForwardingRules[0].Name
				_, err = computeService.ForwardingRules.Delete(project, region, oldLBName).Do()

				// wait for the new LB to be created
				err = wait.PollImmediate(15*time.Second, 5*time.Minute, func() (bool, error) {
					newLBName := "newLBName"
					newLBIP, _ := getLBForService(h, "openshift-kube-apiserver", "rh-api", "ip")
					fmt.Printf("newLBIP:  %s", newLBIP)
					filtertext := "IPAddress = " + newLBIP

					req := &computepb.AggregatedListForwardingRulesRequest{
						Filter:  &filtertext,
						Project: project,
					}
					it := c.AggregatedList(ctx, req)
					//debugging
					spew.Dump(it)

					//Find new LB Name and compare with old
					//Use the first result
					resp, err := it.Next()
					//debugging
					spew.Dump(resp)
					newLBName = *resp.Value.ForwardingRules[0].Name
					//debugging
					spew.Dump(newLBName)
					if err != nil || newLBName == "" {
						// either we couldn't retrieve the LB name, or it wasn't created yet
						return false, nil
					}
					if newLBName != oldLBName {
						// a new LB was successfully recreated
						return true, nil
					}
					// the rh-api svc hasn't been deleted yet
					return false, nil
				})
				Expect(err).NotTo(HaveOccurred())
			})
		}
	})
}
