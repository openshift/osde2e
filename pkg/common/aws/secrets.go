package aws

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	secretsmanagertypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/openshift/osde2e/pkg/common/config"
)

const (
	leftoverNameMarker = "osde2e"
	rosaOIDCSecretPref = "rosa-private-key-oidc-"
)

// CleanupSecrets deletes leftover Secrets Manager secrets created by osde2e ROSA STS/HCP
// provision (unmanaged OIDC private keys and CAPA userdata) that are not tied to a live cluster.
// activeClusters and activeOIDCSecretARNs must be non-nil. Nil means listing was
// unavailable and cleanup is refused; empty means listing succeeded with no live entries.
// olderThan must be > 0 so in-progress OIDC keys are not removed during provision.
func (CcsAwsSession *ccsAwsSession) CleanupSecrets(ctx context.Context, activeClusters, activeOIDCSecretARNs map[string]bool, olderThan time.Duration, dryrun bool, sendSummary bool,
	errorBuilder *strings.Builder,
) (counters Counters, err error) {
	if activeClusters == nil {
		return counters, fmt.Errorf("active clusters set is required for secrets cleanup")
	}
	if activeOIDCSecretARNs == nil {
		return counters, fmt.Errorf("oidc secret arn skip set is required for secrets cleanup")
	}

	err = CcsAwsSession.GetAWSSessions()
	if err != nil {
		return counters, err
	}

	result, err := cleanupSecrets(ctx, cleanupSecretsInput{
		Config:               CcsAwsSession.cfg,
		ActiveClusters:       activeClusters,
		ActiveOIDCSecretARNs: activeOIDCSecretARNs,
		OlderThan:            olderThan,
		DryRun:               dryrun,
	}, cleanupSecretsDeps{})
	counters.Deleted = result.Deleted
	counters.Failed = result.Failed
	if sendSummary {
		for _, msg := range result.Errors {
			if errorBuilder.Len() >= config.SlackMessageLength {
				break
			}
			errorBuilder.WriteString(msg)
			if !strings.HasSuffix(msg, "\n") {
				errorBuilder.WriteByte('\n')
			}
		}
	}
	return counters, err
}

// secretIdentity is the subset of a Secrets Manager secret used for cleanup decisions.
type secretIdentity struct {
	Name        string
	ARN         string
	Description string
	TagText     string
}

// blob concatenates name, ARN, description, and tags for live-cluster skip matching.
func (s secretIdentity) blob() string {
	return strings.Join([]string{s.Name, s.ARN, s.Description, s.TagText}, " ")
}

// secretIdentityFromAWS maps an AWS Secrets Manager list entry into a secretIdentity.
func secretIdentityFromAWS(secret secretsmanagertypes.SecretListEntry) secretIdentity {
	tags := make([]string, 0, len(secret.Tags))
	for _, tag := range secret.Tags {
		tags = append(tags, aws.ToString(tag.Key)+"="+aws.ToString(tag.Value))
	}
	return secretIdentity{
		Name:        aws.ToString(secret.Name),
		ARN:         aws.ToString(secret.ARN),
		Description: aws.ToString(secret.Description),
		TagText:     strings.Join(tags, " "),
	}
}

// belongsToActiveCluster reports whether a secret belongs to a live osde2e cluster.
// Cluster names are matched as complete identifiers, so osde2e-live1 does not
// match a secret for osde2e-live12.
func belongsToActiveCluster(secret secretIdentity, activeClusters map[string]bool) (string, bool) {
	blob := secret.blob()
	for clusterName := range activeClusters {
		if clusterNameInBlob(blob, clusterName) {
			return clusterName, true
		}
	}
	return "", false
}

// clusterNameInBlob reports whether blob contains clusterName as a complete
// identifier. "osde2e-live1" matches "osde2e-live1-installer" but not "osde2e-live12".
func clusterNameInBlob(blob, clusterName string) bool {
	if clusterName == "" {
		return false
	}
	for start := 0; start <= len(blob)-len(clusterName); {
		i := strings.Index(blob[start:], clusterName)
		if i < 0 {
			return false
		}
		pos := start + i
		if identifierBoundaryAt(blob, pos) && identifierBoundaryAt(blob, pos+len(clusterName)) {
			return true
		}
		start = pos + 1
	}
	return false
}

