package verify

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	oauthv1 "github.com/openshift/api/oauth/v1"
	userv1 "github.com/openshift/api/user/v1"

	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var oauthTokensTestName string = "[Suite: e2e] [OSD] OAuth tokens"
var oauthTokensTestClientName = "openshift-challenging-client"
var oauthTokensTestIdleTimeout time.Duration = 5 * time.Minute // minimum accepted inactivity timeout

var _ = ginkgo.Describe(oauthTokensTestName, func() {

	h := helper.New()

	ginkgo.PContext("global token config", func() {

		var oauthcfg *configv1.OAuth

		ginkgo.BeforeEach(func(ctx context.Context) {
			var err error
			oauthcfg, err = h.Cfg().ConfigV1().OAuths().Get(ctx, "cluster", metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
		})

		util.GinkgoIt("should include token max age", func(_ context.Context) {
			tokenConfig := oauthcfg.Spec.TokenConfig
			Expect(tokenConfig).ShouldNot(BeNil(), "tokenConfig should be set")
			Expect(tokenConfig.AccessTokenMaxAgeSeconds).ToNot(BeZero(), "access token max age setting should not be zero")
		})

		util.OnSupportedVersionIt(util.Version460, h, "should include token inactivity timeout", func(_ context.Context) {
			tokenConfig := oauthcfg.Spec.TokenConfig
			Expect(tokenConfig).ShouldNot(BeNil(), "tokenConfig should be set")
			Expect(tokenConfig.AccessTokenInactivityTimeoutSeconds).ToNot(BeZero(), "access token idle timeout setting should not be zero")
		}, 5)

	})

	ginkgo.PContext("oauth token timeout", func() {

		var user *userv1.User
		var client *oauthv1.OAuthClient

		ginkgo.BeforeEach(func(ctx context.Context) {
			user, _ = createUser(ctx, "osde2e-token-user-"+util.RandomStr(5), nil, nil, h)
			Expect(user).ToNot(BeNil())
			client = createOAuthClient(ctx, getOAuthClient(ctx, oauthTokensTestClientName, h), h)
			Expect(client).ToNot(BeNil())
		})

		util.OnSupportedVersionIt(util.Version460, h, "should be present on oauthaccesstokens", func(ctx context.Context) {
			_, oauthAccessToken := simulateLogin(ctx, user, client, h)
			Expect(oauthAccessToken.ExpiresIn).ToNot(BeZero(), "oauthaccesstoken expiry time should not be zero")
			Expect(oauthAccessToken.InactivityTimeoutSeconds).ToNot(BeZero(), "oauthaccesstoken idle timeout should not be zero")
		}, 30)

		util.OnSupportedVersionIt(util.Version460, h, "should not affect active sessions", func(ctx context.Context) {
			bearerToken, _ := simulateLogin(ctx, user, client, h)
			tokenCheck := verifyUserToken(ctx, bearerToken, user, h)
			Expect(tokenCheck()).To(BeTrue(), "bearer token should be valid")
			Consistently(tokenCheck, oauthTokensTestIdleTimeout+time.Minute, time.Minute).
				Should(BeTrue(), "bearer token should still be valid")
		}, (oauthTokensTestIdleTimeout + (2 * time.Minute)).Seconds())

		util.OnSupportedVersionIt(util.Version460, h, "should end idle sessions", func(ctx context.Context) {
			bearerToken, _ := simulateLogin(ctx, user, client, h)
			tokenCheck := verifyUserToken(ctx, bearerToken, user, h)
			Expect(tokenCheck()).To(BeTrue(), "bearer token should be valid")
			time.Sleep(oauthTokensTestIdleTimeout + time.Minute)
			Expect(tokenCheck()).To(BeFalse(), "bearer token should no longer be valid")
		}, (oauthTokensTestIdleTimeout + (2 * time.Minute)).Seconds())

		ginkgo.AfterEach(func(ctx context.Context) {
			deleteUser(ctx, user.Name, h)
			deleteOAuthClient(ctx, client.Name, h)
		})

	})

})

// simulates normal oauth login flow by creating and redeeming an authorization code
func simulateLogin(ctx context.Context, user *userv1.User, client *oauthv1.OAuthClient, h *helper.H) (bearerToken string, oauthAccessToken *oauthv1.OAuthAccessToken) {
	code := createAuthorizeToken(ctx, client, user, h)
	Expect(code).ToNot(BeNil(), "should be able to create oauthauthorizetokens")

	bearerToken = exchangeToken(ctx, code, h)
	Expect(bearerToken).ToNot(BeEmpty(), "should be able to redeem authorization code")

	Eventually(func() *oauthv1.OAuthAccessToken {
		oauthAccessToken = findOAuthAccessToken(ctx, code, h)
		return oauthAccessToken
	}, time.Minute, time.Second).ShouldNot(BeNil(), "should have an oauthaccesstoken")

	return bearerToken, oauthAccessToken
}

// exchanges an authorisation code for an access token
func exchangeToken(ctx context.Context, code *oauthv1.OAuthAuthorizeToken, h *helper.H) (token string) {
	Expect(http.PostForm((&url.URL{
		Scheme: "https",
		User:   url.UserPassword(code.ClientName, ""),
		Host:   oauthRoute(ctx, h).Spec.Host,
		Path:   "/oauth/token",
	}).String(), url.Values{
		"code":          []string{code.Name},
		"grant_type":    []string{"authorization_code"},
		"code_verifier": []string{code.CodeChallenge},
		"redirect_uri":  []string{code.RedirectURI},
	})).Should(And(HaveHTTPStatus(http.StatusOK), WithTransform(func(r *http.Response) bool {
		type T map[string]interface{}
		return Expect(ioutil.ReadAll(r.Body)).To(WithTransform(func(b []byte) (t T) {
			Expect(json.Unmarshal(b, &t)).To(Succeed())
			return t
		}, And(HaveKey("access_token"), WithTransform(func(t T) string {
			Expect(t["access_token"]).To(BeAssignableToTypeOf("string"))
			token = t["access_token"].(string)
			return token
		}, Not(BeEmpty())))))
	}, BeTrue())))
	return token
}

// creates an oauthauthorizetoken for the given oauthclient (must be a challenging client)
func createAuthorizeToken(ctx context.Context, client *oauthv1.OAuthClient, user *userv1.User, h *helper.H) *oauthv1.OAuthAuthorizeToken {
	code, err := h.OAuth().OauthV1().OAuthAuthorizeTokens().Create(ctx, &oauthv1.OAuthAuthorizeToken{
		ObjectMeta:          metav1.ObjectMeta{Name: util.RandomStr(32)},
		ClientName:          client.Name,
		CodeChallenge:       util.RandomStr(48),
		CodeChallengeMethod: "plain",
		ExpiresIn:           60,
		RedirectURI:         client.RedirectURIs[0],
		Scopes:              []string{"user:full"},
		UserName:            user.Name,
		UserUID:             string(user.UID),
	}, metav1.CreateOptions{})
	Expect(err).ToNot(HaveOccurred())
	return code
}

// find an oauthaccesstoken corresponding to an oauthauthorizetoken
func findOAuthAccessToken(ctx context.Context, code *oauthv1.OAuthAuthorizeToken, h *helper.H) *oauthv1.OAuthAccessToken {
	tokens, err := h.OAuth().OauthV1().OAuthAccessTokens().List(ctx, metav1.ListOptions{
		FieldSelector: "authorizeToken==" + code.Name,
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(tokens.Items).To(HaveLen(1))
	return &tokens.Items[0]
}

func getOAuthClient(ctx context.Context, name string, h *helper.H) *oauthv1.OAuthClient {
	client, err := h.OAuth().OauthV1().OAuthClients().Get(ctx, name, metav1.GetOptions{})
	Expect(err).ToNot(HaveOccurred())
	return client
}

// creates an oauthclient with an inactivity timeout
func createOAuthClient(ctx context.Context, tpl *oauthv1.OAuthClient, h *helper.H) *oauthv1.OAuthClient {
	timeoutSeconds := int32(oauthTokensTestIdleTimeout.Seconds())
	client, err := h.OAuth().OauthV1().OAuthClients().Create(ctx, &oauthv1.OAuthClient{
		ObjectMeta:                          metav1.ObjectMeta{Name: "osde2e-oauth-" + util.RandomStr(5)},
		AccessTokenInactivityTimeoutSeconds: &timeoutSeconds,
		GrantMethod:                         tpl.GrantMethod,
		RespondWithChallenges:               tpl.RespondWithChallenges,
		RedirectURIs:                        tpl.RedirectURIs,
		ScopeRestrictions:                   tpl.ScopeRestrictions,
	}, metav1.CreateOptions{})
	Expect(err).ToNot(HaveOccurred())
	return client
}

func deleteOAuthClient(ctx context.Context, name string, h *helper.H) {
	err := h.OAuth().OauthV1().OAuthClients().Delete(ctx, name, metav1.DeleteOptions{})
	Expect(err).ToNot(HaveOccurred())
}

func verifyUserToken(ctx context.Context, token string, user *userv1.User, h *helper.H) func() bool {
	return func() bool {
		tokenuser, err := h.WithToken(token).User().UserV1().Users().Get(ctx, "~", metav1.GetOptions{})
		if err != nil {
			return false
		}
		return user.Name == tokenuser.Name
	}
}

// createUser creates the given user.
// Note that it may take time for operators to reconcile the permissions of new users,
// so it's best to poll your first attempt to use the resulting user for a couple minutes.
func createUser(ctx context.Context, userName string, identities []string, groups []string, h *helper.H) (*userv1.User, error) {
	user := &userv1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: userName,
		},
		Identities: identities,
		Groups:     groups,
	}
	return h.User().UserV1().Users().Create(ctx, user, metav1.CreateOptions{})
}

func deleteUser(ctx context.Context, userName string, h *helper.H) error {
	return h.User().UserV1().Users().Delete(ctx, userName, metav1.DeleteOptions{})
}
