package operators

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"

	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

var pruneJobsTestName string = "[Suite: operators] [OSD] Prune jobs"

func init() {
	alert.RegisterGinkgoAlert(pruneJobsTestName, "SD-SREP", "Haoran Wang", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(pruneJobsTestName, func() {
	h := helper.New()
	ginkgo.Context("pruner jobs should works", func() {
		namespace := "openshift-sre-pruning"
		cronJobs := []string{"builds-pruner", "deployments-pruner"}
		for _, cronJob := range cronJobs {
			util.GinkgoIt(cronJob+" should run successfully", func(ctx context.Context) {
				getOpts := metav1.GetOptions{}
				cjob, err := h.Kube().BatchV1beta1().CronJobs(namespace).Get(ctx, cronJob, getOpts)
				Expect(err).NotTo(HaveOccurred())
				job := createJobFromCronJob(cjob)
				job, err = h.Kube().BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
				defer func() {
					err = h.Kube().BatchV1().Jobs(namespace).Delete(ctx, job.Name, metav1.DeleteOptions{})
					Expect(err).NotTo(HaveOccurred())
				}()
				Expect(err).NotTo(HaveOccurred())
				err = waitJobComplete(ctx, h, namespace, job.Name)
				Expect(err).NotTo(HaveOccurred())
			}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
		}
	})
})

func createJobFromCronJob(cronJob *batchv1beta1.CronJob) *batchv1.Job {
	annotations := make(map[string]string)
	annotations["managed.openshift.io/instantiate"] = "manual"
	for k, v := range cronJob.Spec.JobTemplate.Annotations {
		annotations[k] = v
	}
	jobName := cronJob.Name + "-" + rand.String(5)
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{APIVersion: batchv1.SchemeGroupVersion.String(), Kind: "Job"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        jobName,
			Annotations: annotations,
			Labels:      cronJob.Spec.JobTemplate.Labels,
		},
		Spec: cronJob.Spec.JobTemplate.Spec,
	}
}

func waitJobComplete(ctx context.Context, h *helper.H, namespace, jobName string) error {
	var err error
	var job *batchv1.Job

	// interval is the duration in seconds between polls
	// values here for humans
	interval := 5

	// convert time.Duration type
	timeoutDuration := time.Duration(viper.GetFloat64(config.Tests.PollingTimeout)) * time.Minute
	intervalDuration := time.Duration(interval) * time.Second

	start := time.Now()

Loop:
	for {
		job, err = h.Kube().BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
		elapsed := time.Since(start)

		switch {
		case err == nil && job.Status.Succeeded == 1:
			// Success
			break Loop
		default:
			if elapsed < timeoutDuration {
				log.Printf("Waiting %v for %s job to complete", (timeoutDuration - elapsed), jobName)
				time.Sleep(intervalDuration)
			} else {
				return fmt.Errorf("failed to wait job: %s %s complete before timeout", namespace, jobName)
			}
		}
	}
	return nil
}
