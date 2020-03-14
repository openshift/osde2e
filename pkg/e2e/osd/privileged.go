package osd
import (
    "fmt"
    "log"
    "time"
    "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    projectv1 "github.com/openshift/api/project/v1"
    "github.com/openshift/osde2e/pkg/common/config"
    "github.com/openshift/osde2e/pkg/common/helper"
    "github.com/openshift/osde2e/pkg/common/state"
    v1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/rest"
)
func makePod(name string, privileged bool) v1.Pod {
    return v1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name: name,
        },
        Spec: v1.PodSpec{
            Containers: []v1.Container{
                {
                    Name:  "test",
                    Image: "registry.access.redhat.com/ubi8/ubi-minimal",
                    SecurityContext: &v1.SecurityContext{
                        Privileged: &privileged,
                    },
                },
            },
        },
    }
}
var _ = ginkgo.Describe("[Suite: service-definition] [OSD] Privileged Containers", func() {
    ginkgo.Context("Privileged containers are not allowed", func() {
        ginkgo.It("privileged container should not get created", func() {
            // setup helper
            h := &helper.H{
                State: state.Instance,
            }
            h.SetupNoProj()
            defer h.Cleanup()
            suffix := helper.RandomStr(5)
            projectName := fmt.Sprintf("osde2e-%s", suffix)
            projUser := fmt.Sprintf("%s-user", projectName)
            sa, err := h.Kube().CoreV1().ServiceAccounts("dedicated-admin").Create(&v1.ServiceAccount{
                ObjectMeta: metav1.ObjectMeta{
                    Name: projUser,
                },
			})
			Expect(err).NotTo(HaveOccurred())
            log.Printf("Created SA: %v", sa.GetName())
            proj := &projectv1.Project{
                ObjectMeta: metav1.ObjectMeta{
                    Name: projectName,
                },
            }
            proj, err = h.Project().ProjectV1().Projects().Create(proj)
            Expect(err).NotTo(HaveOccurred())
            log.Printf("Created Project: %v", proj.GetName())
            h.SetProject(proj)
            time.Sleep(65 * time.Second)
            // Let's impersonate
            h.Impersonate(rest.ImpersonationConfig{
                UserName: fmt.Sprintf("system:serviceaccount:dedicated-admin:%s", projUser),
            })
            // Test creating a privileged pod and expect a failure
            pod := makePod("privileged-pod", true)
            _, err = h.Kube().CoreV1().Pods(projectName).Create(&pod)
            Expect(err).To(HaveOccurred())
            // Test creating an unprivileged pod and expect success
            pod = makePod("unprivileged-pod", false)
            _, err = h.Kube().CoreV1().Pods(projectName).Create(&pod)
            Expect(err).NotTo(HaveOccurred())
            h.Impersonate(rest.ImpersonationConfig{
                UserName: "",
            })
            err = h.Kube().CoreV1().ServiceAccounts("dedicated-admin").Delete(projUser, &metav1.DeleteOptions{})
            Expect(err).NotTo(HaveOccurred())
        }, float64(config.Instance.Tests.PollingTimeout))
    })
})