// identifierBoundaryAt reports a word-style boundary at index i, matching Go's \b:
// start/end of string, or a transition between [A-Za-z0-9_] and any other byte.
func identifierBoundaryAt(s string, i int) bool {
	if i <= 0 || i >= len(s) {
		return true
	}
	return isWordChar(s[i-1]) != isWordChar(s[i])
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

const osde2eClusterPrefix = leftoverNameMarker + "-"

// hasOSDe2eClusterToken reports whether s contains an osde2e- cluster name token.
func hasOSDe2eClusterToken(s string) bool {
	s = strings.ToLower(s)
	prefix := osde2eClusterPrefix
	for start := 0; start <= len(s)-len(prefix); {
		i := strings.Index(s[start:], prefix)
		if i < 0 {
			return false
		}
		pos := start + i
		if identifierBoundaryAt(s, pos) {
			return true
		}
		start = pos + 1
	}
	return false
}

// tagHasOSDe2eCluster reports whether any tag key or value contains an osde2e-
// cluster name token. CAPA puts the CAPI cluster name in the key
// (sigs.k8s.io/cluster-api-provider-aws/cluster/osde2e-abcd1=owned), not the value.
func tagHasOSDe2eCluster(tagText string) bool {
	for _, field := range strings.Fields(tagText) {
		key, value, ok := strings.Cut(field, "=")
		if hasOSDe2eClusterToken(key) {
			return true
		}
		if ok && hasOSDe2eClusterToken(value) {
			return true
		}
	}
	return false
}

// isLeftoverSecret reports whether a secret is in a class this janitor may delete.
// Unmanaged ROSA OIDC private keys are named rosa-private-key-oidc-*. That prefix
// is the leak: abort/timeout often leaves the key without an osde2e marker.
// Live keys are kept by the OCM SecretArn skip set and --older-than, not by
// requiring the marker. Other secrets (CAPA bootstrap userdata) are leftovers
// when a tag key or value is an osde2e- cluster name. CAPA stores that name in
// the tag key; a user-defined secret whose name merely contains "osde2e" is not
// deleted.
func isLeftoverSecret(secret secretIdentity) bool {
	if isUnmanagedOIDCSecret(secret) {
		return true
	}
	return tagHasOSDe2eCluster(secret.TagText)
}

// matchesOIDCSecretARN reports whether this secret is a live unmanaged OIDC private key.
// OCM SecretArn is the source of truth. AWS list ARNs often append a 6-character suffix
// to the name portion; match exact ARNs or the same secret name plus that suffix.
func matchesOIDCSecretARN(secret secretIdentity, activeOIDCSecretARNs map[string]bool) bool {
	if activeOIDCSecretARNs == nil || secret.ARN == "" {
		return false
	}
	if activeOIDCSecretARNs[secret.ARN] {
		return true
	}
	for arn := range activeOIDCSecretARNs {
		if oidcARNsMatch(secret.ARN, arn) {
			return true
		}
	}
	return false
}

// oidcARNsMatch reports whether a listed secret ARN is the same secret as an OCM skip ARN.
// AWS list ARNs often append a hyphen and 6 random characters to the name portion.
func oidcARNsMatch(secretARN, skipARN string) bool {
	if secretARN == "" || skipARN == "" {
		return false
	}
	if secretARN == skipARN {
		return true
	}
	a := secretNameFromARN(secretARN)
	b := secretNameFromARN(skipARN)
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}
	return awsSecretARNNameHasSuffix(a, b) || awsSecretARNNameHasSuffix(b, a)
}

// secretNameFromARN returns the secret name portion of a Secrets Manager ARN.
func secretNameFromARN(arn string) string {
	const marker = ":secret:"
	i := strings.LastIndex(arn, marker)
	if i < 0 {
		return ""
	}
	return arn[i+len(marker):]
}

