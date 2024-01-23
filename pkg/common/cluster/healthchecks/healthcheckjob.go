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

	joberr := k8s.WatchJob(ctx, namespace, name)

	if joberr == nil {
		logger.Println("Healthcheck job passed")
		return nil
	}

	filename := filepath.Join(viper.GetString(config.ReportDir), fmt.Sprintf("%s.log", name))
	file, err := os.Create(filename)
	if err != nil {
		fmt.Print("could not create osd-cluster-ready log file: ", err)
	} else {
		logs, err := k8s.GetJobLogs(ctx, name, namespace)
		if err != nil {
			fmt.Print("could not get job logs: ", err)
		} else {
			file.WriteString(logs)
		}
	}
	return joberr
}
