package mcsimulation

import (
	"context"
	"strings"
	"testing"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

// parseCRDForFake decodes a CRD YAML constant into an unstructured object
// suitable for pre-populating a dynamicfake client.
func parseCRDForFake(t *testing.T) *unstructured.Unstructured {
	t.Helper()
	obj := &unstructured.Unstructured{}
	if err := yaml.NewYAMLOrJSONDecoder(strings.NewReader(HostedControlPlaneCRDYAML), 4096).Decode(obj); err != nil {
		t.Fatalf("parseCRDForFake: %v", err)
	}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apiextensions.k8s.io",
		Version: "v1",
		Kind:    "CustomResourceDefinition",
	})
	obj.SetResourceVersion("1")
	return obj
}

func notFoundReactor(action k8stesting.Action) (bool, runtime.Object, error) {
	return true, nil, apierrors.NewNotFound(
		schema.GroupResource{Resource: "customresourcedefinitions"},
		action.(k8stesting.GetAction).GetName(),
	)
}

func TestBuiltinCRDs_YAMLDecodes(t *testing.T) {
	for alias, crdYAML := range BuiltinCRDs {
		t.Run(alias, func(t *testing.T) {
			obj := &unstructured.Unstructured{}
			err := yaml.NewYAMLOrJSONDecoder(strings.NewReader(crdYAML), 4096).Decode(obj)
			if err != nil {
				t.Fatalf("alias %q: decode error: %v", alias, err)
			}
			if obj.GetKind() != "CustomResourceDefinition" {
				t.Errorf("alias %q: expected kind CustomResourceDefinition, got %q", alias, obj.GetKind())
			}
			if obj.GetName() == "" {
				t.Errorf("alias %q: CRD has no name", alias)
			}
		})
	}
}

func TestEnsureCRDsInstalled_UnknownAlias(t *testing.T) {
	client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	err := EnsureCRDsInstalled(context.Background(), client, []string{"nonexistent"})
	if err == nil || !strings.Contains(err.Error(), "unknown CRD alias") {
		t.Fatalf("expected 'unknown CRD alias' error, got: %v", err)
	}
}

func TestEnsureCRDsInstalled_EmptyAliases(t *testing.T) {
	client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	if err := EnsureCRDsInstalled(context.Background(), client, []string{"", "  "}); err != nil {
		t.Fatalf("expected no error for empty aliases, got: %v", err)
	}
}

func TestEnsureCRDsInstalled_AlreadyExists(t *testing.T) {
	existing := parseCRDForFake(t)
	client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), existing)
	if err := EnsureCRDsInstalled(context.Background(), client, []string{"hcp"}); err != nil {
		t.Fatalf("expected no error when CRD already exists, got: %v", err)
	}
}

func TestEnsureCRDsInstalled_Forbidden(t *testing.T) {
	client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	client.PrependReactor("get", "customresourcedefinitions", notFoundReactor)
	client.PrependReactor("create", "customresourcedefinitions", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewForbidden(schema.GroupResource{Resource: "customresourcedefinitions"}, "test", nil)
	})

	err := EnsureCRDsInstalled(context.Background(), client, []string{"hcp"})
	if err == nil || !strings.Contains(err.Error(), "insufficient permissions") {
		t.Fatalf("expected 'insufficient permissions' error, got: %v", err)
	}
}

func TestEnsureCRDsInstalled_RaceConditionAlreadyExists(t *testing.T) {
	// First Get → not found; Create → AlreadyExists (concurrent creator won the race);
	// subsequent Gets → established CRD so waitForCRDEstablished returns quickly.
	established := parseCRDForFake(t)
	if err := unstructured.SetNestedSlice(established.Object, []interface{}{
		map[string]interface{}{"type": "Established", "status": "True"},
	}, "status", "conditions"); err != nil {
		t.Fatalf("SetNestedSlice: %v", err)
	}

	getCallCount := 0
	client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
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
	crdObj := parseCRDForFake(t)
	if err := unstructured.SetNestedSlice(crdObj.Object, []interface{}{
		map[string]interface{}{"type": "Established", "status": "True"},
	}, "status", "conditions"); err != nil {
		t.Fatalf("SetNestedSlice: %v", err)
	}

	client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), crdObj)
	if err := waitForCRDEstablished(context.Background(), client, crdObj.GetName(), 5*time.Second); err != nil {
		t.Fatalf("expected no error for already-established CRD, got: %v", err)
	}
}

func TestWaitForCRDEstablished_TimesOut(t *testing.T) {
	crdObj := parseCRDForFake(t)
	client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), crdObj)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := waitForCRDEstablished(ctx, client, crdObj.GetName(), 2*time.Second); err == nil {
		t.Fatal("expected timeout error for non-established CRD")
	}
}

func TestEnsureCRDsInstalled_EstablishmentTimeout(t *testing.T) {
	// Create succeeds but CRD never reaches Established state — exercises the
	// "not Established after install" error path in EnsureCRDsInstalled.
	// First Get → NotFound (triggers install); subsequent Gets → CRD with no
	// conditions (waitForCRDEstablished keeps polling until context deadline).
	crdObj := parseCRDForFake(t)

	getCallCount := 0
	client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	client.PrependReactor("get", "customresourcedefinitions", func(action k8stesting.Action) (bool, runtime.Object, error) {
		getCallCount++
		if getCallCount == 1 {
			return notFoundReactor(action)
		}
		// Return CRD with no status conditions — never Established.
		return true, crdObj, nil
	})
	client.PrependReactor("create", "customresourcedefinitions", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, crdObj, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := EnsureCRDsInstalled(ctx, client, []string{"hcp"})
	if err == nil || !strings.Contains(err.Error(), "not Established after install") {
		t.Fatalf("expected 'not Established after install' error, got: %v", err)
	}
}
