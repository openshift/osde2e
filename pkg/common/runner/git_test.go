package runner

import (
	"fmt"
	"math/rand"
	"testing"

	. "github.com/onsi/gomega"

	kubev1 "k8s.io/api/core/v1"
)

func TestRunnerGit(t *testing.T) {
	g := NewGomegaWithT(t)

	// setup runner
	runner := setupRunner(t)
	runner.Name = "runner-git-test"
	runner.Namespace = "default"

	// configure runner to clone osde2e repo
	mntPath := "/tmp/osde2e"
	runner.Repos = Repos{
		{
			Name:      "osde2e",
			URL:       "https://github.com/openshift/osde2e.git",
			MountPath: mntPath,
		},
	}

	// have runner write master sha as result
	masterRev := ".git/refs/heads/master"
	resultName := "head-sha.txt"
	runner.Cmd = fmt.Sprintf("cp %s/%s %s/%s", mntPath, masterRev, runner.OutputDir, resultName)

	// execute runner
	stopCh := make(chan struct{})
	err := runner.Run(1800, stopCh)
	g.Expect(err).NotTo(HaveOccurred())

	// get results
	results, err := runner.RetrieveResults()
	g.Expect(err).NotTo(HaveOccurred())

	// verify sha in results
	g.Expect(results).To(HaveKey(resultName))
	g.Expect(results[resultName]).To(HaveLen(41), "should have 40 char SHA hash + newline")
}

func TestGitConfiguresPod(t *testing.T) {
	g := NewGomegaWithT(t)

	// check confirms that configured Pods have proper container count
	check := func(repoCount int) {
		podSpec := kubev1.PodSpec{
			Containers: []kubev1.Container{
				DefaultContainer,
			},
		}

		repos := randRepos(repoCount)
		repos.ConfigurePod(&podSpec)

		g.Expect(podSpec.InitContainers).Should(HaveLen(repoCount))
		g.Expect(podSpec.Containers[0].VolumeMounts).To(HaveLen(repoCount))
		g.Expect(podSpec.Volumes).To(HaveLen(repoCount))
	}

	// check with no repo
	check(0)

	// check with 1 repo
	check(1)

	// check several times with random number of repos
	for i := 0; i < 10; i++ {
		repoCount := rand.Intn(40)
		check(repoCount)
	}
}

func randRepos(count int) (repos Repos) {
	for i := 0; i < count; i++ {
		repos = append(repos, GitRepo{
			Name:      RandomStr(5),
			URL:       fmt.Sprintf("https://%s.com/%s/%s", RandomStr(5), RandomStr(3), RandomStr(3)),
			MountPath: fmt.Sprintf("/mnt/%s", RandomStr(4)),
		})
	}
	return
}
