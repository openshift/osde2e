package cloudingress

import (
	"context"
	"fmt"
	"log"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/constants"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"k8s.io/apimachinery/pkg/util/wait"

	computev1 "google.golang.org/api/compute/v1"

	"google.golang.org/api/option"
)

var _ = ginkgo.Describe(constants.SuiteOperators+TestPrefix, func() {
	ginkgo.BeforeEach(func() {
		if viper.GetBool("rosa.STS") {
			ginkgo.Skip("Cluster is STS. For now we skip rh-api LB reconcile test for STS")
		}
		if viper.GetBool("ocm.ccs") != true {
			ginkgo.Skip("Cluster is non-CCS. For now we skip rh-api LB reconcile test for non-CCS.")
		}
	})

	h := helper.New()
	testLBDeletion(h)
})

// Get forwarding rule for rh-api load balancer in GCP
func getGCPForwardingRuleForIP(computeService *computev1.Service, oldLBIP string, project string, region string) (*computev1.ForwardingRule, error) {
	listCall := computeService.ForwardingRules.List(project, region)
	response, err := listCall.Do()
	var oldLB *computev1.ForwardingRule
	if err != nil {
		return nil, err
	}

	for _, lb := range response.Items {
		// This list of forwardingrules (LBs) includes any service LBs
		// for application routers so check the IP to identify
		// the rh-api LB.
		if lb.IPAddress == oldLBIP {
			oldLB = lb
		}
	}

	return oldLB, nil
}

// getLBForService retrieves the load balancer name or IP associated with a service of type LoadBalancer
func getLBForService(ctx context.Context, h *helper.H, namespace string, service string, idtype string) (string, error) {
	svc, err := h.Kube().CoreV1().Services(namespace).Get(ctx, service, metav1.GetOptions{})
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

	if idtype == "ip" {
		return ingressList[0].IP, nil
	}

	return ingressList[0].Hostname[0:32], nil
}

// deleteSecGroupReferencesToOrphans deletes any security group rules referencing the provided
// security group IDs (assumed to be those of security groups "orphaned" by LB deletion)
func deleteSecGroupReferencesToOrphans(ec2Svc *ec2.EC2, orphanSecGroupIds []*string) error {
	for _, orphanSecGroupId := range orphanSecGroupIds {
		// list all sec groups
		secGroupsAll, err := ec2Svc.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})
		if err != nil {
			return err
		}

		// now that we know which sec groups mention the orphan, we can modify them to remove
		// the referencing rules
		for _, secGroup := range secGroupsAll.SecurityGroups {
			// define an "IpPermissions" pattern that matches all rules referencing orphan
			orphanSecGroupIpPermissions := []*ec2.IpPermission{
				{
					IpProtocol:       aws.String("-1"), // Means "all protocols"
					UserIdGroupPairs: []*ec2.UserIdGroupPair{{GroupId: aws.String(*orphanSecGroupId)}},
				},
			}

			// delete all egress rules matching pattern
			_, err = ec2Svc.RevokeSecurityGroupEgress(&ec2.RevokeSecurityGroupEgressInput{
				GroupId:       aws.String(*secGroup.GroupId),
				IpPermissions: orphanSecGroupIpPermissions,
			})
			if err == nil {
				log.Printf("Removed egress rule referring to orphan from %s", *secGroup.GroupId)
			} else if err.(awserr.Error).Code() != "InvalidPermission.NotFound" {
				// since we're iterating over all security groups, RevokeSecurityGroup*gress
				// will often throw InvalidPermission; this is expected behavior. if a different
				// error arises, report it
				log.Printf("Encountered error while removing egress rule from %s: %s", *secGroup.GroupId, err)
			}

			// delete all ingress rules matching pattern
			_, err = ec2Svc.RevokeSecurityGroupIngress(&ec2.RevokeSecurityGroupIngressInput{
				GroupId:       aws.String(*secGroup.GroupId),
				IpPermissions: orphanSecGroupIpPermissions,
			})
			if err == nil {
				log.Printf("Removed ingress rule referring to orphan from %s", *secGroup.GroupId)
			} else if err.(awserr.Error).Code() != "InvalidPermission.NotFound" {
				log.Printf("Encountered error while removing ingress rule from %s: %s", *secGroup.GroupId, err)
			}
		}
	}
	return nil
}

