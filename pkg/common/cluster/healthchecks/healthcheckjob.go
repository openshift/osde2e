package healthchecks

import (
	"context"
	"fmt"
	"log"

	"github.com/openshift/osde2e/pkg/common/logging"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	batchv1client "k8s.io/client-go/kubernetes/typed/batch/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// CheckHealthcheckJob uses the `openshift-cluster-ready-*` healthcheck job to determine cluster readiness.
func CheckHealthcheckJob(k8sClient v1.CoreV1Interface, ctx context.Context, logger *log.Logger) (bool, error) {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)

	logger.Print("Checking that all Nodes are running or completed...")

	bv1C := batchv1client.New(k8sClient.RESTClient())
	watcher, err := bv1C.Jobs("openshift-monitoring").Watch(ctx, metav1.ListOptions{
		Watch:         true,
		FieldSelector: "metadata.name=osd-cluster-ready",
	})
	if err != nil {
		if errors.IsNotFound(err) {
			// Job doesn't exist yet
			return false, nil
		}
		// Failed checking for job
		return false, fmt.Errorf("failed looking up job: %w", err)
	}
	for {
		select {
		case event := <-watcher.ResultChan():
			switch event.Type {
			case watch.Added:
				fallthrough
			case watch.Modified:
				job := event.Object.(*batchv1.Job)
				if job.Status.Succeeded > 0 {
					return true, nil
				}
			case watch.Deleted:
				return false, fmt.Errorf("cluster readiness job deleted before becoming ready")
			case watch.Error:
				return false, fmt.Errorf("watch returned error event: %v", event)
			default:
				logger.Printf("Unrecognized event type while watching for healthcheck job updates: %v", event.Type)
			}
		case <-ctx.Done():
			return false, fmt.Errorf("healtcheck watch context cancelled while still waiting for success")
		}
	}
}
