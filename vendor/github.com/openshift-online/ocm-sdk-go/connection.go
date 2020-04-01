/*
Copyright (c) 2018 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This file contains the implementations of the Builder and Connection objects.

package sdk

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"

	"github.com/dgrijalva/jwt-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/openshift-online/ocm-sdk-go/accountsmgmt"
	"github.com/openshift-online/ocm-sdk-go/authorizations"
	"github.com/openshift-online/ocm-sdk-go/clustersmgmt"
	"github.com/openshift-online/ocm-sdk-go/servicelogs"
)

// Default values:
const (
	// #nosec G101
	DefaultTokenURL     = "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"
	DefaultClientID     = "cloud-services"
	DefaultClientSecret = ""
	DefaultURL          = "https://api.openshift.com"
	DefaultAgent        = "OCM/" + Version
)

// DefaultScopes is the ser of scopes used by default:
var DefaultScopes = []string{
	"openid",
}

// ConnectionBuilder contains the configuration and logic needed to create connections to
// `api.openshift.com`. Don't create instances of this type directly, use the NewConnectionBuilder
// function instead.
type ConnectionBuilder struct {
	// Basic attributes:
	logger           Logger
	trustedCAs       *x509.CertPool
	insecure         bool
	tokenURL         string
	clientID         string
	clientSecret     string
	apiURL           string
	agent            string
	user             string
	password         string
	tokens           []string
	scopes           []string
	transportWrapper TransportWrapper

	// Metrics:
	subsystem string
}

// TransportWrapper is a wrapper for a transport of type http.RoundTripper.
// Creating a transport wrapper, enables to preform actions and manipulations on the transport
// request and response.
type TransportWrapper func(http.RoundTripper) http.RoundTripper

// Connection contains the data needed to connect to the `api.openshift.com`. Don't create instances
// of this type directly, use the builder instead.
type Connection struct {
	// Basic attributes:
	closed       bool
	logger       Logger
	trustedCAs   *x509.CertPool
	insecure     bool
	client       *http.Client
	tokenURL     *url.URL
	clientID     string
	clientSecret string
	apiURL       *url.URL
	agent        string
	user         string
	password     string
	tokenMutex   *sync.Mutex
	tokenParser  *jwt.Parser
	accessToken  *jwt.Token
	refreshToken *jwt.Token
	scopes       []string

	// Metrics:
	tokenCountMetric    *prometheus.CounterVec
	tokenDurationMetric *prometheus.HistogramVec
	callCountMetric     *prometheus.CounterVec
	callDurationMetric  *prometheus.HistogramVec
}

// NewConnectionBuilder creates an builder that knows how to create connections with the default
// configuration.
func NewConnectionBuilder() *ConnectionBuilder {
	return new(ConnectionBuilder)
}

// Logger sets the logger that will be used by the connection. By default it uses the Go `log`
// package, and with the debug level disabled and the rest enabled. If you need to change that you
// can create a logger and pass it to this method. For example:
//
//	// Create a logger with the debug level enabled:
//	logger, err := client.NewGoLoggerBuilder().
//		Debug(true).
//		Build()
//	if err != nil {
//		panic(err)
//	}
//
//	// Create the connection:
//	cl, err := client.NewConnectionBuilder().
//		Logger(logger).
//		Build()
//	if err != nil {
//		panic(err)
//	}
//
// You can also build your own logger, implementing the Logger interface.
func (b *ConnectionBuilder) Logger(logger Logger) *ConnectionBuilder {
	b.logger = logger
	return b
}

// TokenURL sets the URL that will be used to request OpenID access tokens. The default is
// `https://sso.redhat.com/auth/realms/cloud-services/protocol/openid-connect/token`.
func (b *ConnectionBuilder) TokenURL(url string) *ConnectionBuilder {
	b.tokenURL = url
	return b
}

// Client sets OpenID client identifier and secret that will be used to request OpenID tokens. The
// default identifier is `cloud-services`. The default secret is the empty string. When these two
// values are provided and no user name and password is provided, the connection will use the client
// credentials grant to obtain the token. For example, to create a connection using the client
// credentials grant do the following:
//
//	// Use the client credentials grant:
//	connection, err := client.NewConnectionBuilder().
//		Client("myclientid", "myclientsecret").
//		Build()
//
// Note that some OpenID providers (Keycloak, for example) require the client identifier also for
// the resource owner password grant. In that case use the set only the identifier, and let the
// secret blank. For example:
//
//	// Use the resource owner password grant:
//	connection, err := client.NewConnectionBuilder().
//		User("myuser", "mypassword").
//		Client("myclientid", "").
//		Build()
//
// Note the empty client secret.
func (b *ConnectionBuilder) Client(id string, secret string) *ConnectionBuilder {
	b.clientID = id
	b.clientSecret = secret
	return b
}

// URL sets the base URL of the API gateway. The default is `https://api.openshift.com`.
func (b *ConnectionBuilder) URL(url string) *ConnectionBuilder {
	b.apiURL = url
	return b
}

// Agent sets the `User-Agent` header that the client will use in all the HTTP requests. The default
// is `OCM` followed by an slash and the version of the client, for example `OCM/0.0.0`.
func (b *ConnectionBuilder) Agent(agent string) *ConnectionBuilder {
	b.agent = agent
	return b
}

// User sets the user name and password that will be used to request OpenID access tokens. When
// these two values are provided the connection will use the resource owner password grant type to
// obtain the token. For example:
//
//	// Use the resource owner password grant:
//	connection, err := client.NewConnectionBuilder().
//		User("myuser", "mypassword").
//		Build()
//
// Note that some OpenID providers (Keycloak, for example) require the client identifier also for
// the resource owner password grant. In that case use the set only the identifier, and let the
// secret blank. For example:
//
//	// Use the resource owner password grant:
//	connection, err := client.NewConnectionBuilder().
//		User("myuser", "mypassword").
//		Client("myclientid", "").
//		Build()
//
// Note the empty client secret.
func (b *ConnectionBuilder) User(name string, password string) *ConnectionBuilder {
	b.user = name
	b.password = password
	return b
}

// Scopes sets the OpenID scopes that will be included in the token request. The default is to use
// the `openid` scope. If this method is used then that default will be completely replaced, so you
// will need to specify it explicitly if you want to use it. For example, if you want to add the
// scope 'myscope' without loosing the default you will have to do something like this:
//
//	// Create a connection with the default 'openid' scope and some additional scopes:
//	connection, err := client.NewConnectionBuilder().
//		User("myuser", "mypassword").
//		Scopes("openid", "myscope", "yourscope").
//		Build()
//
// If you just want to use the default 'openid' then there is no need to use this method.
func (b *ConnectionBuilder) Scopes(values ...string) *ConnectionBuilder {
	b.scopes = append(b.scopes, values...)
	return b
}

// Tokens sets the OpenID tokens that will be used to authenticate. Multiple types of tokens are
// accepted, and used according to their type. For example, you can pass a single access token, or
// an access token and a refresh token, or just a refresh token. If no token is provided then the
// connection will the user name and password or the client identifier and client secret (see the
// User and Client methods) to request new ones.
//
// If the connection is created with these tokens and no user or client credentials, it will
// stop working when both tokens expire. That can happen, for example, if the connection isn't used
// for a period of time longer than the life of the refresh token.
func (b *ConnectionBuilder) Tokens(tokens ...string) *ConnectionBuilder {
	b.tokens = append(b.tokens, tokens...)
	return b
}

// TrustedCAs sets the certificate pool that contains the certificate authorities that will be
// trusted by the connection. If this isn't explicitly specified then the client will trust the
// certificate authorities trusted by default by the system.
func (b *ConnectionBuilder) TrustedCAs(value *x509.CertPool) *ConnectionBuilder {
	b.trustedCAs = value
	return b
}

// Insecure enables insecure communication with the server. This disables verification of TLS
// certificates and host names and it isn't recommended for a production environment.
func (b *ConnectionBuilder) Insecure(flag bool) *ConnectionBuilder {
	b.insecure = flag
	return b
}

// TransportWrapper allows setting a transportWrapper layer into the connection for capturing and
// manipulating the request or response.
func (b *ConnectionBuilder) TransportWrapper(transportWrapper TransportWrapper) *ConnectionBuilder {
	b.transportWrapper = transportWrapper
	return b
}

// Metrics sets the name of the subsystem that will be used by the connection to register metrics
// with Prometheus. If this isn't explicitly specified, or if it is an empty string, then no metrics
// will be registered. For example, if the value is `api_outbound` then the following metrics will
// be registered:
//
//	api_outbound_request_count - Number of API requests sent.
//	api_outbound_request_duration_sum - Total time to send API requests, in seconds.
//	api_outbound_request_duration_count - Total number of API requests measured.
//	api_outbound_request_duration_bucket - Number of API requests organized in buckets.
//	api_outbound_token_request_count - Number of token requests sent.
//	api_outbound_token_request_duration_sum - Total time to send token requests, in seconds.
//	api_outbound_token_request_duration_count - Total number of token requests measured.
//	api_outbound_token_request_duration_bucket - Number of token requests organized in buckets.
//
// The duration buckets metrics contain an `le` label that indicates the upper bound. For example if
// the `le` label is `1` then the value will be the number of requests that were processed in less
// than one second.
//
// The API request metrics have the following labels:
//
//	method - Name of the HTTP method, for example GET or POST.
//	path - Request path, for example /api/clusters_mgmt/v1/clusters.
//	code - HTTP response code, for example 200 or 500.
//
// To calculate the average request duration during the last 10 minutes, for example, use a
// Prometheus expression like this:
//
//      rate(api_outbound_request_duration_sum[10m]) / rate(api_outbound_request_duration_count[10m])
//
// In order to reduce the cardinality of the metrics the path label is modified to remove the
// identifiers of the objects. For example, if the original path is .../clusters/123 then it will
// be replaced by .../clusters/-, and the values will be accumulated. The line returned by the
// metrics server will be like this:
//
//      api_outbound_request_count{code="200",method="GET",path="/api/clusters_mgmt/v1/clusters/-"} 56
//
// The meaning of that is that there were a total of 56 requests to get specific clusters,
// independently of the specific identifier of the cluster.
//
// The token request metrics will contain the following labels:
//
//      code - HTTP response code, for example 200 or 500.
//
// The value of the `code` label will be zero when sending the request failed without a response
// code, for example if it wasn't possible to open the connection, or if there was a timeout waiting
// for the response.
//
// Note that setting this attribute is not enough to have metrics published, you also need to
// create and start a metrics server, as described in the documentation of the Prometheus library.
func (b *ConnectionBuilder) Metrics(value string) *ConnectionBuilder {
	b.subsystem = value
	return b
}

// Build uses the configuration stored in the builder to create a new connection. The builder can be
// reused to create multiple connections with the same configuration. It returns a pointer to the
// connection, and an error if something fails when trying to create it.
//
// This operation is potentially lengthy, as it may require network communications. Consider using a
// context and the BuildContext method.
func (b *ConnectionBuilder) Build() (connection *Connection, err error) {
	return b.BuildContext(context.Background())
}

// BuildContext uses the configuration stored in the builder to create a new connection. The builder
// can be reused to create multiple connections with the same configuration. It returns a pointer to
// the connection, and an error if something fails when trying to create it.
func (b *ConnectionBuilder) BuildContext(ctx context.Context) (connection *Connection, err error) {
	// Check that we have some kind of credentials or a token:
	haveTokens := len(b.tokens) > 0
	havePassword := b.user != "" && b.password != ""
	haveSecret := b.clientID != "" && b.clientSecret != ""
	if !haveTokens && !havePassword && !haveSecret {
		err = fmt.Errorf(
			"either a token, and user name and password or a client identifier and secret are " +
				"necessary, but none has been provided",
		)
		return
	}

	// Parse the tokens:
	tokenParser := new(jwt.Parser)
	var accessToken *jwt.Token
	var refreshToken *jwt.Token
	for i, text := range b.tokens {
		var token *jwt.Token
		token, _, err = tokenParser.ParseUnverified(text, jwt.MapClaims{})
		if err != nil {
			err = fmt.Errorf("can't parse token %d: %v", i, err)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			err = fmt.Errorf("claims of token %d are of type '%T'", i, claims)
			return
		}
		claim, ok := claims["typ"]
		if !ok {
			err = fmt.Errorf("token %d doesn't contain the 'typ' claim", i)
			return
		}
		typ, ok := claim.(string)
		if !ok {
			err = fmt.Errorf("claim 'type' of token %d is of type '%T'", i, claim)
			return
		}
		switch {
		case strings.EqualFold(typ, "Bearer"):
			accessToken = token
		case strings.EqualFold(typ, "Refresh"):
			refreshToken = token
		case strings.EqualFold(typ, "Offline"):
			refreshToken = token
		default:
			err = fmt.Errorf("type '%s' of token %d is unknown", typ, i)
			return
		}
	}

	// Create the default logger, if needed:
	if b.logger == nil {
		b.logger, err = NewGoLoggerBuilder().
			Debug(false).
			Info(true).
			Warn(true).
			Error(true).
			Build()
		if err != nil {
			err = fmt.Errorf("can't create default logger: %v", err)
			return
		}
		b.logger.Debug(ctx, "Logger wasn't provided, will use Go log")
	}

	// Set the default authentication details, if needed:
	rawTokenURL := b.tokenURL
	if rawTokenURL == "" {
		rawTokenURL = DefaultTokenURL
		b.logger.Debug(
			ctx,
			"OpenID token URL wasn't provided, will use '%s'",
			rawTokenURL,
		)
	}
	tokenURL, err := url.Parse(rawTokenURL)
	if err != nil {
		err = fmt.Errorf("can't parse token URL '%s': %v", rawTokenURL, err)
		return
	}
	clientID := b.clientID
	if clientID == "" {
		clientID = DefaultClientID
		b.logger.Debug(
			ctx,
			"OpenID client identifier wasn't provided, will use '%s'",
			clientID,
		)
	}
	clientSecret := b.clientSecret
	if clientSecret == "" {
		clientSecret = DefaultClientSecret
		b.logger.Debug(
			ctx,
			"OpenID client secret wasn't provided, will use '%s'",
			clientSecret,
		)
	}

	// Set the default authentication scopes, if needed:
	scopes := b.scopes
	if len(scopes) == 0 {
		scopes = DefaultScopes
	} else {
		scopes = make([]string, len(b.scopes))
		for i := range b.scopes {
			scopes[i] = b.scopes[i]
		}
	}

	// Set the default URL, if needed:
	rawAPIURL := b.apiURL
	if rawAPIURL == "" {
		rawAPIURL = DefaultURL
		b.logger.Debug(ctx, "URL wasn't provided, will use the default '%s'", rawAPIURL)
	}
	apiURL, err := url.Parse(rawAPIURL)
	if err != nil {
		err = fmt.Errorf("can't parse API URL '%s': %v", rawAPIURL, err)
		return
	}

	// Set the default agent, if needed:
	agent := b.agent
	if b.agent == "" {
		agent = DefaultAgent
	}

	// Create the cookie jar:
	jar, err := b.createCookieJar()
	if err != nil {
		return
	}

	// Create the transport:
	transport, err := b.createTransport()
	if err != nil {
		return
	}

	// Create the HTTP client:
	client := &http.Client{
		Jar:       jar,
		Transport: transport,
	}

	// Allocate and populate the connection object:
	connection = &Connection{
		logger:       b.logger,
		trustedCAs:   b.trustedCAs,
		insecure:     b.insecure,
		client:       client,
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		apiURL:       apiURL,
		agent:        agent,
		user:         b.user,
		password:     b.password,
		tokenParser:  tokenParser,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		scopes:       scopes,
	}

	// Create the mutex that protects token manipulations:
	connection.tokenMutex = &sync.Mutex{}

	// Register metrics:
	if b.subsystem != "" {
		err = connection.registerMetrics(b.subsystem)
		if err != nil {
			err = fmt.Errorf("can't register metrics: %v", err)
			return
		}
	}

	return
}

func (b *ConnectionBuilder) createCookieJar() (jar http.CookieJar, err error) {
	jar, err = cookiejar.New(nil)
	return
}

func (b *ConnectionBuilder) createTransport() (transport http.RoundTripper, err error) {
	// Create the raw transport:
	// #nosec 402
	transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: b.insecure,
			RootCAs:            b.trustedCAs,
		},
	}

	// If debug is enabled then wrap the raw transport with the round tripper that sends the
	// details of requests and responses to the log:
	if b.logger.DebugEnabled() {
		transport = &dumpRoundTripper{
			logger: b.logger,
			next:   transport,
		}
	}

	// Wrap the transport with the round trippers provided by the user:
	if b.transportWrapper != nil {
		transport = b.transportWrapper(transport)
	}

	return
}

// Logger returns the logger that is used by the connection.
func (c *Connection) Logger() Logger {
	return c.logger
}

// TokenURL returns the URL that the connection is using request OpenID access tokens.
func (c *Connection) TokenURL() string {
	return c.tokenURL.String()
}

// Client returns OpenID client identifier and secret that the connection is using to request OpenID
// access tokens.
func (c *Connection) Client() (id, secret string) {
	return c.clientID, c.clientSecret
}

// URL returns the base URL of the API gateway.
func (c *Connection) URL() string {
	return c.apiURL.String()
}

// Agent returns the `User-Agent` header that the client is using for all HTTP requests.
func (c *Connection) Agent() string {
	return c.agent
}

// User returns the user name and password that the is using to request OpenID access tokens.
func (c *Connection) User() (user, password string) {
	return c.user, c.password
}

// Scopes returns the OpenID scopes that the connection is using to request OpenID access tokens.
func (c *Connection) Scopes() []string {
	result := make([]string, len(c.scopes))
	copy(result, c.scopes)
	return result
}

// TrustedCAs sets returns the certificate pool that contains the certificate authorities that are
// trusted by the connection.
func (c *Connection) TrustedCAs() *x509.CertPool {
	return c.trustedCAs
}

// Insecure returns the flag that indicates if insecure communication with the server is enabled.
func (c *Connection) Insecure() bool {
	return c.insecure
}

// AccountsMgmt returns the client for the accounts management service.
func (c *Connection) AccountsMgmt() *accountsmgmt.Client {
	return accountsmgmt.NewClient(c, "/api/accounts_mgmt", "/api/accounts_mgmt")
}

// ClustersMgmt returns the client for the clusters management service.
func (c *Connection) ClustersMgmt() *clustersmgmt.Client {
	return clustersmgmt.NewClient(c, "/api/clusters_mgmt", "/api/clusters_mgmt")
}

// Authorizations returns the client for the authorizations service.
func (c *Connection) Authorizations() *authorizations.Client {
	return authorizations.NewClient(c, "/api/authorizations", "/api/authorizations")
}

// ServiceLogs returns the client for the logs service.
func (c *Connection) ServiceLogs() *servicelogs.Client {
	return servicelogs.NewClient(c, "/api/service_logs", "/api/service_logs")
}

// Close releases all the resources used by the connection. It is very important to always close it
// once it is no longer needed, as otherwise those resources may be leaked. Trying to use a
// connection that has been closed will result in a error.
func (c *Connection) Close() error {
	err := c.checkClosed()
	if err != nil {
		return err
	}
	c.closed = true
	return nil
}

func (c *Connection) checkClosed() error {
	if c.closed {
		return fmt.Errorf("connection is closed")
	}
	return nil
}