// awsSecretARNNameHasSuffix reports whether full is base plus AWS's hyphen and 6-character suffix.
func awsSecretARNNameHasSuffix(full, base string) bool {
	return strings.HasPrefix(full, base+"-") && len(full) == len(base)+7
}

// isUnmanagedOIDCSecret reports whether the secret name is an unmanaged ROSA OIDC private key.
func isUnmanagedOIDCSecret(secret secretIdentity) bool {
	return strings.HasPrefix(secret.Name, rosaOIDCSecretPref)
}

// shouldSkipSecret reports whether a leftover must be kept. A nil OIDC skip set is
// fail-closed for unmanaged ROSA OIDC keys; an empty set means no live OIDC keys.
func shouldSkipSecret(secret secretIdentity, activeClusters map[string]bool, activeOIDCSecretARNs map[string]bool) (string, bool) {
	// Fail closed: never delete unmanaged OIDC keys unless the caller supplied an OCM ARN set.
	if isUnmanagedOIDCSecret(secret) && activeOIDCSecretARNs == nil {
		return "oidc skip set not provided", true
	}
	if matchesOIDCSecretARN(secret, activeOIDCSecretARNs) {
		return secret.ARN, true
	}
	return belongsToActiveCluster(secret, activeClusters)
}

// tooNewToDelete reports whether a leftover is within the age floor.
// Unmanaged OIDC keys are created before the cluster appears in OCM; without a
// cutoff those in-progress keys look like orphans. A missing CreatedDate is
// treated as too new. olderThan <= 0 disables the floor.
func tooNewToDelete(created *time.Time, olderThan time.Duration, now time.Time) (string, bool) {
	if olderThan <= 0 {
		return "", false
	}
	if created == nil {
		return "unknown created date", true
	}
	if now.Sub(*created) < olderThan {
		return "newer than cutoff", true
	}
	return "", false
}

// cleanupSecretsInput configures account-wide Secrets Manager leftover cleanup.
type cleanupSecretsInput struct {
	Config aws.Config
	// ActiveClusters must be non-nil. Nil means cluster listing was unavailable
	// and cleanup is refused. An empty map means listing succeeded with no live clusters.
	ActiveClusters map[string]bool
	// ActiveOIDCSecretARNs is the OCM skip set for unmanaged ROSA OIDC keys.
	// Nil means the skip set was unavailable (fail closed for those keys).
	// An empty map means listing succeeded with no live OIDC keys.
	ActiveOIDCSecretARNs map[string]bool
	OlderThan            time.Duration
	DryRun               bool
}

type secretCleanupResult struct {
	Deleted int
	Failed  int
	Errors  []string
}

type secretsAPI interface {
	ListSecrets(ctx context.Context, params *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error)
	DeleteSecret(ctx context.Context, params *secretsmanager.DeleteSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.DeleteSecretOutput, error)
}

type regionsAPI interface {
	DescribeRegions(ctx context.Context, params *ec2.DescribeRegionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error)
}

type cleanupSecretsDeps struct {
	regions   []string
	ec2       regionsAPI
	newClient func(region string) secretsAPI
}

// cleanupSecrets walks regions and deletes leftover secrets using the given dependencies.
func cleanupSecrets(ctx context.Context, in cleanupSecretsInput, deps cleanupSecretsDeps) (secretCleanupResult, error) {
	if in.ActiveClusters == nil {
		return secretCleanupResult{}, fmt.Errorf("active clusters set is required for secrets cleanup")
	}

	regions, err := listSecretCleanupRegions(ctx, in, deps)
	if err != nil {
		return secretCleanupResult{}, err
	}

	newClient := deps.newClient
	if newClient == nil {
		newClient = func(region string) secretsAPI {
			cfg := in.Config.Copy()
			cfg.Region = region
			return secretsmanager.NewFromConfig(cfg)
		}
	}

	var result secretCleanupResult
	var regionErrs []error
	for _, region := range regions {
		regionResult, regionErr := cleanupSecretsInRegion(ctx, in, region, newClient(region))
		result.Deleted += regionResult.Deleted
		result.Failed += regionResult.Failed
		result.Errors = append(result.Errors, regionResult.Errors...)
		if regionErr != nil {
			log.Printf("Continuing secrets cleanup after error in %s: %v\n", region, regionErr)
			regionErrs = append(regionErrs, regionErr)
		}
	}
	return result, errors.Join(regionErrs...)
}

