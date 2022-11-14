package operators

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/openshift/osde2e/pkg/common/util"

	routev1 "github.com/openshift/api/route/v1"
	customdomainv1alpha1 "github.com/openshift/custom-domains-operator/pkg/apis/customdomain/v1alpha1"

	"github.com/openshift/osde2e/pkg/common/alert"
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
	pollInterval                  = 10 * time.Second
	defaultTimeout                = 5 * time.Minute
	endpointReadyTimeout          = 5 * time.Minute
	endpointResolveTimeout        = 20 * time.Minute
	dnsResolverTimeout            = 10 * time.Second
	minConsecutiveResolves        = 5
	routeHostname                 = "hello-openshift"
)

func init() {
	alert.RegisterGinkgoAlert(customDomainsOperatorTestName, "SD-SREP", "@custom-domains-operator", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(customDomainsOperatorTestName, func() {
	// custom dialer for use w/ resolver and http.client
	dialer := &net.Dialer{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: dnsResolverTimeout,
				}
				return d.DialContext(ctx, network, address)
			},
		},
	}

	ginkgo.Context("Should allow dedicated-admins to", func() {
		var (
			h = helper.New()

			testInstanceName = "test-" + time.Now().Format("20060102-150405-") + fmt.Sprint(time.Now().Nanosecond()/1000000) + "-" + fmt.Sprint(ginkgo.GinkgoParallelProcess())
			testDomain       *customdomainv1alpha1.CustomDomain
			testDomainSecret *corev1.Secret
			err              error
		)

		// BeforeEach initializes a CustomDomain for testing
		ginkgo.BeforeEach(func(ctx context.Context) {
			ginkgo.By("Logging in as a dedicated-admin")
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"dedicated-admins",
				},
			})

			ginkgo.By("Creating an ssl certificate and tls secret in OSD")
			testDomainName := fmt.Sprintf("%s.io", testInstanceName)
			testDnsNames := []string{fmt.Sprintf("*.%s", testDomainName)}
			testDomainSecret, err = createTlsSecret(ctx, h, testInstanceName, testDnsNames)
			Expect(err).ToNot(HaveOccurred())
			log.Printf("Created secret %s", testDomainSecret.Name)

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
				DoRaw(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for CustomDomain CR Endpoint to be ready
			ginkgo.By("Wait for CustomDomain CR Endpoint to be ready")
			err = wait.PollImmediate(pollInterval, endpointReadyTimeout, func() (bool, error) {
				byteResult, err := h.Cfg().ConfigV1().RESTClient().Get().
					AbsPath("/apis/managed.openshift.io/v1alpha1").
					Resource("customdomains").
					Name(testInstanceName).
					DoRaw(ctx)
				if err != nil {
					return false, err
				}
				err = json.Unmarshal(byteResult, testDomain)
				if err != nil {
					return false, err
				}
				if testDomain.Status.State == "Ready" && testDomain.Status.Endpoint != "" {
					return true, err
				}
				return false, err
			})
			Expect(err).ToNot(HaveOccurred(), "Time out or error waiting for customdomain '"+testDomain.GetName()+"' to become ready.")
			Expect(string(testDomain.Status.State)).To(Equal("Ready"), "Customdomain may be unstable (.status.state is not 'Ready' anymore)")
			Expect(string(testDomain.Status.Endpoint)).ToNot(Equal(""), "Customdomain's .status.endpoint field empty when .status.state field is 'Ready'")
			log.Printf("Created CustomDomain %s", testDomain.Name)

			// Customdomain ready, now wait for endpoint to resolve consistently.
			// Customdomain endpoints have been known to resolve once, then fail to resolve for a time after
			// To ensure the endpoint is ready & stable, check that it resolves successfully several times before continuing
			ginkgo.By("Wait for CustomDomain endpoint to resolve")
			consecutiveResolves := 0
			err = wait.PollImmediate(pollInterval, endpointResolveTimeout, func() (bool, error) {
				endpointIPs, err := dialer.Resolver.LookupHost(ctx, testDomain.Status.Endpoint)
				log.Printf("Waiting for CustomDomain %s endpoint to resolve: %s", testDomain.Name, endpointIPs)
				if len(endpointIPs) == 0 {
					// No IPs returned
					consecutiveResolves = 0
					return false, nil
				}
				if err != nil {
					return false, err
				}
				for _, addr := range endpointIPs {
					if net.ParseIP(addr) == nil {
						// Not a valid ip address
						consecutiveResolves = 0
						return false, nil
					}
				}
				if consecutiveResolves < minConsecutiveResolves-1 {
					consecutiveResolves++
					return false, err
				}
				return true, err
			})
			Expect(err).ToNot(HaveOccurred(), "Time out or error waiting for customdomain endpoint '"+testDomain.Status.Endpoint+"' to resolve.")
		}, float64(defaultTimeout.Seconds()*3))

		// Now that the endpoint is stable, make sure it's resolvable and usable.
		util.GinkgoIt("Create customdomains that are resolvable by external services", func(ctx context.Context) {
			ginkgo.By("Logging in as a dedicated-admin")
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"dedicated-admins",
				},
			})

			ginkgo.By("Creating a new app")
			testAppReplicas := int32(1)
			testApp, err := h.Kube().AppsV1().Deployments(h.CurrentProject()).Create(ctx, &appsv1.Deployment{
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
								{
									Name:  "hello-openshift",
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
			ginkgo.By("Ensuring the new app is created")
			err = wait.PollImmediate(pollInterval, defaultTimeout, func() (bool, error) {
				testApp, err = h.Kube().AppsV1().Deployments(h.CurrentProject()).Get(ctx, testApp.GetName(), metav1.GetOptions{})
				if err != nil {
					return false, err
				}
				if testApp.Status.AvailableReplicas != testAppReplicas {
					return false, err
				}
				return true, err
			})
			Expect(err).ToNot(HaveOccurred(), "Time out or error waiting for hello-openshift (deployment name: '"+testApp.GetName()+"') to become ready.")
			defer func() {
				log.Printf("Deleting app %s", testApp.Name)
				err = h.Kube().AppsV1().Deployments(h.CurrentProject()).Delete(ctx, testApp.Name, metav1.DeleteOptions{})
				Expect(err).ToNot(HaveOccurred())
			}()
			log.Printf("Created app %s", testApp.Name)

			ginkgo.By("Exposing the new app via an Openshift route")
			testAppService, err := h.Kube().CoreV1().Services(h.CurrentProject()).Create(ctx, &corev1.Service{
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
			defer func() {
				log.Printf("Deleting service %s", testAppService.Name)
				err = h.Kube().CoreV1().Services(h.CurrentProject()).Delete(ctx, testAppService.Name, metav1.DeleteOptions{})
				Expect(err).ToNot(HaveOccurred())
			}()

			testRoute, err := h.Route().RouteV1().Routes(h.CurrentProject()).Create(ctx, &routev1.Route{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Route",
					APIVersion: "route.openshift.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: testInstanceName,
				},
				Spec: routev1.RouteSpec{
					Host: routeHostname + "." + testDomain.Spec.Domain,
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
			defer func() {
				log.Printf("Deleting route %s", testRoute.Name)
				err = h.Route().RouteV1().Routes(h.CurrentProject()).Delete(ctx, testRoute.Name, metav1.DeleteOptions{})
				Expect(err).ToNot(HaveOccurred())
			}()

			ginkgo.By("Requesting the app using the custom domain")
			// dialContext customized for http client to simulate DNS lookup
			dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
				if addr == testRoute.Spec.Host+":443" {
					addr = testDomain.Status.Endpoint + ":443"
				}
				return dialer.DialContext(ctx, network, addr)
			}
			http.DefaultTransport.(*http.Transport).DialContext = dialContext
			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
					DialContext: dialContext,
				},
			}

			// This ensures that the app is available and returns a response
			var response *http.Response
			err = wait.PollImmediate(pollInterval, defaultTimeout*3, func() (bool, error) {
				log.Printf("Waiting for app %s to respond", testApp.Name)
				response, err = client.Get("https://" + testRoute.Spec.Host)
				defer func(response *http.Response) {
					if response != nil {
						response.Body.Close()
					}
				}(response)
				if err != nil {
					// Check for DNS error
					if e, ok := err.(*net.DNSError); ok && e.IsNotFound {
						// do not abort on flaky DNS responses
						return false, nil
					}
					// Check for URL DNS error
					if urlError, ok := err.(*url.Error); ok {
						if _, ok := urlError.Err.(*net.OpError); ok {
							// do not abort on flaky network operations
							return false, nil
						}
					}
					// Unhandled error
					return false, err
				}
				if response != nil && response.StatusCode == http.StatusOK {
					return true, nil
				}
				return false, nil
			})
			Expect(err).ToNot(HaveOccurred(), "Timed out or error requesting hello-openshift service via custom domain (customdomain endpoint: '"+testDomain.Status.Endpoint+"').")
		}, float64(defaultTimeout.Seconds()*4))

		// Ensure dedicated-admins can update CustomDomain certificates
		util.GinkgoIt("Replace certificates", func(ctx context.Context) {
			ginkgo.By("Retrieving the original certificate")
			h.Impersonate(rest.ImpersonationConfig{})
			oldIngressSecret, err := getSecretForCustomDomain(ctx, h, testDomain)
			Expect(err).ToNot(HaveOccurred())
			log.Printf("Retrieved ingress secret: %s", oldIngressSecret.Name)

			ginkgo.By("Impersonating a dedicated-admin")
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "dummy-admin@redhat.com",
				Groups: []string{
					"dedicated-admins",
				},
			})

			ginkgo.By("Generating a new certificate")
			newSecretName := fmt.Sprintf("%s-new-secret", testInstanceName)
			newDnsNames := []string{fmt.Sprintf("*.%s.io", testInstanceName)}
			newSecret, err := createTlsSecret(ctx, h, newSecretName, newDnsNames)
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				log.Printf("Deleting secret %s", newSecret.Name)
				err = h.Kube().CoreV1().Secrets(h.CurrentProject()).Delete(ctx, newSecret.Name, metav1.DeleteOptions{})
				Expect(err).ToNot(HaveOccurred())
			}()
			log.Printf("Generated new secret: %s", newSecret.Name)

			ginkgo.By("Replacing the certificate in the customdomain")
			testDomain.Spec.Certificate.Name = newSecret.Name
			_, err = h.Cfg().ConfigV1().RESTClient().Put().
				AbsPath("/apis/managed.openshift.io/v1alpha1").
				Resource("customdomains").
				Name(testDomain.GetName()).
				Body(testDomain).
				DoRaw(ctx)
			Expect(err).ToNot(HaveOccurred())
			log.Printf("Updated CustomDomain with new secret")

			ginkgo.By("Checking that the ingress secret matches the new tls secret")
			h.Impersonate(rest.ImpersonationConfig{})
			var currentIngressSecret *corev1.Secret
			err = wait.PollImmediate(pollInterval, defaultTimeout, func() (bool, error) {
				log.Printf("Checking that secret change propagated")
				currentIngressSecret, err = getSecretForCustomDomain(ctx, h, testDomain)
				if err != nil || bytes.Equal(currentIngressSecret.Data["tls.crt"], oldIngressSecret.Data["tls.crt"]) {
					return false, err
				}
				return true, err
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(currentIngressSecret.Data["tls.crt"]).To(Equal(newSecret.Data["tls.crt"]))
			Expect(currentIngressSecret.Data["tls.key"]).To(Equal(newSecret.Data["tls.key"]))
			log.Printf("Verified that new secret is being used as expected")
		}, float64(defaultTimeout.Seconds()*2))

		// AfterEach deletes resources created by BeforeEach
		ginkgo.AfterEach(func(ctx context.Context) {
			log.Printf("Cleaning up after testing")
			_, err := h.Cfg().ConfigV1().RESTClient().Delete().
				AbsPath("/apis/managed.openshift.io/v1alpha1").
				Resource("customdomains").
				Name(testDomain.GetName()).
				DoRaw(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = h.Kube().CoreV1().Secrets(h.CurrentProject()).Delete(ctx, testDomainSecret.Name, metav1.DeleteOptions{})
			Expect(err).ToNot(HaveOccurred())

			h.Impersonate(rest.ImpersonationConfig{})
		}, float64(defaultTimeout.Seconds()))
	})
})

