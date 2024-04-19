package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/go-github/v31/github"
	"github.com/kylelemons/godebug/diff"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GenerateDiff attempts to pull a dependency list from a previous job (job, jobID) and generate a diff against a provided string
func GenerateDiff(phase, dependencies string) error {
	baseJobURL := viper.GetString(config.BaseJobURL)
	baseProwURL := viper.GetString(config.BaseProwURL)
	jobName := viper.GetString(config.JobName)
	jobNameSafe := os.Getenv("JOB_NAME_SAFE")

	jobID, err := getLastJobID(baseProwURL, jobName)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s/%d/artifacts/%s/osde2e-test/artifacts/%s/dependencies.txt", baseJobURL, jobName, jobID, jobNameSafe, phase)
	log.Printf("Grabbing diff from %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("dependencies.txt not found at %s", url)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("expected HTTP-200 code at %s", url)
	}

	newDiff := strings.Split(diff.Diff(string(body), dependencies), "\n")
	for _, s := range newDiff {
		if strings.HasPrefix(s, "-") {
			log.Printf("\033[0;31m%s\033[0m\n", s)
		} else if strings.HasPrefix(s, "+") {
			log.Printf("\033[0;32m%s\033[0m\n", s)
		} else {
			log.Println(s)
		}
	}
	return nil
}

// GenerateDependencies creates a list of images and the MCC hash
func GenerateDependencies(kube kubernetes.Interface) (dependencies string, err error) {
	pods, err := kube.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	data := []string{}
	if err != nil {
		return "", err
	}
	images, err := GetImageList(pods)
	if err != nil {
		return "", err
	}
	hash, err := GetCurrentMCCHash()
	if err != nil {
		return "", err
	}

	dependencies = "MCC: " + hash + "\n-----\n"

	for k, v := range images {
		data = append(data, fmt.Sprintf("%-80s: %s", v, k))
	}

	sort.Strings(data)
	dependencies += strings.Join(data, "\n")

	return dependencies, nil
}

// GetImageList gathers all images used in a PodList
func GetImageList(list *v1.PodList) (images map[string]string, err error) {
	tmp := make(map[string]string)

	for _, pod := range list.Items {
		// Default to Unknown for the name
		name := "Unknown"
		if _, ok := pod.ObjectMeta.Labels["app"]; ok {
			name = pod.ObjectMeta.Labels["app"]
		} else if _, ok := pod.ObjectMeta.Labels["name"]; ok {
			name = pod.ObjectMeta.Labels["name"]
		} else if _, ok := pod.ObjectMeta.Labels["k8s-app"]; ok {
			name = pod.ObjectMeta.Labels["k8s-app"]
		}
		tmp = appendUniqueContainers(name, pod.Spec.Containers, tmp)
		tmp = appendUniqueContainers(name, pod.Spec.InitContainers, tmp)
	}

	return tmp, nil
}

// GetCurrentMCCHash attempts to pull back the current master SHA1 hash from GitHub
func GetCurrentMCCHash() (hash string, err error) {
	gh := github.NewClient(nil)

	commits, _, err := gh.Repositories.ListCommits(context.Background(), "openshift", "managed-cluster-config", &github.CommitsListOptions{})
	if err != nil {
		return "", err
	}

	if len(commits) > 0 {
		return commits[0].GetSHA(), nil
	}

	return "", fmt.Errorf("no commits found for openshift/machine-cluster-config")
}

func getLastJobID(baseProwURL, jobName string) (int, error) {
	// Look up the list of previous jobs with a given name
	url := fmt.Sprintf("%s/job-history/gs/origin-ci-test/logs/%s", baseProwURL, jobName)
	if os.Getenv("JOB_TYPE") == "presubmit" {
		url = fmt.Sprintf("%s/job-history/gs/origin-ci-test/pr-logs/directory/%s", baseProwURL, jobName)
	}

	log.Printf("Looking up job history from %s", url)
	res, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return 0, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return 0, err
	}

	// Load the page and look for the JSON object containing the build
	// information. Unmarshal it and look for the second build, assuming that
	// the current running build (this process) is the first build.
	expr := regexp.MustCompile(`var allBuilds = (.*);`)
	scriptBlock := strings.TrimSpace(doc.Find("script").Last().Text())
	exprMatches := expr.FindStringSubmatch(scriptBlock)
	if len(exprMatches) == 0 {
		return 0, fmt.Errorf("failed to find match for builds object in %s", scriptBlock)
	}

	values := []map[string]any{}
	err = json.Unmarshal([]byte(exprMatches[1]), &values)
	if err != nil {
		return 0, fmt.Errorf("unable to unmarshal regex match to json object: %w", err)
	}

	return strconv.Atoi(values[1]["ID"].(string))
}

func appendUniqueContainers(name string, containers []v1.Container, images map[string]string) map[string]string {
	for _, container := range containers {
		if _, ok := images[container.Image]; !ok {
			images[container.Image] = name + "/" + container.Name
		}
	}
	return images
}
