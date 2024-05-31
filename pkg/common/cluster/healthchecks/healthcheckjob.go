package healthchecks

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/openshift/osde2e-common/pkg/clients/openshift"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/logging"
	"k8s.io/client-go/rest"
)

// CheckHealthcheckJob uses the `osd-cluster-ready` healthcheck job to determine cluster readiness. If the cluster
// is not ready, it will return an error.
func CheckHealthcheckJob(ctx context.Context, restconfig *rest.Config, logger *log.Logger) error {
	logger = logging.CreateNewStdLoggerOrUseExistingLogger(logger)
	var k8s *openshift.Client

	name := "osd-cluster-ready"

	k8s, err := openshift.NewFromRestConfig(restconfig, ginkgo.GinkgoLogr)
	if err != nil {
		return fmt.Errorf("unable to setup k8s client: %w", err)
	}

	timeout, err := time.ParseDuration(viper.GetString(config.Tests.ClusterHealthChecksTimeout))
	if err != nil {
		return fmt.Errorf("failed parsing health check timeout: %w", err)
	}
	joberr := k8s.OSDClusterHealthy(ctx, name, viper.GetString(config.ReportDir), timeout)

	if joberr == nil {
		logger.Println("Healthcheck job passed")
		return nil
	}

	return joberr
}
