package healthchecks

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/openshift/osde2e-common/pkg/clients/openshift"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/logging"
	"k8s.io/client-go/kubernetes"
)

// CheckHealthcheckJob uses the `osd-cluster-ready` healthcheck job to determine cluster readiness. If the cluster
// is not ready, it will return an error.
func CheckHealthcheckJob(k8sClient *kubernetes.Clientset, ctx context.Context, logger *log.Logger) error {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)
	var k8s *openshift.Client

	namespace := "openshift-monitoring"
	name := "osd-cluster-ready"

	k8s, err := openshift.New(ginkgo.GinkgoLogr)
	if err != nil {
		return fmt.Errorf("Unable to setup k8s client: %w", err)
	}

	err = k8s.WatchJob(ctx, namespace, name)

	if err == nil {
		logger.Println("Healthcheck job passed")
		return nil
	} else {
		filename := filepath.Join(viper.GetString(config.ReportDir), fmt.Sprintf("%s.log", name))
		file, err := os.Create(filename)
		if err != nil {
			fmt.Printf("could not create osd-cluster-ready log file:  %w", err)
		} else {
			logs, err := k8s.GetJobLogs(ctx, name, namespace)
			if err != nil {
				fmt.Printf("could not get job logs:  %w", err)
			} else {
				file.WriteString(logs)
			}
		}
	}
	return nil
}
