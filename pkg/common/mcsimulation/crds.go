// Package mcsimulation provides Management Cluster simulation for HCP operator
// testing. It installs minimal CRDs onto a ROSA Classic cluster so that
// operators that watch HyperShift resources can be tested without a real
// Management Cluster.
package mcsimulation

import (
	"context"
	"fmt"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
)

// Minimal CRD definitions with x-kubernetes-preserve-unknown-fields: true
// so that any spec/status content is accepted without full schema validation.
// Sourced from route-monitor-operator e2e tests.
const (
	HostedControlPlaneCRDYAML = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: hostedcontrolplanes.hypershift.openshift.io
spec:
  group: hypershift.openshift.io
  names:
    kind: HostedControlPlane
    listKind: HostedControlPlaneList
    plural: hostedcontrolplanes
    shortNames:
    - hcp
    singular: hostedcontrolplane
  scope: Namespaced
  versions:
  - name: v1beta1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        x-kubernetes-preserve-unknown-fields: true
    subresources:
      status: {}
`

	VpcEndpointCRDYAML = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: vpcendpoints.avo.openshift.io
spec:
  group: avo.openshift.io
  names:
    kind: VpcEndpoint
    listKind: VpcEndpointList
    plural: vpcendpoints
    singular: vpcendpoint
  scope: Namespaced
  versions:
  - name: v1alpha2
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        x-kubernetes-preserve-unknown-fields: true
    subresources:
      status: {}
`
)

// BuiltinCRDs maps short aliases (used in config) to embedded CRD YAML.
// Keys: "hcp" (HostedControlPlane), "vpce" (VpcEndpoint).
var BuiltinCRDs = map[string]string{
	"hcp":  HostedControlPlaneCRDYAML,
	"vpce": VpcEndpointCRDYAML,
}

var crdGVR = schema.GroupVersionResource{
	Group:    "apiextensions.k8s.io",
	Version:  "v1",
	Resource: "customresourcedefinitions",
}

// EnsureCRDsInstalled resolves a list of CRD aliases and installs any that are
// not already present on the cluster. Forbidden errors are collected and returned
// together so callers get a single actionable message with all missing CRDs.
func EnsureCRDsInstalled(ctx context.Context, dynClient dynamic.Interface, crdAliases []string) error {
	var forbidden []string

	for _, alias := range crdAliases {
		alias = strings.TrimSpace(alias)
		if alias == "" {
			continue
		}

		crdYAML, ok := BuiltinCRDs[alias]
		if !ok {
			return fmt.Errorf("unknown CRD alias %q (available: hcp, vpce)", alias)
		}

		// Decode the embedded YAML into an unstructured object to get the CRD name.
		obj := &unstructured.Unstructured{}
		if err := yaml.NewYAMLOrJSONDecoder(strings.NewReader(crdYAML), 4096).Decode(obj); err != nil {
			return fmt.Errorf("decoding CRD YAML for %q: %w", alias, err)
		}
		name := obj.GetName()

		// Check existence first to avoid unnecessary write attempts.
		if _, err := dynClient.Resource(crdGVR).Get(ctx, name, metav1.GetOptions{}); err == nil {
			klog.V(2).InfoS("CRD already exists, skipping", "name", name)
			continue
		} else if !apierrors.IsNotFound(err) {
			return fmt.Errorf("checking CRD %s: %w", name, err)
		}

		// Install the CRD.
		klog.V(2).InfoS("Installing CRD", "name", name)
		if _, err := dynClient.Resource(crdGVR).Create(ctx, obj, metav1.CreateOptions{}); err != nil {
			switch {
			case apierrors.IsAlreadyExists(err):
				// Another process created it concurrently; still wait for Established below.
				klog.V(2).InfoS("CRD appeared concurrently, waiting for Established", "name", name)
			case apierrors.IsForbidden(err):
				klog.InfoS("Warning: permission denied installing CRD", "name", name)
				forbidden = append(forbidden, name)
				continue
			default:
				return fmt.Errorf("creating CRD %s: %w", name, err)
			}
		}

		// Wait for the API server to mark the CRD as Established before returning.
		// Operators that detect CRD presence at startup will fail if it isn't ready.
		if err := waitForCRDEstablished(ctx, dynClient, name, 30*time.Second); err != nil {
			return fmt.Errorf("CRD %s not Established after install: %w", name, err)
		}
		klog.V(2).InfoS("CRD installed and Established", "name", name)
	}

	if len(forbidden) > 0 {
		return fmt.Errorf(
			"insufficient permissions to install CRDs: [%s] — grant cluster-admin or pre-install them",
			strings.Join(forbidden, ", "),
		)
	}
	return nil
}

// waitForCRDEstablished polls until the CRD's Established condition is True.
func waitForCRDEstablished(ctx context.Context, dynClient dynamic.Interface, name string, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, 2*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		obj, err := dynClient.Resource(crdGVR).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		conditions, found, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
		if !found {
			return false, nil
		}
		for _, c := range conditions {
			cond, ok := c.(map[string]interface{})
			if ok && cond["type"] == "Established" && cond["status"] == "True" {
				return true, nil
			}
		}
		return false, nil
	})
}
