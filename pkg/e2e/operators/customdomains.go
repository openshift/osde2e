package operators

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"

	routev1 "github.com/openshift/api/route/v1"
	customdomainv1alpha1 "github.com/openshift/custom-domains-operator/pkg/apis/customdomain/v1alpha1"

	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

const (
	customDomainsOperatorTestName = "[Suite: operators] [OSD] Custom Domains Operator"
	pollInterval                  = 15 * time.Second
	defaultTimeout                = 20 * time.Minute
	endpointTimeout               = 20 * time.Minute
	endpointStabilityTimeout      = 20 * time.Minute
	dnsTimeout                    = 5 * time.Minute
	minConsecutiveResolves        = 3
)

func init() {
	alert.RegisterGinkgoAlert(customDomainsOperatorTestName, "SD-SREP", "Dustin Row", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(customDomainsOperatorTestName, func() {
	ginkgo.Context("Should allow dedicated-admins to create domains", func() {
		var (
			err error
			h   = helper.New()

			testInstanceName = "test-customdomain-" + time.Now().Format("20060102-150405-") + fmt.Sprint(time.Now().Nanosecond()/1000000) + "-" + fmt.Sprint(ginkgo.GinkgoParallelNode())
			testDomain       *customdomainv1alpha1.CustomDomain
		)

		ginkgo.BeforeEach(func() {
			ginkgo.By("Logging in as a dedicated-admin")
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"dedicated-admins",
				},
			})

			ginkgo.By("Creating an ssl certificate and tls secret in OSD")
			customDomainRSAKey, err := rsa.GenerateKey(rand.Reader, 4096)
			Expect(err).ToNot(HaveOccurred())

			testDomainName := testInstanceName + ".io"
			customDomainX509Template := &x509.Certificate{
				SerialNumber: big.NewInt(1),
				Subject: pkix.Name{
					Organization:       []string{"Red Hat, Inc"},
					OrganizationalUnit: []string{"Openshift Dedicated End-to-End Testing"},
				},
				DNSNames:              []string{"*." + testDomainName},
				NotBefore:             time.Now(),
				NotAfter:              time.Now().AddDate(0, 0, 1),
				IsCA:                  true,
				BasicConstraintsValid: true,
			}

			customDomainX509, err := x509.CreateCertificate(rand.Reader, customDomainX509Template, customDomainX509Template, &customDomainRSAKey.PublicKey, customDomainRSAKey)
			Expect(err).ToNot(HaveOccurred())

			secretData := make(map[string][]byte)
			secretData["tls.key"] = pem.EncodeToMemory(&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(customDomainRSAKey),
			})
			secretData["tls.crt"] = pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: customDomainX509,
			})

			// Create TLS Secret for CustomDomain
			testDomainSecret, err := h.Kube().CoreV1().Secrets(h.CurrentProject()).Create(context.TODO(), &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: testInstanceName + "-secret",
				},
				Type: corev1.SecretTypeTLS,
				Data: secretData,
			}, metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred())

			ginkgo.By("Creating a CustomDomain CR from the tls secret")
			testDomain = &customdomainv1alpha1.CustomDomain{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CustomDomain",
					APIVersion: "managed.openshift.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: testInstanceName,
				},
				Spec: customdomainv1alpha1.CustomDomainSpec{
					Domain: testDomainName,
					Certificate: corev1.SecretReference{
						Name:      testDomainSecret.GetName(),
						Namespace: testDomainSecret.GetNamespace(),
					},
				},
			}

			// Finally submit the CustomDomain CR for creation
			_, err = h.Cfg().ConfigV1().RESTClient().Post().
				AbsPath("/apis/managed.openshift.io/v1alpha1").
				Resource("customdomains").
				Name(testDomain.GetName()).
				Body(testDomain).
				DoRaw(context.TODO())
			Expect(err).ToNot(HaveOccurred())

			// Wait for CustomDomain CR Endpoint to be ready
			wait.PollImmediate(pollInterval, endpointTimeout, func() (bool, error) {
				byteResult, err := h.Cfg().ConfigV1().RESTClient().Get().
					AbsPath("/apis/managed.openshift.io/v1alpha1").
					Resource("customdomains").
					Name(testInstanceName).
					DoRaw(context.TODO())
				if err != nil {
					return false, err
				}
				err = json.Unmarshal(byteResult, testDomain)
				if err != nil {
					return false, err
				}

				if testDomain.Status.State == "Ready" {
					return true, err
				}
				return false, err
			})
			Expect(err).ToNot(HaveOccurred(), "Time out or error waiting for customdomain '"+testDomain.GetName()+"' to become ready.")
			Expect(string(testDomain.Status.State)).To(Equal("Ready"), "Customdomain may be unstable (.status.state is not 'Ready' anymore)")
			Expect(string(testDomain.Status.Endpoint)).ToNot(Equal(""), "Customdomain's .status.endpoint field empty when .status.state field is 'Ready'")

			// Customdomain endpoints have been known to:
			// 1) not resolve initially and/or
			// 2) resolve once, then fail to resolve for a time after
			// To ensure the endpoint is ready & stable, check that it resolves successfully several times before continuing
			var consecutiveResolves int
			wait.PollImmediate(pollInterval, endpointStabilityTimeout, func() (bool, error) {
				endpointIPs, err := net.LookupHost(testDomain.Status.Endpoint)
				if len(endpointIPs) == 0 {
					consecutiveResolves = 0
					return false, nil
				}
				if err != nil {
					return false, err
				}
				if consecutiveResolves < minConsecutiveResolves-1 {
					consecutiveResolves++
					return false, err
				}
				return true, err
			})
			Expect(err).ToNot(HaveOccurred(), "Time out or error waiting for customdomain endpoint '"+testDomain.Status.Endpoint+"' to resolve.")
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		// Now that the endpoint is stable, make sure it's resolvable and usable.
		ginkgo.It("Should be resolvable by external services", func() {
			ginkgo.By("Logging in as a dedicated-admin")

			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"dedicated-admins",
				},
			})

			ginkgo.By("Creating a new app and exposing it via an Openshift route")
			testAppReplicas := int32(1)
			testApp, err := h.Kube().AppsV1().Deployments(h.CurrentProject()).Create(context.TODO(), &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: testInstanceName,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &testAppReplicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"deployment": testInstanceName},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"deployment": testInstanceName},
						}, Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Name: "hello-openshift",
									Image: "docker.io/openshift/hello-openshift",
									Ports: []corev1.ContainerPort{
										{
											ContainerPort: 8080,
										},
										{
											ContainerPort: 8888,
										},
									},
								},
							},
						},
					},
				},
			},
				metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred())

			// Ensure the "hello-world" app is created
			wait.PollImmediate(pollInterval, defaultTimeout, func() (bool, error) {
				testApp, err = h.Kube().AppsV1().Deployments(h.CurrentProject()).Get(context.TODO(), testApp.GetName(), metav1.GetOptions{})
				if err != nil {
					return false, err
				}
				if testApp.Status.AvailableReplicas != testAppReplicas {
					return false, err
				}
				return true, err
			})
			Expect(err).ToNot(HaveOccurred(), "Time out or error waiting for hello-openshift (deployment name: '"+testApp.GetName()+"') to become ready.")

			testAppService, err := h.Kube().CoreV1().Services(h.CurrentProject()).Create(context.TODO(), &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: testInstanceName,
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name:     "8080-tcp",
							Protocol: corev1.ProtocolTCP,
							Port:     8080,
							TargetPort: intstr.IntOrString{
								Type:   intstr.Int,
								IntVal: 8080,
							},
						},
						{
							Name:     "8888-tcp",
							Protocol: corev1.ProtocolTCP,
							Port:     8888,
							TargetPort: intstr.IntOrString{
								Type:   intstr.Int,
								IntVal: 8888,
							},
						},
					},
					Selector: map[string]string{"deployment": testInstanceName},
					Type:     corev1.ServiceTypeClusterIP,
				},
			}, metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred())

			testRoute, err := h.Route().RouteV1().Routes(h.CurrentProject()).Create(context.TODO(), &routev1.Route{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Route",
					APIVersion: "route.openshift.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: testInstanceName,
				},
				Spec: routev1.RouteSpec{
					Host: "osde2e-customdomain." + testDomain.Spec.Domain,
					Port: &routev1.RoutePort{
						TargetPort: intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "8080-tcp",
						},
					},
					TLS: &routev1.TLSConfig{
						Termination: routev1.TLSTerminationEdge,
					},
					To: routev1.RouteTargetReference{
						Kind: "Service",
						Name: testAppService.GetName(),
					},
				},
			}, metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred())

			// TODO: Look at tweaking local DNS client configuration here
			ginkgo.By("Requesting the app using the custom domain")
			dialer := &net.Dialer{
				Timeout: dnsTimeout,
			}
			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
					DialContext: func(context context.Context, network, addr string) (net.Conn, error) {
						// Simulate DNS lookup
						if addr == testRoute.Spec.Host+":443" {
							addr = testDomain.Status.Endpoint + ":443"
						}
						return dialer.DialContext(context, network, addr)
					},
				},
			}

			// This ensures that the app is available and returns a response
			var response *http.Response
			wait.PollImmediate(pollInterval, defaultTimeout, func() (bool, error) {
				response, err = client.Get("https://" + testRoute.Spec.Host)
				defer func(response *http.Response) {
					if response != nil {
						response.Body.Close()
					}
				}(response)

				if response != nil && response.StatusCode == http.StatusOK {
					return true, err
				}
				return false, err
			})
			Expect(err).ToNot(HaveOccurred(), "Timed out or error requesting hello-openshift service via custom domain (customdomain endpoint: '"+testDomain.Status.Endpoint+"').")
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.AfterEach(func() {
			ginkgo.By("Deleting resources after testing")
			_, err = h.Cfg().ConfigV1().RESTClient().Delete().
				AbsPath("/apis/managed.openshift.io/v1alpha1").
				Resource("customdomains").
				Name(testDomain.GetName()).
				DoRaw(context.TODO())
			Expect(err).ToNot(HaveOccurred())

			err = h.Kube().CoreV1().Secrets(h.CurrentProject()).Delete(context.TODO(), testDomain.Spec.Certificate.Name, metav1.DeleteOptions{})
			Expect(err).ToNot(HaveOccurred())
			err = h.Kube().AppsV1().Deployments(h.CurrentProject()).Delete(context.TODO(), testInstanceName, metav1.DeleteOptions{})
			Expect(err).ToNot(HaveOccurred())
			err = h.Kube().CoreV1().Services(h.CurrentProject()).Delete(context.TODO(), testInstanceName, metav1.DeleteOptions{})
			Expect(err).ToNot(HaveOccurred())
			err = h.Route().RouteV1().Routes(h.CurrentProject()).Delete(context.TODO(), testInstanceName, metav1.DeleteOptions{})
			Expect(err).ToNot(HaveOccurred())

			h.Impersonate(rest.ImpersonationConfig{})

		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})
