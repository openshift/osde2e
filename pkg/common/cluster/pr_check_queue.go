package cluster

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openshift/osde2e/pkg/common/clusterproperties"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/spi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	interval = 5 * time.Minute
	timeout  = 190 * time.Minute
)

func PrCheckQueue(o spi.Provider) error {

	cluster, err := o.GetCluster(viper.GetString(config.Cluster.ID))
	if err != nil {
		return fmt.Errorf("error getting cluster: %v", err)
	}

	properties := cluster.Properties()

	err = wait.PollImmediate(interval, timeout, func() (bool, error) {

		if properties[clusterproperties.JobID] == "" {
			err = o.AddProperty(cluster, clusterproperties.JobStartedAt, viper.GetString(config.JobStartedAt))
			if err != nil {
				return false, fmt.Errorf("error adding property to cluster: %v", err)
			}
			return true, nil
		}

		startTime, err := time.Parse(time.RFC3339, viper.GetString(config.JobStartedAt))
		if err != nil {
			return false, fmt.Errorf("error parsing Time for queue started at: %v", err)
		}

		//Check that there is a running job and if there is, check that it is not older than 3 hours
		if properties[clusterproperties.JobID] != "" && startTime.After(startTime.Add(3*time.Hour)) {
			if err := prCheckClean(o, cluster); err != nil {
				return false, fmt.Errorf("error cleaning up aborted jobs: %v", err)
			}
		}

		return false, err
	})
	return nil
}

func prCheckClean(o spi.Provider, cluster *spi.Cluster) error {

	h := helper.NewOutsideGinkgo()

	namespaces, err := h.Kube().CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error getting namespaces: %v", err)
	}

	//This is a very slow implementation. Can KubeClient be used to list namespaces on a regex or can we add a filter to the list?
	for _, namespace := range namespaces.Items {
		if strings.Contains(namespace.Name, "osde2e-") {
			if err := h.Kube().CoreV1().Namespaces().Delete(context.Background(), namespace.Name, metav1.DeleteOptions{}); err != nil {
				return fmt.Errorf("error deleting namespace: %v", err)
			}
		}
	}

	//Sets the JobID to empty string in order to exit the polling loop
	if err = o.AddProperty(cluster, clusterproperties.JobID, ""); err != nil {
		return fmt.Errorf("error adding property to cluster: %v", err)
	}

	//Sets the job started at to the current job's start time.
	if err = o.AddProperty(cluster, clusterproperties.JobStartedAt, viper.GetString(config.JobStartedAt)); err != nil {
		return fmt.Errorf("error adding property to cluster: %v", err)
	}

	return nil
}
