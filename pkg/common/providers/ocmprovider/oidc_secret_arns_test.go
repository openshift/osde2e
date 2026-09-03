package ocmprovider

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func clusterJSONWithOIDC(id, name, oidcJSON string) string {
	return fmt.Sprintf(`{
		"kind": "Cluster",
		"id": "%s",
		"href": "/api/clusters_mgmt/v1/clusters/%s",
		"name": "%s",
		"state": "ready",
		"region": {"id": "us-east-1"},
		"cloud_provider": {"id": "aws"},
		"properties": {"MadeByOSDe2e": "true"},
		"version": {"id": "openshift-v4.14.0", "channel_group": "stable"},
		"aws": {"sts": {"oidc_config": %s}}
	}`, id, id, name, oidcJSON)
}

func ownedClustersMux(t *testing.T, accountID string, ownedIDs []string, clusters []string, clusterGET http.HandlerFunc, oidcList http.HandlerFunc) http.Handler {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/accounts_mgmt/v1/current_account", accountHandler(accountID))

	items := make([]string, 0, len(ownedIDs))
	for _, id := range ownedIDs {
		items = append(items, fmt.Sprintf(`{"kind":"Subscription","id":"sub-%s","cluster_id":"%s"}`, id, id))
	}
	mux.HandleFunc("/api/accounts_mgmt/v1/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"kind":"SubscriptionList","page":1,"size":%d,"total":%d,"items":[%s]}`,
			len(items), len(items), strings.Join(items, ","))
	})

	mux.HandleFunc("/api/clusters_mgmt/v1/clusters", clusterListHandler(len(clusters), strings.Join(clusters, ",")))
	if clusterGET != nil {
		mux.HandleFunc("/api/clusters_mgmt/v1/clusters/", clusterGET)
	}
	if oidcList != nil {
		mux.HandleFunc("/api/clusters_mgmt/v1/oidc_configs", oidcList)
	}
	mux.HandleFunc("/token", tokenHandler)
	return mux
}

func TestListOwnedOIDCSecretARNs_FromClusterAndLookup(t *testing.T) {
	const (
		inlineARN = "arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-aaa"
		lookupARN = "arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-bbb"
	)

	clusters := []string{
		clusterJSON("cluster-aaa", "osde2e-ours1"),
		clusterJSON("cluster-bbb", "osde2e-ours2"),
		clusterJSON("cluster-ccc", "osde2e-theirs"),
		clusterJSON("cluster-ddd", "osde2e-nosts"),
	}

	mux := ownedClustersMux(t, "acct-123",
		[]string{"cluster-aaa", "cluster-bbb", "cluster-ddd"},
		clusters,
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/clusters_mgmt/v1/clusters/"), "/")
			switch id {
			case "cluster-aaa":
				_, _ = fmt.Fprint(w, clusterJSONWithOIDC("cluster-aaa", "osde2e-ours1",
					fmt.Sprintf(`{"id":"oidc-aaa","secret_arn":"%s","managed":false}`, inlineARN)))
			case "cluster-bbb":
				_, _ = fmt.Fprint(w, clusterJSONWithOIDC("cluster-bbb", "osde2e-ours2", `{"id":"oidc-bbb"}`))
			case "cluster-ddd":
				_, _ = fmt.Fprint(w, clusterJSON("cluster-ddd", "osde2e-nosts"))
			default:
				w.WriteHeader(http.StatusNotFound)
				_, _ = fmt.Fprintf(w, `{"kind":"Error","reason":"cluster %s not found"}`, id)
			}
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			page := r.URL.Query().Get("page")
			if page != "" && page != "1" {
				_, _ = fmt.Fprint(w, `{"kind":"OidcConfigList","page":2,"size":0,"total":2,"items":[]}`)
				return
			}
			_, _ = fmt.Fprintf(w, `{
				"kind":"OidcConfigList","page":1,"size":2,"total":2,
				"items":[
					{"kind":"OidcConfig","id":"oidc-bbb","secret_arn":"%s","managed":false},
					{"kind":"OidcConfig","id":"oidc-ccc","secret_arn":"arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-theirs","managed":false}
				]
			}`, lookupARN)
		},
	)

	provider := newTestProvider(t, mux)
	arns, err := provider.ListOwnedOIDCSecretARNs(context.Background(), "properties.MadeByOSDe2e='true'")
	if err != nil {
		t.Fatalf("ListOwnedOIDCSecretARNs returned error: %v", err)
	}
	if !arns[inlineARN] {
		t.Fatalf("missing inline cluster SecretArn: %v", arns)
	}
	if !arns[lookupARN] {
		t.Fatalf("missing looked-up oidc_configs SecretArn: %v", arns)
	}
	if arns["arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-theirs"] {
		t.Fatalf("unowned cluster OIDC ARN should not be included: %v", arns)
	}
	if len(arns) != 2 {
		t.Fatalf("arns = %v, want 2 entries", arns)
	}
}

