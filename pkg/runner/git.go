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
func (repos Repos) ConfigurePod(pod *kubev1.Pod) {
	if pod == nil {
		return
	}

	for _, r := range repos {
		// add init container
		pod.Spec.InitContainers = append(pod.Spec.InitContainers, r.Container())

		// configure volume
		pod.Spec.Volumes = append(pod.Spec.Volumes, r.Volume())

		// mount volume on each container
		for i := range pod.Spec.Containers {
			pod.Spec.Containers[i].VolumeMounts = append(pod.Spec.Containers[i].VolumeMounts, r.VolumeMount())
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
	return kubev1.Container{
		Name:  r.Name,
		Image: GitImage,
		Args:  []string{"clone", r.URL, tmpClonePath},
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
