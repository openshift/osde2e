package mcsimulation

import (
	"context"
	"strings"
	"testing"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	fakeapiextensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stesting "k8s.io/client-go/testing"
)

func hcpCRD() *apiextensionsv1.CustomResourceDefinition {
	crd := builtinCRDs["hcp"].DeepCopy()
	crd.ResourceVersion = "1"
	return crd
}

func establishedCRD() *apiextensionsv1.CustomResourceDefinition {
	crd := hcpCRD()
	crd.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{{
		Type:   apiextensionsv1.Established,
		Status: apiextensionsv1.ConditionTrue,
	}}
	return crd
}

func notFoundReactor(action k8stesting.Action) (bool, runtime.Object, error) {
	return true, nil, apierrors.NewNotFound(
		schema.GroupResource{Resource: "customresourcedefinitions"},
		action.(k8stesting.GetAction).GetName(),
	)
}

func TestBuiltinCRDs_Defined(t *testing.T) {
	for _, alias := range []string{"hcp", "vpce"} {
		t.Run(alias, func(t *testing.T) {
			crd, ok := builtinCRDs[alias]
			if !ok {
				t.Fatalf("alias %q not found in builtinCRDs", alias)
			}
			if crd.Name == "" {
				t.Error("CRD has no name")
			}
			if crd.Spec.Group == "" {
				t.Error("CRD has no group")
			}
			if len(crd.Spec.Versions) == 0 {
				t.Error("CRD has no versions")
			}
		})
	}
}

func TestEnsureCRDsInstalled_UnknownAlias(t *testing.T) {
	client := fakeapiextensions.NewClientset()
	err := EnsureCRDsInstalled(context.Background(), client, []string{"nonexistent"})
	if err == nil || !strings.Contains(err.Error(), "unknown CRD alias") {
		t.Fatalf("expected 'unknown CRD alias' error, got: %v", err)
	}
}

func TestEnsureCRDsInstalled_EmptyAliases(t *testing.T) {
	client := fakeapiextensions.NewClientset()
	if err := EnsureCRDsInstalled(context.Background(), client, []string{"", "  "}); err != nil {
		t.Fatalf("expected no error for empty aliases, got: %v", err)
	}
}

func TestEnsureCRDsInstalled_AlreadyExists(t *testing.T) {
	existing := hcpCRD()
	client := fakeapiextensions.NewClientset(existing)
	if err := EnsureCRDsInstalled(context.Background(), client, []string{"hcp"}); err != nil {
		t.Fatalf("expected no error when CRD already exists, got: %v", err)
	}
}

func TestEnsureCRDsInstalled_Forbidden(t *testing.T) {
	client := fakeapiextensions.NewClientset()
	client.PrependReactor("create", "customresourcedefinitions", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewForbidden(schema.GroupResource{Resource: "customresourcedefinitions"}, "test", nil)
	})

	err := EnsureCRDsInstalled(context.Background(), client, []string{"hcp"})
	if err == nil || !strings.Contains(err.Error(), "insufficient permissions") {
		t.Fatalf("expected 'insufficient permissions' error, got: %v", err)
	}
}

func TestEnsureCRDsInstalled_RaceConditionAlreadyExists(t *testing.T) {
	established := establishedCRD()

	getCallCount := 0
	client := fakeapiextensions.NewClientset()
	client.PrependReactor("get", "customresourcedefinitions", func(action k8stesting.Action) (bool, runtime.Object, error) {
		getCallCount++
		if getCallCount == 1 {
			return notFoundReactor(action)
		}
		return true, established, nil
	})
	client.PrependReactor("create", "customresourcedefinitions", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewAlreadyExists(schema.GroupResource{Resource: "customresourcedefinitions"}, "test")
	})

	if err := EnsureCRDsInstalled(context.Background(), client, []string{"hcp"}); err != nil {
		t.Fatalf("expected no error for concurrent race, got: %v", err)
	}
	if getCallCount < 2 {
		t.Errorf("expected waitForCRDEstablished to poll after race, got %d get calls", getCallCount)
	}
}

func TestWaitForCRDEstablished_AlreadyEstablished(t *testing.T) {
	established := establishedCRD()
	client := fakeapiextensions.NewClientset(established)
	if err := waitForCRDEstablished(context.Background(), client, established.Name, 5*time.Second); err != nil {
		t.Fatalf("expected no error for already-established CRD, got: %v", err)
	}
}

func TestWaitForCRDEstablished_TimesOut(t *testing.T) {
	crd := hcpCRD()
	client := fakeapiextensions.NewClientset(crd)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := waitForCRDEstablished(ctx, client, crd.Name, 2*time.Second); err == nil {
		t.Fatal("expected timeout error for non-established CRD")
	}
}

func TestEnsureCRDsInstalled_EstablishmentTimeout(t *testing.T) {
	crd := hcpCRD()

	getCallCount := 0
	client := fakeapiextensions.NewClientset()
	client.PrependReactor("get", "customresourcedefinitions", func(action k8stesting.Action) (bool, runtime.Object, error) {
		getCallCount++
		if getCallCount == 1 {
			return notFoundReactor(action)
		}
		return true, crd, nil
	})
	client.PrependReactor("create", "customresourcedefinitions", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, crd, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := EnsureCRDsInstalled(ctx, client, []string{"hcp"})
	if err == nil || !strings.Contains(err.Error(), "not Established after install") {
		t.Fatalf("expected 'not Established after install' error, got: %v", err)
	}
}
