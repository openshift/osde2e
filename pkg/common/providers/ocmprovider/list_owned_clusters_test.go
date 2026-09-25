package ocmprovider

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	ocm "github.com/openshift-online/ocm-sdk-go"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
)

var testJWTKey *rsa.PrivateKey

func init() {
	var err error
	testJWTKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
}

func makeTestToken() string {
	claims := jwt.MapClaims{
		"typ": "Bearer",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iss": "https://sso.redhat.com/auth/realms/redhat-external",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key"
	signed, err := token.SignedString(testJWTKey)
	if err != nil {
		panic(err)
	}
	return signed
}

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"access_token":"%s","token_type":"Bearer","expires_in":900}`, makeTestToken())
}

func newTestProvider(t *testing.T, handler http.Handler) *OCMProvider {
	t.Helper()
	viper.Set(config.Addons.SkipAddonList, true)
	t.Cleanup(func() { viper.Set(config.Addons.SkipAddonList, false) })

	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)

	conn, err := ocm.NewConnectionBuilder().
		URL(server.URL).
		TokenURL(server.URL+"/token").
		Client("test", "test").
		TransportWrapper(func(_ http.RoundTripper) http.RoundTripper {
			return server.Client().Transport
		}).
		Insecure(true).
		Build()
	if err != nil {
		t.Fatalf("failed to build OCM connection: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	return &OCMProvider{
		env:              "test",
		conn:             conn,
		clusterCache:     make(map[string]*spi.Cluster),
		credentialCache:  make(map[string]string),
		versionGateLabel: "api.openshift.com/gate-ocp",
	}
}

func accountHandler(id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"kind":"Account","id":"%s","username":"cicd-sa"}`, id)
	}
}

func clusterJSON(id, name string) string {
	return fmt.Sprintf(`{
		"kind": "Cluster",
		"id": "%s",
		"href": "/api/clusters_mgmt/v1/clusters/%s",
		"name": "%s",
		"state": "ready",
		"region": {"id": "us-east-1"},
		"cloud_provider": {"id": "aws"},
		"properties": {"MadeByOSDe2e": "true"},
		"version": {"id": "openshift-v4.14.0", "channel_group": "stable"}
	}`, id, id, name)
}

// clusterListHandler returns items on page=1 and an empty list on subsequent pages,
// matching what ListClusters' pagination loop expects.
func clusterListHandler(total int, itemsJSON string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		page := r.URL.Query().Get("page")
		if page != "" && page != "1" {
			fmt.Fprintf(w, `{"kind":"ClusterList","page":%s,"size":0,"total":%d,"items":[]}`, page, total)
			return
		}
		fmt.Fprintf(w, `{"kind":"ClusterList","page":1,"size":%d,"total":%d,"items":[%s]}`, total, total, itemsJSON)
	}
}

func TestListOwnedClusters_FiltersOutUnownedClusters(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/accounts_mgmt/v1/current_account", accountHandler("acct-123"))

	mux.HandleFunc("/api/accounts_mgmt/v1/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		search := r.URL.Query().Get("search")
		if !strings.Contains(search, "acct-123") {
			t.Errorf("unexpected subscription search query: %s", search)
		}
		fmt.Fprint(w, `{
			"kind": "SubscriptionList",
			"page": 1, "size": 2, "total": 2,
			"items": [
				{"kind":"Subscription","id":"sub-1","cluster_id":"cluster-aaa"},
				{"kind":"Subscription","id":"sub-2","cluster_id":"cluster-bbb"}
			]
		}`)
	})

	items := strings.Join([]string{
		clusterJSON("cluster-aaa", "osde2e-ours1"),
		clusterJSON("cluster-bbb", "osde2e-ours2"),
		clusterJSON("cluster-ccc", "osde2e-rosa-theirs"),
	}, ",")
	mux.HandleFunc("/api/clusters_mgmt/v1/clusters", clusterListHandler(3, items))
	mux.HandleFunc("/token", tokenHandler)

	provider := newTestProvider(t, mux)

	clusters, err := provider.ListOwnedClusters("properties.MadeByOSDe2e='true'")
	if err != nil {
		t.Fatalf("ListOwnedClusters returned error: %v", err)
	}

	if len(clusters) != 2 {
		t.Fatalf("expected 2 owned clusters, got %d", len(clusters))
	}

	ids := map[string]bool{}
	for _, c := range clusters {
		ids[c.ID()] = true
	}

	if !ids["cluster-aaa"] {
		t.Error("expected cluster-aaa to be included")
	}
	if !ids["cluster-bbb"] {
		t.Error("expected cluster-bbb to be included")
	}
	if ids["cluster-ccc"] {
		t.Error("expected cluster-ccc (not owned) to be filtered out")
	}
}

