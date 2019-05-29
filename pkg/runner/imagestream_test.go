package runner

import (
	"testing"

	imagev1 "github.com/openshift/api/image/v1"
	"github.com/openshift/client-go/image/clientset/versioned/fake"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetLatestImageStreamTag(t *testing.T) {
	tagName := "latest"
	expectedFromName := "quay.io/run/tests"

	// copy default runner
	def := *DefaultRunner
	r := &def

	// create example ImageStream
	imageStream := &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.ImageStreamName,
			Namespace: r.ImageStreamNamespace,
		},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				{
					Name: tagName,
					From: &kubev1.ObjectReference{
						Name: expectedFromName,
					},
				},
			},
		},
	}

	// set client to use example ImageStream
	streamClient := fake.NewSimpleClientset(imageStream)
	r.Image = streamClient

	// confirm tag From name
	if fromName, err := r.getLatestImageStreamTag(); err != nil {
		t.Fatalf("encountered error getting tag From name: %v", err)
	} else if fromName != expectedFromName {
		t.Fatalf("expected '%s' not '%s'", expectedFromName, fromName)
	}
}
