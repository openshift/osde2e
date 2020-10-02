package create

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/openshift/osde2e/cmd/osde2e/common"
	"github.com/openshift/osde2e/cmd/osde2e/helpers"
	clusterutil "github.com/openshift/osde2e/pkg/common/cluster"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"github.com/openshift/osde2e/pkg/common/versions"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Cmd = &cobra.Command{
	Use:   "create",
	Short: "Creates new clusters",
	Long:  "Creates new clusters using the provided arguments.",
	Args:  cobra.OnlyValidArgs,
	RunE:  run,
}

var args struct {
	configString          string
	customConfig          string
	secretLocations       string
	environment           string
	kubeConfig            string
	numberOfClusters      int
	batchSize             int
	secondsBetweenBatches int
}

var discardLogger *log.Logger

func init() {
	pfs := Cmd.PersistentFlags()
	pfs.StringVar(
		&args.configString,
		"configs",
		"",
		"A comma separated list of built in configs to use",
	)
	Cmd.RegisterFlagCompletionFunc("configs", helpers.ConfigComplete)
	pfs.StringVar(
		&args.customConfig,
		"custom-config",
		"",
		"Custom config file for osde2ectl",
	)
	pfs.StringVar(
		&args.secretLocations,
		"secret-locations",
		"",
		"A comma separated list of possible secret directory locations for loading secret configs.",
	)
	pfs.StringVarP(
		&args.environment,
		"environment",
		"e",
		"",
		"Cluster provider environment to use.",
	)
	pfs.StringVarP(
		&args.kubeConfig,
		"kube-config",
		"k",
		"",
		"Path to local Kube config for running tests against.",
	)
	pfs.IntVarP(
		&args.numberOfClusters,
		"number-of-clusters",
		"n",
		1,
		"Specify the total number of clusters to create.",
	)
	pfs.IntVarP(
		&args.batchSize,
		"batch-size",
		"b",
		-1,
		"The number of clusters that should be created at one time. A value of 0 or less will create all clusters at once.",
	)
	pfs.IntVarP(
		&args.secondsBetweenBatches,
		"seconds-between-batches",
		"s",
		120,
		"The number of seconds between batches of cluster provisions.",
	)

	viper.BindPFlag(config.Cluster.ID, Cmd.PersistentFlags().Lookup("cluster-id"))
	viper.BindPFlag(ocmprovider.Env, Cmd.PersistentFlags().Lookup("environment"))
	viper.BindPFlag(config.Kubeconfig.Path, Cmd.PersistentFlags().Lookup("kube-config"))

	discardLogger = log.New(ioutil.Discard, "", 0)
}

func run(cmd *cobra.Command, argv []string) error {
	if err := common.LoadConfigs(args.configString, args.customConfig, args.secretLocations); err != nil {
		return fmt.Errorf("error loading initial state: %v", err)
	}

	// configure cluster and upgrade versions
	if err := versions.ChooseVersions(); err != nil {
		return fmt.Errorf("failed to configure versions: %v", err)
	}

	batchSize := args.batchSize
	if batchSize <= 0 {
		log.Printf("Provisioning %d clusters all at once.", args.numberOfClusters)
		batchSize = args.numberOfClusters
	} else {
		log.Printf("Provisioning %d clusters in batches of %d, waiting %d seconds in between.", args.numberOfClusters, args.batchSize, args.secondsBetweenBatches)
	}

	clustersDir := filepath.Join(viper.GetString(config.ReportDir), "clusters")

	if _, err := os.Stat(clustersDir); os.IsNotExist(err) {
		err := os.Mkdir(clustersDir, os.FileMode(0755))

		if err != nil {
			return fmt.Errorf("unable to create clusters directory: %v", err)
		}
	}

	var successfulClustersCounter int32 = 0
	createClusters(args.numberOfClusters, batchSize, args.secondsBetweenBatches, &successfulClustersCounter)

	fmt.Printf("Successfully provisioned %d/%d clusters.\n", successfulClustersCounter, args.numberOfClusters)

	return nil
}

func createClusters(numClusters, batchSize, waitSecondsBetweenBatches int, successfulClustersCounter *int32) {
	totalBatches := int(math.Ceil(float64(numClusters) / float64(batchSize)))
	batchWg := &sync.WaitGroup{}
	batchWg.Add(totalBatches)

	for batchIteration := 0; batchIteration < totalBatches; batchIteration++ {
		remainingClusters := numClusters - batchIteration*batchSize
		adjustedBatchSize := batchSize

		if remainingClusters < batchSize {
			adjustedBatchSize = remainingClusters
		}

		log.Printf("Provisioning %d clusters in batch %d", adjustedBatchSize, batchIteration)
		go createBatch(batchIteration, adjustedBatchSize, batchWg, successfulClustersCounter)

		if remainingClusters > batchSize {
			log.Printf("Sleeping for %d seconds before next batch", waitSecondsBetweenBatches)
			time.Sleep(time.Second * time.Duration(waitSecondsBetweenBatches))
		} else {
			log.Printf("Provisioned final batch of %d clusters.\n", adjustedBatchSize)
		}
	}

	log.Printf("Waiting for all batches.")
	batchWg.Wait()
}

func createBatch(batchIteration int, numClustersInBatch int, batchWg *sync.WaitGroup, successfulClustersCounter *int32) {
	wg := &sync.WaitGroup{}
	wg.Add(numClustersInBatch)

	for i := 0; i < numClustersInBatch; i++ {
		go setupCluster(wg, successfulClustersCounter)
	}

	wg.Wait()
	log.Printf("Finished waiting for batch %d.", batchIteration)
	batchWg.Done()
}

func setupCluster(wg *sync.WaitGroup, successfulClustersCounter *int32) {
	defer wg.Done()
	cluster, err := clusterutil.ProvisionCluster(discardLogger)

	if err != nil {
		if cluster != nil {
			fmt.Printf("error while trying to provision up cluster with ID %s: %v\n", cluster.ID(), err)
		} else {
			fmt.Printf("error while provisioning the cluster: %v\n", err)
		}
	} else {
		log.Printf("Starting provisioning cluster %s.", cluster.ID())
		outputFilePath := filepath.Join(viper.GetString(config.ReportDir), "clusters", cluster.ID()+".log")

		outputFile, err := os.Create(outputFilePath)
		defer outputFile.Close()

		if err != nil {
			fmt.Printf("error opening logfile for writing: %v", err)
			return
		}

		logger := log.New(outputFile, "", log.LstdFlags)

		terminate := make(chan bool)

		go func() {
			timeout := make(chan bool)

			for {

				go func() {
					time.Sleep(time.Minute * time.Duration(5))
					timeout <- true
				}()

				select {
				case <-timeout:
					isHealthy, failures, _ := clusterutil.PollClusterHealth(cluster.ID(), discardLogger)
					if isHealthy {
						fmt.Printf("Cluster %s is healthy (could be transient).\n", cluster.ID())
					} else {
						fmt.Printf("Cluster %s is not healthy yet.\n", cluster.ID())
						if len(failures) > 0 {
							fmt.Printf("Currently failing %s health checks", strings.Join(failures, ", "))
						}
					}
				case <-terminate:
					return
				}
			}
		}()

		err = clusterutil.WaitForClusterReady(cluster.ID(), logger)

		terminate <- true

		if err != nil {
			fmt.Printf("Cluster %s never became healthy.\n", cluster.ID())
		} else {
			fmt.Printf("Cluster %s is healthy.\n", cluster.ID())

			atomic.AddInt32(successfulClustersCounter, 1)
		}
	}
}
