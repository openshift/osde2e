package verify

import "log"

var ClusterTests []TestRunner

const projectPrefix = "osd-e2e"

func RunTests(kubeconfig []byte) (interface{}, error) {
	cluster, err := NewCluster(kubeconfig)
	if err != nil {
		return nil, err
	}

	for _, t := range ClusterTests {
		_, err = runTest(cluster, t)
		if err != nil {
			log.Printf("Failed running tests: %v", err)
		}
	}

	return nil, nil
}

func runTest(cluster *Cluster, runner TestRunner) (interface{}, error) {
	proj, err := cluster.createProject(projectPrefix)
	if err != nil {
		return nil, err
	}
	defer cluster.cleanup(proj.Name)

	return runner(cluster, proj.Name)
}

type TestRunner func(cluster *Cluster, project string) (interface{}, error)
