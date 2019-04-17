package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/openshift/osde2e/pkg/verify"

	uhc "github.com/openshift-online/uhc-sdk-go/pkg/client"
)

const (
	UHCTokenEnv = "UHC_TOKEN"
	AWSIDEnv    = "AWS_ACCESS_KEY_ID"
	AWSKeyEnv   = "AWS_SECRET_ACCESS_KEY"
	ProdEnv     = "USE_PROD"

	ClusterStateReady = "ready"
)

var UHCConn *uhc.Connection

func main() {
	uhcToken := getVar(UHCTokenEnv)
	awsID := getVar(AWSIDEnv)
	awsKey := getVar(AWSKeyEnv)

	// use staging unless told to use prod
	prod := os.Getenv(ProdEnv)
	staging := len(prod) == 0

	var err error
	UHCConn, err = setupConn(uhcToken, staging)
	defer UHCConn.Close()
	if err != nil {
		log.Fatalf("couldn't build connection: %v", err)
	}

	// Create cluster
	name := "osde2e-test5"
	cluster, err := LaunchCluster(name, awsID, awsKey)
	if err != nil {
		log.Fatal("Could not launch cluster:", err)
	}

	// get cluster ID
	clusterId, err := getStr(cluster, "id")
	if err != nil {
		log.Fatalf("can't test cluster: %v", err)
	}

	// wait for cluster to be ready and ensure its destroyed after testing
	defer teardown(clusterId)
	if err = waitForClusterReady(clusterId); err != nil {
		log.Println(err)
		return
	}

	kubeconfig, err := ClusterKubeconfig(clusterId)
	if err != nil {
		log.Printf("Error getting kubeconfig for cluster '%s': %v", clusterId, err)
		return
	}

	_, err = verify.RunTests(kubeconfig)
	if err != nil {
		log.Printf("Failed running tests on cluster '%s': %v", clusterId, err)
	}
	//- Run cluster verification tests in Pod within cluster
	//- Submit results to TestGrid
}

func waitForClusterReady(clusterId string) error {
	times, wait := 6, 30*time.Second
	log.Printf("Waiting %v for cluster '%s' to be ready...\n", time.Duration(times)*wait, clusterId)

	for i := 0; i < times; i++ {
		if state, err := ClusterState(clusterId); state == ClusterStateReady {
			return nil
		} else if err != nil {
			log.Print("Encountered error waiting for cluster:", err)
		} else {
			log.Printf("Cluster is not ready, current status '%s'.", state)
		}

		time.Sleep(wait)
	}

	time.Sleep(time.Second)
	return fmt.Errorf("timed out waiting for cluster '%s' to be ready", clusterId)
}

func teardown(clusterId string) {
	log.Printf("Destroying cluster '%s'...", clusterId)
	if err := deleteClusterReq(clusterId); err != nil {
		log.Fatalf("Failed to destroy cluster '%s': %v", clusterId, err)
	}
}

func setupConn(uhcToken string, staging bool) (*uhc.Connection, error) {
	logger, err := uhc.NewGoLoggerBuilder().
		Debug(true).
		Build()
	if err != nil {
		log.Fatalf("couldn't build logger: %v", err)
	}

	builder := uhc.NewConnectionBuilder().
		Logger(logger).
		Tokens(uhcToken)

	if staging {
		builder.URL(StagingURL)
	}

	return builder.Build()

}

func getVar(name string) string {
	contents, ok := os.LookupEnv(name)
	if !ok {
		log.Fatalf("'%s' must be provided", name)
	}
	return contents
}