// listSecretCleanupRegions returns enabled AWS regions to scan, or the session region on DescribeRegions failure.
func listSecretCleanupRegions(ctx context.Context, in cleanupSecretsInput, deps cleanupSecretsDeps) ([]string, error) {
	if len(deps.regions) > 0 {
		return deps.regions, nil
	}

	client := deps.ec2
	if client == nil {
		client = ec2.NewFromConfig(in.Config)
	}

	out, err := client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(false),
	})
	if err != nil {
		region := in.Config.Region
		if region == "" {
			return nil, fmt.Errorf("list aws regions: %w", err)
		}
		log.Printf("DescribeRegions failed; falling back to session region %s: %v\n", region, err)
		return []string{region}, nil
	}

	regions := make([]string, 0, len(out.Regions))
	for _, region := range out.Regions {
		if name := aws.ToString(region.RegionName); name != "" {
			regions = append(regions, name)
		}
	}
	if len(regions) == 0 {
		region := in.Config.Region
		if region == "" {
			return nil, fmt.Errorf("no aws regions available for secrets cleanup")
		}
		return []string{region}, nil
	}
	return regions, nil
}

// cleanupSecretsInRegion lists Secrets Manager secrets in one region and deletes leftovers that are safe to remove.
func cleanupSecretsInRegion(ctx context.Context, in cleanupSecretsInput, region string, client secretsAPI) (secretCleanupResult, error) {
	var result secretCleanupResult
	paginator := secretsmanager.NewListSecretsPaginator(client, &secretsmanager.ListSecretsInput{
		IncludePlannedDeletion: aws.Bool(true),
	})

	for paginator.HasMorePages() {
		page, pageErr := paginator.NextPage(ctx)
		if pageErr != nil {
			return result, fmt.Errorf("list secrets in %s: %w", region, pageErr)
		}
		for _, secret := range page.SecretList {
			ident := secretIdentityFromAWS(secret)
			if secret.DeletedDate != nil {
				continue
			}
			if !isLeftoverSecret(ident) {
				continue
			}
			if reason, ok := shouldSkipSecret(ident, in.ActiveClusters, in.ActiveOIDCSecretARNs); ok {
				log.Printf("Skipping secret for live cluster or OIDC config %s: %s (%s)\n", reason, ident.Name, region)
				continue
			}
			if isUnmanagedOIDCSecret(ident) && in.OlderThan <= 0 {
				log.Printf("Skipping unmanaged OIDC secret; age floor not provided: %s (%s)\n", ident.Name, region)
				continue
			}
			if reason, ok := tooNewToDelete(secret.CreatedDate, in.OlderThan, time.Now()); ok {
				log.Printf("Skipping secret newer than cutoff (%s): %s (%s) olderThan=%s\n", reason, ident.Name, region, in.OlderThan)
				continue
			}

			log.Printf("Secret will be deleted: %s (%s)\n", ident.Name, region)
			if in.DryRun {
				result.Deleted++
				continue
			}

			_, delErr := client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
				SecretId:                   aws.String(ident.ARN),
				ForceDeleteWithoutRecovery: aws.Bool(true),
			})
			if delErr != nil {
				result.Failed++
				msg := fmt.Sprintf("secret %s (%s): not deleted: %v", ident.Name, region, delErr)
				log.Printf("Failed to delete secret %s (%s): %v\n", ident.Name, region, delErr)
				result.Errors = append(result.Errors, msg)
				continue
			}
			result.Deleted++
			log.Printf("Deleted secret: %s (%s)\n", ident.Name, region)
		}
	}

	return result, nil
}