func getSecretForCustomDomain(ctx context.Context, h *helper.H, customDomain *customdomainv1alpha1.CustomDomain) (secret *corev1.Secret, err error) {
	return h.Kube().CoreV1().Secrets("openshift-ingress").Get(ctx, customDomain.Name, metav1.GetOptions{})
}

// createTlsSecret creates a TLS-type secret on the cluster, returning the resulting object
func createTlsSecret(ctx context.Context, h *helper.H, secretName string, dnsNames []string) (secret *corev1.Secret, err error) {
	customDomainRSAKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	customDomainX509Template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:       []string{"Red Hat, Inc"},
			OrganizationalUnit: []string{"Openshift Dedicated End-to-End Testing"},
		},
		DNSNames:              dnsNames,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, 1),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	customDomainX509, err := x509.CreateCertificate(rand.Reader, customDomainX509Template, customDomainX509Template, &customDomainRSAKey.PublicKey, customDomainRSAKey)
	if err != nil {
		return
	}

	secretData := make(map[string][]byte)
	secretData["tls.key"] = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(customDomainRSAKey),
	})
	secretData["tls.crt"] = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: customDomainX509,
	})

	secret, err = h.Kube().CoreV1().Secrets(h.CurrentProject()).Create(ctx, &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Type: corev1.SecretTypeTLS,
		Data: secretData,
	}, metav1.CreateOptions{})
	return
}
