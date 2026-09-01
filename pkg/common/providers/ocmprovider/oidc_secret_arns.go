package ocmprovider

import (
	"context"
	"fmt"

	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	ocmclient "github.com/openshift/osde2e-common/pkg/clients/ocm"
	"github.com/openshift/osde2e/pkg/common/spi"
)

// ListOIDCSecretARNs returns AWS Secrets Manager ARNs for unmanaged ROSA OIDC
// private keys used by the given clusters. Callers must treat a nil map as
// "skip set unavailable" and an empty map as "listed successfully; no live OIDC keys".
func (o *OCMProvider) ListOIDCSecretARNs(ctx context.Context, clusters []*spi.Cluster) (map[string]bool, error) {
	ocmClusters := make([]*v1.Cluster, 0, len(clusters))
	for _, cluster := range clusters {
		if cluster == nil || cluster.ID() == "" {
			continue
		}
		ocmCluster, err := o.GetOCMCluster(cluster.ID())
		if err != nil {
			return nil, fmt.Errorf("get cluster %s for oidc secret arn: %w", cluster.ID(), err)
		}
		ocmClusters = append(ocmClusters, ocmCluster)
	}
	return ocmclient.OIDCSecretARNs(ctx, o.conn, ocmClusters)
}

// ListOwnedOIDCSecretARNs returns AWS Secrets Manager ARNs for unmanaged ROSA OIDC
// private keys used by owned clusters matching query. Callers must treat a nil map as
// "skip set unavailable" and an empty map as "listed successfully; no live OIDC keys".
func (o *OCMProvider) ListOwnedOIDCSecretARNs(ctx context.Context, query string) (map[string]bool, error) {
	clusters, err := o.ListOwnedClusters(query)
	if err != nil {
		return nil, err
	}
	return o.ListOIDCSecretARNs(ctx, clusters)
}
