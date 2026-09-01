package aws

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	secretsmanagertypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

func TestIsLeftoverSecret(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		secret secretIdentity
		want   bool
	}{
		{
			name:   "unmanaged rosa oidc without osde2e marker",
			secret: secretIdentity{Name: "rosa-private-key-oidc-abc123"},
			want:   true,
		},
		{
			name:   "unmanaged rosa oidc with osde2e description",
			secret: secretIdentity{Name: "rosa-private-key-oidc-abc123", Description: "osde2e leftover"},
			want:   true,
		},
		{
			name:   "unmanaged rosa oidc with owner tag",
			secret: secretIdentity{Name: "rosa-private-key-oidc-abc123", TagText: "owner=osde2e"},
			want:   true,
		},
		{
			name:   "user-defined name containing osde2e",
			secret: secretIdentity{Name: "osde2e-shared-token"},
			want:   false,
		},
		{
			name:   "osde2e cluster-style name without cluster tag",
			secret: secretIdentity{Name: "osde2e-abcd1-installer"},
			want:   false,
		},
		{
			name:   "osde2e only in arn",
			secret: secretIdentity{Name: "unrelated", ARN: "arn:aws:secretsmanager:us-east-1:123:secret:osde2e-abcd1"},
			want:   false,
		},
		{
			name:   "osde2e only in description",
			secret: secretIdentity{Name: "bootstrap", Description: "created for osde2e"},
			want:   false,
		},
		{
			name:   "owner tag without cluster name",
			secret: secretIdentity{Name: "bootstrap", TagText: "owner=osde2e"},
			want:   false,
		},
		{
			name:   "capa userdata name",
			secret: secretIdentity{Name: "aws.cluster.x-k8s.io/11111111-2222-3333-4444-555555555555-0"},
			want:   false,
		},
		{
			name:   "capa cluster tag",
			secret: secretIdentity{Name: "bootstrap", TagText: "cluster.x-k8s.io/cluster-name=capa-test"},
			want:   false,
		},
		{
			name:   "capa provider tag without cluster name",
			secret: secretIdentity{Name: "bootstrap", TagText: "sigs.k8s.io/cluster-api-provider-aws/cluster=owned"},
			want:   false,
		},
		{
			name:   "capa kubernetes cluster tag value owned",
			secret: secretIdentity{Name: "aws.cluster.x-k8s.io/11111111-2222-3333-4444-555555555555-0", TagText: "kubernetes.io/cluster/abc12=owned"},
			want:   false,
		},
		{
			name:   "capa tag value for osde2e cluster",
			secret: secretIdentity{Name: "bootstrap", TagText: "cluster.x-k8s.io/cluster-name=osde2e-abcd1"},
			want:   true,
		},
		{
			name:   "capa ownership tag key for osde2e cluster",
			secret: secretIdentity{Name: "aws.cluster.x-k8s.io/11111111-2222-3333-4444-555555555555-0", TagText: "sigs.k8s.io/cluster-api-provider-aws/cluster/osde2e-abcd1=owned sigs.k8s.io/cluster-api-provider-aws/role=node"},
			want:   true,
		},
		{
			name:   "capa kubernetes.io cluster tag key for osde2e cluster",
			secret: secretIdentity{Name: "aws.cluster.x-k8s.io/11111111-2222-3333-4444-555555555555-0", TagText: "kubernetes.io/cluster/osde2e-abcd1=owned"},
			want:   true,
		},
		{
			name:   "capa Name tag value for osde2e machine",
			secret: secretIdentity{Name: "aws.cluster.x-k8s.io/11111111-2222-3333-4444-555555555555-0", TagText: "Name=osde2e-abcd1-workers-abcde"},
			want:   true,
		},
		{
			name:   "unrelated secret",
			secret: secretIdentity{Name: "prod/db-password", ARN: "arn:aws:secretsmanager:us-east-1:123:secret:prod/db-password"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isLeftoverSecret(tt.secret); got != tt.want {
				t.Fatalf("isLeftoverSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBelongsToActiveCluster(t *testing.T) {
	t.Parallel()

	active := map[string]bool{"osde2e-live1": true}

	tests := []struct {
		name        string
		secret      secretIdentity
		wantCluster string
		wantSkip    bool
	}{
		{
			name:        "cluster name in secret name",
			secret:      secretIdentity{Name: "osde2e-live1-oidc"},
			wantCluster: "osde2e-live1",
			wantSkip:    true,
		},
		{
			name:        "cluster name in description",
			secret:      secretIdentity{Name: "rosa-private-key-oidc-xyz", Description: "for osde2e-live1"},
			wantCluster: "osde2e-live1",
			wantSkip:    true,
		},
		{
			name:     "orphaned osde2e secret",
			secret:   secretIdentity{Name: "osde2e-dead1-installer"},
			wantSkip: false,
		},
		{
			name:     "prefix collision with longer sibling cluster",
			secret:   secretIdentity{Name: "osde2e-live12-installer"},
			wantSkip: false,
		},
		{
			name:        "cluster name followed by hyphenated suffix",
			secret:      secretIdentity{Name: "osde2e-live1-installer"},
			wantCluster: "osde2e-live1",
			wantSkip:    true,
		},
		{
			// Cluster-name skip only. Live unmanaged OIDC keys are skipped by
			// SecretArn in shouldSkipSecret, not by belongsToActiveCluster.
			name:     "rosa oidc key without cluster name",
			secret:   secretIdentity{Name: "rosa-private-key-oidc-orphan"},
			wantSkip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cluster, skip := belongsToActiveCluster(tt.secret, active)
			if skip != tt.wantSkip {
				t.Fatalf("belongsToActiveCluster() skip = %v, want %v", skip, tt.wantSkip)
			}
			if cluster != tt.wantCluster {
				t.Fatalf("belongsToActiveCluster() cluster = %q, want %q", cluster, tt.wantCluster)
			}
		})
	}
}

func TestTooNewToDelete(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	old := now.Add(-48 * time.Hour)
	recent := now.Add(-time.Hour)

	if reason, skip := tooNewToDelete(&old, 24*time.Hour, now); skip {
		t.Fatalf("old secret skipped: %s", reason)
	}
	if _, skip := tooNewToDelete(&recent, 24*time.Hour, now); !skip {
		t.Fatal("recent secret should be skipped")
	}
	if _, skip := tooNewToDelete(nil, 24*time.Hour, now); !skip {
		t.Fatal("unknown created date should be skipped")
	}
	if _, skip := tooNewToDelete(&recent, 0, now); skip {
		t.Fatal("age floor disabled should not skip")
	}
}

func TestOIDCARNsMatch(t *testing.T) {
	t.Parallel()

	listARN := "arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-live-AbCdEf"
	ocmARN := "arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-live"
	otherARN := "arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-other-XyZ123"

	if !oidcARNsMatch(listARN, listARN) {
		t.Fatal("exact ARN should match")
	}
	if !oidcARNsMatch(listARN, ocmARN) {
		t.Fatal("AWS 6-char suffix should match OCM SecretArn without suffix")
	}
	if oidcARNsMatch(listARN, otherARN) {
		t.Fatal("different OIDC secret names must not match")
	}
	if oidcARNsMatch("arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-ab-XXXXXX",
		"arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-abc") {
		t.Fatal("name prefix collision must not match")
	}
}

func TestShouldSkipSecretByOIDCARN(t *testing.T) {
	t.Parallel()

	liveARN := "arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-live-AbCdEf"
	secret := secretIdentity{Name: "rosa-private-key-oidc-live", ARN: liveARN}

	_, skip := shouldSkipSecret(secret, map[string]bool{"osde2e-live1": true}, map[string]bool{liveARN: true})
	if !skip {
		t.Fatal("should skip live OIDC secret by exact ARN")
	}

	_, skip = shouldSkipSecret(secret, map[string]bool{"osde2e-live1": true}, map[string]bool{
		"arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-live": true,
	})
	if !skip {
		t.Fatal("should skip when OCM SecretArn is the AWS ARN without the 6-char suffix")
	}

	_, skip = shouldSkipSecret(
		secretIdentity{Name: "rosa-private-key-oidc-orphan", ARN: "arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-orphan-XyZ123"},
		map[string]bool{"osde2e-live1": true},
		map[string]bool{liveARN: true},
	)
	if skip {
		t.Fatal("should not skip orphaned OIDC key whose ARN is not in the live set")
	}

	_, skip = shouldSkipSecret(secret, map[string]bool{"osde2e-live1": true}, nil)
	if !skip {
		t.Fatal("should skip unmanaged OIDC keys when the OCM ARN set was not provided")
	}

	_, skip = shouldSkipSecret(
		secretIdentity{Name: "rosa-private-key-oidc-orphan", ARN: "arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-orphan-XyZ123"},
		map[string]bool{"osde2e-live1": true},
		map[string]bool{},
	)
	if skip {
		t.Fatal("empty OCM ARN set means no live OIDC keys; orphans must not be skipped")
	}
}

type mockSecrets struct {
	pages   [][]secretsmanagertypes.SecretListEntry
	deleted []string
	listErr error
	delErr  error
}

func (m *mockSecrets) ListSecrets(_ context.Context, _ *secretsmanager.ListSecretsInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	if len(m.pages) == 0 {
		return &secretsmanager.ListSecretsOutput{}, nil
	}
	page := m.pages[0]
	m.pages = m.pages[1:]
	return &secretsmanager.ListSecretsOutput{SecretList: page}, nil
}

func (m *mockSecrets) DeleteSecret(_ context.Context, params *secretsmanager.DeleteSecretInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.DeleteSecretOutput, error) {
	if m.delErr != nil {
		return nil, m.delErr
	}
	m.deleted = append(m.deleted, aws.ToString(params.SecretId))
	return &secretsmanager.DeleteSecretOutput{}, nil
}

type mockRegions struct {
	names []string
	err   error
}

func (m *mockRegions) DescribeRegions(_ context.Context, _ *ec2.DescribeRegionsInput, _ ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	out := &ec2.DescribeRegionsOutput{}
	for _, name := range m.names {
		out.Regions = append(out.Regions, ec2types.Region{RegionName: aws.String(name)})
	}
	return out, nil
}

func leftoverSecret(name, arn string) secretsmanagertypes.SecretListEntry {
	return secretsmanagertypes.SecretListEntry{
		Name:        aws.String(name),
		ARN:         aws.String(arn),
		Description: aws.String("osde2e leftover"),
		CreatedDate: aws.Time(time.Now().Add(-48 * time.Hour)),
	}
}

func orphanOIDCSecret() secretsmanagertypes.SecretListEntry {
	return leftoverSecret("rosa-private-key-oidc-orphan", "arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-orphan-XyZ123")
}

func mixedSecretPage() []secretsmanagertypes.SecretListEntry {
	liveOIDC := leftoverSecret("rosa-private-key-oidc-live", "arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-live-AbCdEf")
	liveNamed := leftoverSecret("osde2e-live1-oidc", "arn:aws:secretsmanager:us-east-1:1:secret:osde2e-live1-oidc")
	unrelated := leftoverSecret("prod/db-password", "arn:aws:secretsmanager:us-east-1:1:secret:prod/db-password")
	unrelated.Name = aws.String("prod/db-password")
	unrelated.Description = nil
	pending := leftoverSecret("rosa-private-key-oidc-pending", "arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-pending")
	pending.DeletedDate = aws.Time(time.Now())
	return []secretsmanagertypes.SecretListEntry{orphanOIDCSecret(), liveOIDC, liveNamed, unrelated, pending}
}

func runSecretsCleanup(secrets []secretsmanagertypes.SecretListEntry, in cleanupSecretsInput, delErr error) (*mockSecrets, secretCleanupResult, error) {
	client := &mockSecrets{
		pages:  [][]secretsmanagertypes.SecretListEntry{secrets},
		delErr: delErr,
	}
	result, err := cleanupSecrets(context.Background(), in, cleanupSecretsDeps{
		regions:   []string{"us-east-1"},
		newClient: func(string) secretsAPI { return client },
	})
	return client, result, err
}

func TestCleanupSecretsDryRun(t *testing.T) {
	t.Parallel()
	page := mixedSecretPage()
	client, result, err := runSecretsCleanup(page, cleanupSecretsInput{
		ActiveClusters:       map[string]bool{"osde2e-live1": true},
		ActiveOIDCSecretARNs: map[string]bool{aws.ToString(page[1].ARN): true},
		OlderThan:            24 * time.Hour,
		DryRun:               true,
	}, nil)
	if err != nil {
		t.Fatalf("cleanupSecrets() error = %v", err)
	}
	if result.Deleted != 1 || result.Failed != 0 {
		t.Fatalf("dry run counters = %+v, want Deleted=1 Failed=0", result)
	}
	if len(client.deleted) != 0 {
		t.Fatalf("dry run deleted %v, want none", client.deleted)
	}
}

func TestCleanupSecretsExecuteDeletesLeftoverOnly(t *testing.T) {
	t.Parallel()
	page := mixedSecretPage()
	orphan := page[0]
	client, result, err := runSecretsCleanup(page, cleanupSecretsInput{
		ActiveClusters:       map[string]bool{"osde2e-live1": true},
		ActiveOIDCSecretARNs: map[string]bool{aws.ToString(page[1].ARN): true},
		OlderThan:            24 * time.Hour,
	}, nil)
	if err != nil {
		t.Fatalf("cleanupSecrets() error = %v", err)
	}
	if result.Deleted != 1 {
		t.Fatalf("Deleted = %d, want 1", result.Deleted)
	}
	if result.Failed != 0 {
		t.Fatalf("Failed = %d, want 0", result.Failed)
	}
	if len(client.deleted) != 1 || client.deleted[0] != aws.ToString(orphan.ARN) {
		t.Fatalf("deleted = %v, want [%s]", client.deleted, aws.ToString(orphan.ARN))
	}
}

func TestCleanupSecretsDeleteFailure(t *testing.T) {
	t.Parallel()
	client, result, err := runSecretsCleanup([]secretsmanagertypes.SecretListEntry{orphanOIDCSecret()}, cleanupSecretsInput{
		ActiveClusters:       map[string]bool{},
		ActiveOIDCSecretARNs: map[string]bool{},
		OlderThan:            24 * time.Hour,
	}, errors.New("access denied"))
	if err != nil {
		t.Fatalf("cleanupSecrets() error = %v", err)
	}
	if result.Failed != 1 || result.Deleted != 0 {
		t.Fatalf("counters = %+v, want Failed=1 Deleted=0", result)
	}
	if len(result.Errors) != 1 {
		t.Fatalf("Errors = %v, want 1 entry", result.Errors)
	}
	if len(client.deleted) != 0 {
		t.Fatalf("deleted = %v, want none", client.deleted)
	}
}

func TestCleanupSecretsNilOIDCSkipSet(t *testing.T) {
	t.Parallel()
	client, result, err := runSecretsCleanup([]secretsmanagertypes.SecretListEntry{orphanOIDCSecret()}, cleanupSecretsInput{
		ActiveClusters: map[string]bool{},
	}, nil)
	if err != nil {
		t.Fatalf("cleanupSecrets() error = %v", err)
	}
	if result.Deleted != 0 || len(client.deleted) != 0 {
		t.Fatalf("nil OIDC skip set deleted %v", client.deleted)
	}
}

func TestCleanupSecretsEmptyOIDCSkipSet(t *testing.T) {
	t.Parallel()
	client, result, err := runSecretsCleanup([]secretsmanagertypes.SecretListEntry{orphanOIDCSecret()}, cleanupSecretsInput{
		ActiveClusters:       map[string]bool{},
		ActiveOIDCSecretARNs: map[string]bool{},
		OlderThan:            24 * time.Hour,
	}, nil)
	if err != nil {
		t.Fatalf("cleanupSecrets() error = %v", err)
	}
	if result.Deleted != 1 || len(client.deleted) != 1 {
		t.Fatalf("empty OIDC skip set deleted %v, want orphan", client.deleted)
	}
}

func TestCleanupSecretsKeepsLiveUnmarkedOIDC(t *testing.T) {
	t.Parallel()
	live := leftoverSecret("rosa-private-key-oidc-live", "arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-live-AbCdEf")
	live.Description = nil
	client, result, err := runSecretsCleanup([]secretsmanagertypes.SecretListEntry{live}, cleanupSecretsInput{
		ActiveClusters:       map[string]bool{},
		ActiveOIDCSecretARNs: map[string]bool{aws.ToString(live.ARN): true},
		OlderThan:            24 * time.Hour,
	}, nil)
	if err != nil {
		t.Fatalf("cleanupSecrets() error = %v", err)
	}
	if result.Deleted != 0 || len(client.deleted) != 0 {
		t.Fatalf("live unmarked OIDC deleted %v", client.deleted)
	}
}

func TestCleanupSecretsUnmarkedOIDCPrefix(t *testing.T) {
	t.Parallel()
	orphan := leftoverSecret("rosa-private-key-oidc-other", "arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-other-AbCdEf")
	orphan.Description = nil
	client, result, err := runSecretsCleanup([]secretsmanagertypes.SecretListEntry{orphan}, cleanupSecretsInput{
		ActiveClusters:       map[string]bool{},
		ActiveOIDCSecretARNs: map[string]bool{},
		OlderThan:            24 * time.Hour,
	}, nil)
	if err != nil {
		t.Fatalf("cleanupSecrets() error = %v", err)
	}
	if result.Deleted != 1 || len(client.deleted) != 1 {
		t.Fatalf("unmarked OIDC orphan deleted %v, want 1", client.deleted)
	}
}

func TestCleanupSecretsNilActiveClusters(t *testing.T) {
	t.Parallel()
	client, result, err := runSecretsCleanup([]secretsmanagertypes.SecretListEntry{orphanOIDCSecret()}, cleanupSecretsInput{
		ActiveOIDCSecretARNs: map[string]bool{},
		OlderThan:            24 * time.Hour,
	}, nil)
	if err == nil {
		t.Fatal("cleanupSecrets() error = nil, want active clusters required")
	}
	if result.Deleted != 0 || len(client.deleted) != 0 {
		t.Fatalf("nil ActiveClusters deleted %v", client.deleted)
	}
}

func TestCleanupSecretsUserDefinedNameNotDeleted(t *testing.T) {
	t.Parallel()
	user := leftoverSecret("osde2e-shared-token", "arn:aws:secretsmanager:us-east-1:1:secret:osde2e-shared-token")
	user.Description = nil
	client, result, err := runSecretsCleanup([]secretsmanagertypes.SecretListEntry{user}, cleanupSecretsInput{
		ActiveClusters:       map[string]bool{},
		ActiveOIDCSecretARNs: map[string]bool{},
		OlderThan:            24 * time.Hour,
	}, nil)
	if err != nil {
		t.Fatalf("cleanupSecrets() error = %v", err)
	}
	if result.Deleted != 0 || len(client.deleted) != 0 {
		t.Fatalf("user-defined secret deleted %v", client.deleted)
	}
}

func TestCleanupSecretsCAPAOwnershipTagKey(t *testing.T) {
	t.Parallel()
	orphan := leftoverSecret("aws.cluster.x-k8s.io/11111111-2222-3333-4444-555555555555-0", "arn:aws:secretsmanager:us-east-1:1:secret:aws.cluster.x-k8s.io/11111111-2222-3333-4444-555555555555-0")
	orphan.Description = nil
	orphan.Tags = []secretsmanagertypes.Tag{{
		Key:   aws.String("sigs.k8s.io/cluster-api-provider-aws/cluster/osde2e-dead1"),
		Value: aws.String("owned"),
	}}
	live := leftoverSecret("aws.cluster.x-k8s.io/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee-0", "arn:aws:secretsmanager:us-east-1:1:secret:aws.cluster.x-k8s.io/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee-0")
	live.Description = nil
	live.Tags = []secretsmanagertypes.Tag{{
		Key:   aws.String("sigs.k8s.io/cluster-api-provider-aws/cluster/osde2e-live1"),
		Value: aws.String("owned"),
	}}
	client, result, err := runSecretsCleanup([]secretsmanagertypes.SecretListEntry{orphan, live}, cleanupSecretsInput{
		ActiveClusters:       map[string]bool{"osde2e-live1": true},
		ActiveOIDCSecretARNs: map[string]bool{},
		OlderThan:            24 * time.Hour,
	}, nil)
	if err != nil {
		t.Fatalf("cleanupSecrets() error = %v", err)
	}
	if result.Deleted != 1 || len(client.deleted) != 1 {
		t.Fatalf("CAPA leftovers deleted %v, want 1", client.deleted)
	}
}

func TestCleanupSecretsPrefixCollision(t *testing.T) {
	t.Parallel()
	sibling := leftoverSecret("osde2e-live12-installer", "arn:aws:secretsmanager:us-east-1:1:secret:osde2e-live12-installer")
	sibling.Tags = []secretsmanagertypes.Tag{{
		Key:   aws.String("cluster.x-k8s.io/cluster-name"),
		Value: aws.String("osde2e-live12"),
	}}
	client, result, err := runSecretsCleanup([]secretsmanagertypes.SecretListEntry{sibling}, cleanupSecretsInput{
		ActiveClusters:       map[string]bool{"osde2e-live1": true},
		ActiveOIDCSecretARNs: map[string]bool{},
		OlderThan:            24 * time.Hour,
	}, nil)
	if err != nil {
		t.Fatalf("cleanupSecrets() error = %v", err)
	}
	if result.Deleted != 1 || len(client.deleted) != 1 {
		t.Fatalf("sibling leftover deleted %v, want 1", client.deleted)
	}
}

func TestCleanupSecretsNoAgeFloor(t *testing.T) {
	t.Parallel()
	client, result, err := runSecretsCleanup([]secretsmanagertypes.SecretListEntry{orphanOIDCSecret()}, cleanupSecretsInput{
		ActiveClusters:       map[string]bool{},
		ActiveOIDCSecretARNs: map[string]bool{},
	}, nil)
	if err != nil {
		t.Fatalf("cleanupSecrets() error = %v", err)
	}
	if result.Deleted != 0 || len(client.deleted) != 0 {
		t.Fatalf("OIDC delete without age floor deleted %v", client.deleted)
	}
}

func TestCleanupSecretsNewerThanCutoff(t *testing.T) {
	t.Parallel()
	recent := leftoverSecret("rosa-private-key-oidc-recent", "arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-recent-AbCdEf")
	recent.CreatedDate = aws.Time(time.Now().Add(-time.Hour))
	client, result, err := runSecretsCleanup([]secretsmanagertypes.SecretListEntry{recent}, cleanupSecretsInput{
		ActiveClusters:       map[string]bool{},
		ActiveOIDCSecretARNs: map[string]bool{},
		OlderThan:            24 * time.Hour,
	}, nil)
	if err != nil {
		t.Fatalf("cleanupSecrets() error = %v", err)
	}
	if result.Deleted != 0 || len(client.deleted) != 0 {
		t.Fatalf("recent OIDC key deleted %v", client.deleted)
	}
}

func TestCleanupSecretsUnknownCreatedDate(t *testing.T) {
	t.Parallel()
	unknown := leftoverSecret("rosa-private-key-oidc-unknown", "arn:aws:secretsmanager:us-east-1:1:secret:rosa-private-key-oidc-unknown-AbCdEf")
	unknown.CreatedDate = nil
	client, result, err := runSecretsCleanup([]secretsmanagertypes.SecretListEntry{unknown}, cleanupSecretsInput{
		ActiveClusters:       map[string]bool{},
		ActiveOIDCSecretARNs: map[string]bool{},
		OlderThan:            24 * time.Hour,
	}, nil)
	if err != nil {
		t.Fatalf("cleanupSecrets() error = %v", err)
	}
	if result.Deleted != 0 || len(client.deleted) != 0 {
		t.Fatalf("unknown-age OIDC key deleted %v", client.deleted)
	}
}

func TestCleanupSecretsContinuesAfterRegionListError(t *testing.T) {
	t.Parallel()
	failClient := &mockSecrets{listErr: errors.New("timeout")}
	okClient := &mockSecrets{
		pages: [][]secretsmanagertypes.SecretListEntry{{orphanOIDCSecret()}},
	}
	result, err := cleanupSecrets(context.Background(), cleanupSecretsInput{
		ActiveClusters:       map[string]bool{},
		ActiveOIDCSecretARNs: map[string]bool{},
		OlderThan:            24 * time.Hour,
	}, cleanupSecretsDeps{
		regions: []string{"us-east-1", "us-west-2"},
		newClient: func(region string) secretsAPI {
			if region == "us-east-1" {
				return failClient
			}
			return okClient
		},
	})
	if err == nil {
		t.Fatal("cleanupSecrets() error = nil, want region list error")
	}
	if !strings.Contains(err.Error(), "us-east-1") {
		t.Fatalf("error = %v, want us-east-1 list failure", err)
	}
	if result.Deleted != 1 || len(okClient.deleted) != 1 {
		t.Fatalf("other region deleted %v, want orphan", okClient.deleted)
	}
	if len(failClient.deleted) != 0 {
		t.Fatalf("failed region deleted %v", failClient.deleted)
	}
}

func TestListSecretCleanupRegions(t *testing.T) {
	t.Parallel()

	t.Run("uses describe regions", func(t *testing.T) {
		t.Parallel()
		got, err := listSecretCleanupRegions(context.Background(), cleanupSecretsInput{}, cleanupSecretsDeps{
			ec2: &mockRegions{names: []string{"us-east-1", "us-west-2"}},
		})
		if err != nil {
			t.Fatalf("listSecretCleanupRegions() error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("regions = %v, want 2", got)
		}
	})

	t.Run("falls back to config region", func(t *testing.T) {
		t.Parallel()
		got, err := listSecretCleanupRegions(context.Background(), cleanupSecretsInput{
			Config: aws.Config{Region: "eu-west-1"},
		}, cleanupSecretsDeps{
			ec2: &mockRegions{err: errors.New("denied")},
		})
		if err != nil {
			t.Fatalf("listSecretCleanupRegions() error = %v", err)
		}
		if len(got) != 1 || got[0] != "eu-west-1" {
			t.Fatalf("regions = %v, want [eu-west-1]", got)
		}
	})
}
