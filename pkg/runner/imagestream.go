package runner

import (
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// name of test ImageStream
	testImageStreamName = "tests"

	// namespace containing test ImageStream
	testImageStreamNamespace = "openshift"
)

// getLatestImageStreamTag returns the From name of the latest ImageStream tag.
func (r *Runner) getLatestImageStreamTag() (string, error) {
	return r.getImageStreamTag("latest")
}

// getImageStreamTag returns the From name of the given ImageStream tag.
func (r *Runner) getImageStreamTag(tag string) (string, error) {
	if r.Image == nil {
		return "", errors.New("client for Image must be set")
	}

	stream, err := r.Image.ImageV1().ImageStreams(r.ImageStreamNamespace).Get(r.ImageStreamName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("can't get ImageStream '%s/%s': %v", r.ImageStreamNamespace, r.ImageStreamName, err)
	}

	for _, imageTag := range stream.Spec.Tags {
		if imageTag.Name == tag {
			if imageTag.From != nil {
				return imageTag.From.Name, nil
			}
			return "", fmt.Errorf("ImageStream '%s/%s' tag '%s' has a nil From", r.ImageStreamNamespace, r.ImageStreamName, tag)
		}
	}
	return "", fmt.Errorf("ImageStream '%s/%s' does not have tag '%s'", r.ImageStreamNamespace, r.ImageStreamName, tag)
}
