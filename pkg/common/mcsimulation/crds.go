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

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

// Minimal stub CRDs with x-kubernetes-preserve-unknown-fields: true.
// These accept any spec/status content without full schema validation,
// keeping test fixtures lightweight.
var builtinCRDs = map[string]*apiextensionsv1.CustomResourceDefinition{
	"hcp": {
		ObjectMeta: metav1.ObjectMeta{
			Name: "hostedcontrolplanes.hypershift.openshift.io",
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "hypershift.openshift.io",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:       "HostedControlPlane",
				ListKind:   "HostedControlPlaneList",
				Plural:     "hostedcontrolplanes",
				Singular:   "hostedcontrolplane",
				ShortNames: []string{"hcp"},
			},
			Scope: apiextensionsv1.NamespaceScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{{
				Name:    "v1beta1",
				Served:  true,
				Storage: true,
				Schema: &apiextensionsv1.CustomResourceValidation{
					OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
						Type:                   "object",
						XPreserveUnknownFields: ptr.To(true),
					},
				},
				Subresources: &apiextensionsv1.CustomResourceSubresources{
					Status: &apiextensionsv1.CustomResourceSubresourceStatus{},
				},
			}},
		},
	},
	"vpce": {
		ObjectMeta: metav1.ObjectMeta{
			Name: "vpcendpoints.avo.openshift.io",
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "avo.openshift.io",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:     "VpcEndpoint",
				ListKind: "VpcEndpointList",
				Plural:   "vpcendpoints",
				Singular: "vpcendpoint",
			},
			Scope: apiextensionsv1.NamespaceScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{{
				Name:    "v1alpha2",
				Served:  true,
				Storage: true,
				Schema: &apiextensionsv1.CustomResourceValidation{
					OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
						Type:                   "object",
						XPreserveUnknownFields: ptr.To(true),
					},
				},
				Subresources: &apiextensionsv1.CustomResourceSubresources{
					Status: &apiextensionsv1.CustomResourceSubresourceStatus{},
				},
			}},
		},
	},
}

// BuiltinCRDAliases returns the known alias names for documentation/validation.
func BuiltinCRDAliases() []string {
	aliases := make([]string, 0, len(builtinCRDs))
	for k := range builtinCRDs {
		aliases = append(aliases, k)
	}
	return aliases
}

// EnsureCRDsInstalled resolves a list of CRD aliases and installs any that are
// not already present on the cluster. Forbidden errors are collected and returned
// together so callers get a single actionable message with all missing CRDs.
func EnsureCRDsInstalled(ctx context.Context, extClient apiextensionsclient.Interface, crdAliases []string) error {
	crdClient := extClient.ApiextensionsV1().CustomResourceDefinitions()
	var forbidden []string

	for _, alias := range crdAliases {
		alias = strings.TrimSpace(alias)
		if alias == "" {
			continue
		}

		crd, ok := builtinCRDs[alias]
		if !ok {
			return fmt.Errorf("unknown CRD alias %q (available: %s)", alias, strings.Join(BuiltinCRDAliases(), ", "))
		}
		name := crd.Name

		if _, err := crdClient.Get(ctx, name, metav1.GetOptions{}); err == nil {
			klog.V(2).InfoS("CRD already exists, skipping", "name", name)
			continue
		} else if !apierrors.IsNotFound(err) {
			return fmt.Errorf("checking CRD %s: %w", name, err)
		}

		klog.V(2).InfoS("Installing CRD", "name", name)
		if _, err := crdClient.Create(ctx, crd, metav1.CreateOptions{}); err != nil {
			switch {
			case apierrors.IsAlreadyExists(err):
				klog.V(2).InfoS("CRD appeared concurrently, waiting for Established", "name", name)
			case apierrors.IsForbidden(err):
				klog.InfoS("Warning: permission denied installing CRD", "name", name)
				forbidden = append(forbidden, name)
				continue
			default:
				return fmt.Errorf("creating CRD %s: %w", name, err)
			}
		}

		if err := waitForCRDEstablished(ctx, extClient, name, 30*time.Second); err != nil {
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
func waitForCRDEstablished(ctx context.Context, extClient apiextensionsclient.Interface, name string, timeout time.Duration) error {
	crdClient := extClient.ApiextensionsV1().CustomResourceDefinitions()
	return wait.PollUntilContextTimeout(ctx, 2*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		crd, err := crdClient.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, c := range crd.Status.Conditions {
			if c.Type == apiextensionsv1.Established && c.Status == apiextensionsv1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	})
}
