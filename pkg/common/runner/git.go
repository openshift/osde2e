package runner

import (
	kubev1 "k8s.io/api/core/v1"
)

const (
	// GitImage is the Docker image used to invoke Git.
	GitImage = "docker.io/alpine/git@sha256:8f5659025d83a60e9d140123bb1b27f3c334578aef10d002da4e5848580f1a6c"

	// tmpClonePath is the path cloned to by the container.
	tmpClonePath = "/git"
)

// Repos can modify a Pod to clone each contained GitRepo.
type Repos []GitRepo

// ConfigurePod modifies the given Pod to clone the Repo and make it available to any containers.
func (repos Repos) ConfigurePod(podSpec *kubev1.PodSpec) {
	if podSpec == nil {
		return
	}

	for _, r := range repos {
		// add init container
		podSpec.InitContainers = append(podSpec.InitContainers, r.Container())

		// configure volume
		podSpec.Volumes = append(podSpec.Volumes, r.Volume())

		// mount volume on each container
		for i := range podSpec.Containers {
			podSpec.Containers[i].VolumeMounts = append(podSpec.Containers[i].VolumeMounts, r.VolumeMount())
		}
	}
}

// GitRepo specifies a repository to be cloned and how it should be unpacked.
type GitRepo struct {
	// Name is used to identify the cloned repository.
	Name string

	// URL where the repository to be cloned is located.
	URL string

	// MountPath is the path where the cloned repository should be mounted.
	MountPath string

	// Branch is the branch to mount
	Branch string
}

// VolumeMount configured to mount the cloned repository in the primary container.
func (r GitRepo) VolumeMount() kubev1.VolumeMount {
	return kubev1.VolumeMount{
		Name:      r.Name,
		MountPath: r.MountPath,
	}
}

// Container configured to clone the specified repository. Typically used as an init container.
func (r GitRepo) Container() kubev1.Container {
	var args []string

	// Clone a specific branch if specified
	if r.Branch != "" {
		args = []string{"clone", "--single-branch", "-b", r.Branch, r.URL, tmpClonePath}
	} else {
		args = []string{"clone", r.URL, tmpClonePath}
	}

	return kubev1.Container{
		Name:  r.Name,
		Image: GitImage,
		Args:  args,
		VolumeMounts: []kubev1.VolumeMount{
			{
				Name:      r.Name,
				MountPath: tmpClonePath,
			},
		},
	}
}

// Volume configured as empty-dir to hold clone data.
func (r GitRepo) Volume() kubev1.Volume {
	return kubev1.Volume{
		Name: r.Name,
		VolumeSource: kubev1.VolumeSource{
			EmptyDir: &kubev1.EmptyDirVolumeSource{},
		},
	}
}
