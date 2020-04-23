package debug

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-github/v31/github"
	"github.com/kylelemons/godebug/diff"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GenerateDiff attempts to pull a dependency list from a previous job (job, jobID) and generate a diff against a provided string
func GenerateDiff(baseURL, jobName, jobID, phase, dependencies string) (string, error) {
	parsedJobID, err := strconv.Atoi(jobID)
	if err != nil {
		return "", err
	}
	resp, err := http.Get(fmt.Sprintf("%s/%s/%d/artifacts/%s/dependencies.txt", baseURL, jobName, parsedJobID-1, phase))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return diff.Diff(string(body), dependencies), nil
}

// GenerateDependencies creates a list of images and the MCC hash
func GenerateDependencies(kube kubernetes.Interface) (dependencies string, err error) {
	pods, err := kube.CoreV1().Pods("").List(metav1.ListOptions{})
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
	images = make(map[string]string)

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

	return "", fmt.Errorf("No commits found for openshift/machine-cluster-config")
}

func appendUniqueContainers(name string, containers []v1.Container, images map[string]string) map[string]string {
	for _, container := range containers {
		if _, ok := images[container.Image]; !ok {
			images[container.Image] = name + "/" + container.Name
		}
	}
	return images
}
