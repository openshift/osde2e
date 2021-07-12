package verify

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/onsi/ginkgo"
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

		ginkgo.BeforeEach(func() {
			var err error
			oauthcfg, err = h.Cfg().ConfigV1().OAuths().Get(context.TODO(), "cluster", metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
		})

		ginkgo.It("should include token max age", func() {
			tokenConfig := oauthcfg.Spec.TokenConfig
			Expect(tokenConfig).ShouldNot(BeNil(), "tokenConfig should be set")
			Expect(tokenConfig.AccessTokenMaxAgeSeconds).ToNot(BeZero(), "access token max age setting should not be zero")
		})

		util.OnSupportedVersionIt(util.Version460, h, "should include token inactivity timeout", func() {
			tokenConfig := oauthcfg.Spec.TokenConfig
			Expect(tokenConfig).ShouldNot(BeNil(), "tokenConfig should be set")
			Expect(tokenConfig.AccessTokenInactivityTimeoutSeconds).ToNot(BeZero(), "access token idle timeout setting should not be zero")
		})

	})

	ginkgo.PContext("oauth token timeout", func() {

		var user *userv1.User
		var client *oauthv1.OAuthClient

		ginkgo.BeforeEach(func() {
			user, _ = createUser("osde2e-token-user-"+util.RandomStr(5), nil, nil, h)
			Expect(user).ToNot(BeNil())
			client = createOAuthClient(getOAuthClient(oauthTokensTestClientName, h), h)
			Expect(client).ToNot(BeNil())
		})

		util.OnSupportedVersionIt(util.Version460, h, "should be present on oauthaccesstokens", func() {
			_, oauthAccessToken := simulateLogin(user, client, h)
			Expect(oauthAccessToken.ExpiresIn).ToNot(BeZero(), "oauthaccesstoken expiry time should not be zero")
			Expect(oauthAccessToken.InactivityTimeoutSeconds).ToNot(BeZero(), "oauthaccesstoken idle timeout should not be zero")
		})

		util.OnSupportedVersionIt(util.Version460, h, "should not affect active sessions", func() {
			bearerToken, _ := simulateLogin(user, client, h)
			tokenCheck := verifyUserToken(bearerToken, user, h)
			Expect(tokenCheck()).To(BeTrue(), "bearer token should be valid")
			Consistently(tokenCheck, oauthTokensTestIdleTimeout+time.Minute, time.Minute).
				Should(BeTrue(), "bearer token should still be valid")
		}, (oauthTokensTestIdleTimeout + (2 * time.Minute)).Seconds())

		util.OnSupportedVersionIt(util.Version460, h, "should end idle sessions", func() {
			bearerToken, _ := simulateLogin(user, client, h)
			tokenCheck := verifyUserToken(bearerToken, user, h)
			Expect(tokenCheck()).To(BeTrue(), "bearer token should be valid")
			time.Sleep(oauthTokensTestIdleTimeout + time.Minute)
			Expect(tokenCheck()).To(BeFalse(), "bearer token should no longer be valid")
		}, (oauthTokensTestIdleTimeout + (2 * time.Minute)).Seconds())

		ginkgo.AfterEach(func() {
			deleteUser(user.Name, h)
			deleteOAuthClient(client.Name, h)
		})

	})

})

// simulates normal oauth login flow by creating and redeeming an authorization code
func simulateLogin(user *userv1.User, client *oauthv1.OAuthClient, h *helper.H) (bearerToken string, oauthAccessToken *oauthv1.OAuthAccessToken) {
	code := createAuthorizeToken(client, user, h)
	Expect(code).ToNot(BeNil(), "should be able to create oauthauthorizetokens")

	bearerToken = exchangeToken(code, h)
	Expect(bearerToken).ToNot(BeEmpty(), "should be able to redeem authorization code")

	Eventually(func() *oauthv1.OAuthAccessToken {
		oauthAccessToken = findOAuthAccessToken(code, h)
		return oauthAccessToken
	}, time.Minute, time.Second).ShouldNot(BeNil(), "should have an oauthaccesstoken")

	return bearerToken, oauthAccessToken
}

// exchanges an authorisation code for an access token
func exchangeToken(code *oauthv1.OAuthAuthorizeToken, h *helper.H) (token string) {
	Expect(http.PostForm((&url.URL{
		Scheme: "https",
		User:   url.UserPassword(code.ClientName, ""),
		Host:   oauthRoute(h).Spec.Host,
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
func createAuthorizeToken(client *oauthv1.OAuthClient, user *userv1.User, h *helper.H) *oauthv1.OAuthAuthorizeToken {
	code, err := h.OAuth().OauthV1().OAuthAuthorizeTokens().Create(context.TODO(), &oauthv1.OAuthAuthorizeToken{
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
func findOAuthAccessToken(code *oauthv1.OAuthAuthorizeToken, h *helper.H) *oauthv1.OAuthAccessToken {
	tokens, err := h.OAuth().OauthV1().OAuthAccessTokens().List(context.TODO(), metav1.ListOptions{
		FieldSelector: "authorizeToken==" + code.Name,
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(tokens.Items).To(HaveLen(1))
	return &tokens.Items[0]
}

func getOAuthClient(name string, h *helper.H) *oauthv1.OAuthClient {
	client, err := h.OAuth().OauthV1().OAuthClients().Get(context.TODO(), name, metav1.GetOptions{})
	Expect(err).ToNot(HaveOccurred())
	return client
}

// creates an oauthclient with an inactivity timeout
func createOAuthClient(tpl *oauthv1.OAuthClient, h *helper.H) *oauthv1.OAuthClient {
	timeoutSeconds := int32(oauthTokensTestIdleTimeout.Seconds())
	client, err := h.OAuth().OauthV1().OAuthClients().Create(context.TODO(), &oauthv1.OAuthClient{
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

func deleteOAuthClient(name string, h *helper.H) {
	err := h.OAuth().OauthV1().OAuthClients().Delete(context.Background(), name, metav1.DeleteOptions{})
	Expect(err).ToNot(HaveOccurred())
}

func verifyUserToken(token string, user *userv1.User, h *helper.H) func() bool {
	return func() bool {
		tokenuser, err := h.WithToken(token).User().UserV1().Users().Get(context.TODO(), "~", metav1.GetOptions{})
		if err != nil {
			return false
		}
		return user.Name == tokenuser.Name
	}
}