func TestListOwnedClusters_NoneOwned(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/accounts_mgmt/v1/current_account", accountHandler("acct-empty"))

	mux.HandleFunc("/api/accounts_mgmt/v1/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"kind": "SubscriptionList",
			"page": 1, "size": 0, "total": 0,
			"items": []
		}`)
	})

	items := strings.Join([]string{
		clusterJSON("cluster-x", "osde2e-rosa1"),
		clusterJSON("cluster-y", "osde2e-rosa2"),
	}, ",")
	mux.HandleFunc("/api/clusters_mgmt/v1/clusters", clusterListHandler(2, items))
	mux.HandleFunc("/token", tokenHandler)

	provider := newTestProvider(t, mux)

	clusters, err := provider.ListOwnedClusters("properties.MadeByOSDe2e='true'")
	if err != nil {
		t.Fatalf("ListOwnedClusters returned error: %v", err)
	}

	if len(clusters) != 0 {
		t.Fatalf("expected 0 clusters, got %d", len(clusters))
	}
}

func TestListOwnedClusters_PaginatedSubscriptions(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/accounts_mgmt/v1/current_account", accountHandler("acct-paged"))

	subsCallCount := 0
	mux.HandleFunc("/api/accounts_mgmt/v1/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		subsCallCount++
		if subsCallCount == 1 {
			fmt.Fprint(w, `{
				"kind": "SubscriptionList",
				"page": 1, "size": 1, "total": 2,
				"items": [
					{"kind":"Subscription","id":"sub-1","cluster_id":"cluster-page1"}
				]
			}`)
		} else {
			fmt.Fprint(w, `{
				"kind": "SubscriptionList",
				"page": 2, "size": 1, "total": 2,
				"items": [
					{"kind":"Subscription","id":"sub-2","cluster_id":"cluster-page2"}
				]
			}`)
		}
	})

	items := strings.Join([]string{
		clusterJSON("cluster-page1", "osde2e-p1"),
		clusterJSON("cluster-page2", "osde2e-p2"),
		clusterJSON("cluster-other", "osde2e-other"),
	}, ",")
	mux.HandleFunc("/api/clusters_mgmt/v1/clusters", clusterListHandler(3, items))
	mux.HandleFunc("/token", tokenHandler)

	provider := newTestProvider(t, mux)

	clusters, err := provider.ListOwnedClusters("properties.MadeByOSDe2e='true'")
	if err != nil {
		t.Fatalf("ListOwnedClusters returned error: %v", err)
	}

	if len(clusters) != 2 {
		t.Fatalf("expected 2 clusters from paginated subscriptions, got %d", len(clusters))
	}

	ids := map[string]bool{}
	for _, c := range clusters {
		ids[c.ID()] = true
	}

	if !ids["cluster-page1"] || !ids["cluster-page2"] {
		t.Error("expected both paginated cluster IDs to be included")
	}
	if ids["cluster-other"] {
		t.Error("expected cluster-other to be filtered out")
	}

	if subsCallCount != 2 {
		t.Errorf("expected 2 subscription API calls for pagination, got %d", subsCallCount)
	}
}

func TestListOwnedClusters_SubscriptionError(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/accounts_mgmt/v1/current_account", accountHandler("acct-err"))

	mux.HandleFunc("/api/accounts_mgmt/v1/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"kind":"Error","reason":"subscription service down"}`)
	})

	mux.HandleFunc("/token", tokenHandler)

	provider := newTestProvider(t, mux)

	_, err := provider.ListOwnedClusters("properties.MadeByOSDe2e='true'")
	if err == nil {
		t.Fatal("expected error when subscription listing fails, got nil")
	}
	if !strings.Contains(err.Error(), "couldn't list subscriptions") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestListOwnedClusters_ClusterListError(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/accounts_mgmt/v1/current_account", accountHandler("acct-ok"))

	mux.HandleFunc("/api/accounts_mgmt/v1/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"kind":"SubscriptionList","page":1,"size":1,"total":1,"items":[{"kind":"Subscription","id":"sub-1","cluster_id":"c1"}]}`)
	})

	mux.HandleFunc("/api/clusters_mgmt/v1/clusters", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"kind":"Error","reason":"clusters service down"}`)
	})

	mux.HandleFunc("/token", tokenHandler)

	provider := newTestProvider(t, mux)

	_, err := provider.ListOwnedClusters("properties.MadeByOSDe2e='true'")
	if err == nil {
		t.Fatal("expected error when cluster listing fails, got nil")
	}
}

func TestListOwnedClusters_AccountError(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/accounts_mgmt/v1/current_account", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"kind":"Error","reason":"internal error"}`)
	})

	mux.HandleFunc("/token", tokenHandler)

	provider := newTestProvider(t, mux)

	_, err := provider.ListOwnedClusters("properties.MadeByOSDe2e='true'")
	if err == nil {
		t.Fatal("expected error when CurrentAccount fails, got nil")
	}
	if !strings.Contains(err.Error(), "couldn't get current account") {
		t.Errorf("unexpected error message: %v", err)
	}
}