// testLBDeletion deletes the load balancer of rh-api service and ensures that cloud-ingress-operator recreates it
func testLBDeletion(h *helper.H) {
	ginkgo.Context("rh-api-lb-test", func() {
		if viper.GetString(config.CloudProvider.CloudProviderID) == "aws" {
			util.GinkgoIt("manually deleted LB should be recreated in AWS", func(ctx context.Context) {
				awsAccessKey := viper.GetString("ocm.aws.accessKey")
				awsSecretKey := viper.GetString("ocm.aws.secretKey")
				awsRegion := viper.GetString(config.CloudProvider.Region)

				// getLoadBalancer name currently associated with rh-api service
				oldLBName, err := getLBForService(ctx, h, "openshift-kube-apiserver", "rh-api", "hostname")
				Expect(err).NotTo(HaveOccurred())
				log.Printf("Old LB name %s ", oldLBName)

				// delete the load balancer in aws
				awsSession, err := session.NewSession(aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, "")).WithRegion(awsRegion))
				Expect(err).NotTo(HaveOccurred())

				lb := elb.New(awsSession)
				input := &elb.DeleteLoadBalancerInput{
					LoadBalancerName: aws.String(oldLBName),
				}

				// must store security groups associated with LB so we can delete them
				oldLBDesc, err := lb.DescribeLoadBalancersWithContext(ctx, &elb.DescribeLoadBalancersInput{
					LoadBalancerNames: []*string{aws.String(oldLBName)},
				})
				Expect(err).NotTo(HaveOccurred())
				orphanSecGroupIds := oldLBDesc.LoadBalancerDescriptions[0].SecurityGroups

				_, err = lb.DeleteLoadBalancer(input)

				Expect(err).NotTo(HaveOccurred())
				log.Printf("Old LB deleted")

				// wait for the new LB to be created
				err = wait.PollImmediate(15*time.Second, 5*time.Minute, func() (bool, error) {
					newLBName, err := getLBForService(ctx, h, "openshift-kube-apiserver", "rh-api", "hostname")
					log.Printf("Looking for new LB")

					if err != nil || newLBName == "" {
						// either we couldn't retrieve the LB name, or it wasn't created yet
						log.Printf("LB not found yet")
						return false, nil
					}
					if newLBName != oldLBName {
						// the LB was successfully recreated
						log.Printf("New LB found. LB name: %s", newLBName)
						return true, nil
					}
					// the rh-api svc hasn't been deleted yet
					log.Printf("rh-api service not deleted yet")
					return false, nil
				})
				Expect(err).NotTo(HaveOccurred())

				// old LB's security groups ("orphans") will leak if not explicitly deleted
				// first, delete sec group rule references to the orphans
				ec2Svc := ec2.New(awsSession)
				log.Printf("Cleaning up references to security groups orphaned by old LB deletion")
				err = deleteSecGroupReferencesToOrphans(ec2Svc, orphanSecGroupIds)
				Expect(err).NotTo(HaveOccurred())

				// then delete the orphaned sec groups themselves
				for _, orphanSecGroupId := range orphanSecGroupIds {
					_, err := ec2Svc.DeleteSecurityGroupWithContext(ctx, &ec2.DeleteSecurityGroupInput{
						GroupId: aws.String(*orphanSecGroupId),
					})
					if err != nil {
						log.Printf("Failed to delete security group %s: %s", *orphanSecGroupId, err)
					} else {
						log.Printf("Deleted orphaned security group %s", *orphanSecGroupId)
					}
				}
			}, 600)
		}

		if viper.GetString(config.CloudProvider.CloudProviderID) == "gcp" {
			util.GinkgoIt("manually deleted LB should be recreated in GCP", func(ctx context.Context) {
				region := viper.GetString("cloudProvider.region")

				ginkgo.By("Getting rh-api IP")
				oldLBIP, err := getLBForService(ctx, h, "openshift-kube-apiserver", "rh-api", "ip")
				Expect(err).NotTo(HaveOccurred())
				log.Printf("old LB IP:  %s ", oldLBIP)

				ginkgo.By("Getting GCP creds")
				gcpCreds, status := h.GetGCPCreds(ctx)
				Expect(status).To(BeTrue())
				project := gcpCreds.ProjectID

				ginkgo.By("Initializing GCP compute service")
				computeService, err := computev1.NewService(ctx, option.WithCredentials(gcpCreds), option.WithScopes("https://www.googleapis.com/auth/compute"))
				Expect(err).NotTo(HaveOccurred())

				ginkgo.By("Getting GCP forwarding rule for rh-api")
				oldLB, err := getGCPForwardingRuleForIP(computeService, oldLBIP, project, region)
				Expect(err).NotTo(HaveOccurred())

				// There's no single command to delete a load balancer in GCP
				// Delete all GCP resources related to rh-api LB setup
				ginkgo.By("Deleting rh-api load balancer related resources in GCP")
				if oldLB == nil {
					log.Printf("GCP forwarding rule for rh-api does not exist; Skipping deletion ")
				} else {
					log.Printf("Old lb name:  %s ", oldLB.Name)
					_, err = computeService.ForwardingRules.Get(project, region, oldLB.Name).Do()
					if err != nil {
						log.Printf("GCP forwarding rule for rh-api not found! ")
					} else {
						ginkgo.By("Deleting GCP forwarding rule for rh-api")
						_, err = computeService.ForwardingRules.Delete(project, region, oldLB.Name).Do()
						if err != nil {
							log.Printf("Error deleting forwarding rule ")
						}
					}

					ginkgo.By("Deleting GCP backend service rule for rh-api")
					_, err = computeService.BackendServices.Get(project, oldLB.Name).Do()
					if err != nil {
						log.Printf("GCP backend service already deleted. ")
					} else {
						_, err = computeService.BackendServices.Delete(project, oldLB.Name).Do()
						if err != nil {
							log.Printf("Error deleting backend service ")
						}
					}

					ginkgo.By("Deleting GCP health check for rh-api ")
					_, err = computeService.HealthChecks.Get(project, oldLB.Name).Do()
					if err != nil {
						log.Printf("GCP health check already deleted ")
					} else {
						_, err = computeService.HealthChecks.Delete(project, oldLB.Name).Do()
						if err != nil {
							log.Printf("Error deleting health check ")
						}
					}

					ginkgo.By("Deleting GCP target pool for rh-api ")
					_, err = computeService.TargetPools.Get(project, region, oldLB.Name).Do()
					if err != nil {
						log.Printf("GCP target pool already deleted ")
					} else {
						_, err = computeService.TargetPools.Delete(project, region, oldLB.Name).Do()
						if err != nil {
							log.Printf("Error deleting target pool ")
						}
					}
				}

				ginkgo.By("Deleting GCP address for rh-api")
				_, err = computeService.Addresses.Get(project, region, oldLBIP).Do()
				if err != nil {
					log.Printf("GCP IP address already deleted ")
				} else {
					_, err = computeService.Addresses.Delete(project, region, oldLBIP).Do()
					if err != nil {
						log.Printf("Error deleting address ")
					}
				}

				newLBIP := ""
				// Getting the new LB from GCP
				err = wait.PollImmediate(15*time.Second, 10*time.Minute, func() (bool, error) {
					// Getting the newly created IP from rh-api service
					ginkgo.By("Getting new IP from rh-api service in OCM")

					newLBIP, err = getLBForService(ctx, h, "openshift-kube-apiserver", "rh-api", "ip")
					if (err != nil) || (newLBIP == "") || (newLBIP == oldLBIP) {
						log.Printf("New rh-api svc not created yet...")
						return false, nil
					} else {
						log.Printf("Found new rh-api svc! ")
						log.Printf("new lb IP: %s ", newLBIP)
						return true, nil
					}
				})
				Expect(err).NotTo(HaveOccurred())

				err = wait.PollImmediate(15*time.Second, 10*time.Minute, func() (bool, error) {
					ginkgo.By("Polling GCP to get new forwarding rule for rh-api")
					newLB, err := getGCPForwardingRuleForIP(computeService, newLBIP, project, region)
					if err != nil || newLB == nil {
						// Either we couldn't retrieve the LB, or it wasn't created yet
						log.Printf("New forwarding rule not found yet...")
						return false, nil
					}
					log.Printf("new lb name: %s ", newLB.Name)

					if newLB.Name != oldLB.Name {
						// A new LB was successfully recreated in GCP
						return true, nil
					}
					// rh-api lb hasn't been deleted yet
					log.Printf("Old forwarding rule not deleted yet...")
					return false, nil
				})
				Expect(err).NotTo(HaveOccurred())
			}, 600)
		}
	})
}