func TestListOwnedOIDCSecretARNs_None(t *testing.T) {
	mux := ownedClustersMux(t, "acct-empty", nil, nil, nil, nil)
	provider := newTestProvider(t, mux)

	arns, err := provider.ListOwnedOIDCSecretARNs(context.Background(), "properties.MadeByOSDe2e='true'")
	if err != nil {
		t.Fatalf("ListOwnedOIDCSecretARNs returned error: %v", err)
	}
	if arns == nil {
		t.Fatal("expected empty map, got nil")
	}
	if len(arns) != 0 {
		t.Fatalf("arns = %v, want empty", arns)
	}
}

func TestListOwnedOIDCSecretARNs_ManagedHasNoARN(t *testing.T) {
	clusters := []string{clusterJSON("cluster-aaa", "osde2e-ours1")}
	mux := ownedClustersMux(t, "acct-123",
		[]string{"cluster-aaa"},
		clusters,
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, clusterJSONWithOIDC("cluster-aaa", "osde2e-ours1", `{"id":"oidc-managed"}`))
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"kind":"OidcConfigList","page":1,"size":1,"total":1,"items":[{"kind":"OidcConfig","id":"oidc-managed","managed":true}]}`)
		},
	)

	provider := newTestProvider(t, mux)
	arns, err := provider.ListOwnedOIDCSecretARNs(context.Background(), "properties.MadeByOSDe2e='true'")
	if err != nil {
		t.Fatalf("ListOwnedOIDCSecretARNs returned error: %v", err)
	}
	if len(arns) != 0 {
		t.Fatalf("managed OIDC should not contribute a CCS secret ARN, got %v", arns)
	}
}

func TestListOwnedOIDCSecretARNs_OIDCListError(t *testing.T) {
	clusters := []string{clusterJSON("cluster-aaa", "osde2e-ours1")}
	mux := ownedClustersMux(t, "acct-123",
		[]string{"cluster-aaa"},
		clusters,
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, clusterJSONWithOIDC("cluster-aaa", "osde2e-ours1", `{"id":"oidc-aaa"}`))
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(w, `{"kind":"Error","reason":"oidc configs down"}`)
		},
	)

	provider := newTestProvider(t, mux)
	_, err := provider.ListOwnedOIDCSecretARNs(context.Background(), "properties.MadeByOSDe2e='true'")
	if err == nil {
		t.Fatal("expected error when oidc_configs listing fails, got nil")
	}
	if !strings.Contains(err.Error(), "list oidc configs") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestListOwnedOIDCSecretARNs_UnresolvedOIDCConfig(t *testing.T) {
	clusters := []string{clusterJSON("cluster-aaa", "osde2e-ours1")}
	mux := ownedClustersMux(t, "acct-123",
		[]string{"cluster-aaa"},
		clusters,
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, clusterJSONWithOIDC("cluster-aaa", "osde2e-ours1", `{"id":"oidc-missing"}`))
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"kind":"OidcConfigList","page":1,"size":0,"total":0,"items":[]}`)
		},
	)

	provider := newTestProvider(t, mux)
	_, err := provider.ListOwnedOIDCSecretARNs(context.Background(), "properties.MadeByOSDe2e='true'")
	if err == nil {
		t.Fatal("expected error when live cluster OIDC config cannot be resolved, got nil")
	}
	if !strings.Contains(err.Error(), "oidc-missing") {
		t.Errorf("unexpected error message: %v", err)
	}
}
