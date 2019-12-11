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

// This file contains the implementations of the methods of the connection that handle OpenID
// authentication tokens.

package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/openshift-online/ocm-sdk-go/internal"
)

// Tokens returns the access and refresh tokens that is currently in use by the connection. If it is
// necessary to request a new token because it wasn't requested yet, or because it is expired, this
// method will do it and will return an error if it fails.
//
// This operation is potentially lengthy, as it may require network communication. Consider using a
// context and the TokensContext method.
func (c *Connection) Tokens() (access, refresh string, err error) {
	return c.TokensContext(context.Background())
}

// TokensContext returns the access and refresh tokens that is currently in use by the connection.
// If it is necessary to request a new token because it wasn't requested yet, or because it is
// expired, this method will do it and will return an error if it fails.
func (c *Connection) TokensContext(ctx context.Context) (access, refresh string, err error) {
	// We need to make sure that this method isn't execute concurrently, as we will be updating
	// multiple attributes of the connection:
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	// Check the expiration times of the tokens:
	now := time.Now()
	var accessExpires bool
	var accessLeft time.Duration
	if c.accessToken != nil {
		accessExpires, accessLeft, err = tokenExpiry(c.accessToken, now)
		if err != nil {
			return
		}
	}
	var refreshExpires bool
	var refreshLeft time.Duration
	if c.refreshToken != nil {
		refreshExpires, refreshLeft, err = tokenExpiry(c.refreshToken, now)
		if err != nil {
			return
		}
	}
	if c.logger.DebugEnabled() {
		c.debugExpiry(ctx, "Bearer", c.accessToken, accessExpires, accessLeft)
		c.debugExpiry(ctx, "Refresh", c.refreshToken, refreshExpires, refreshLeft)
	}

	// If the access token is available and it isn't expired or about to expire then we can
	// return the current tokens directly:
	if c.accessToken != nil && (!accessExpires || accessLeft >= 1*time.Minute) {
		access, refresh = c.currentTokens()
		return
	}

	// At this point we know that the access token is unavailable, expired or about to expire.
	// So we need to check if we can use the refresh token to request a new one.
	if c.refreshToken != nil && (!refreshExpires || refreshLeft >= 1*time.Minute) {
		err = c.sendRefreshTokenForm(ctx)
		if err != nil {
			return
		}
		access, refresh = c.currentTokens()
		return
	}

	// Now we know that both the access and refresh tokens are unavaiable, expired or about to
	// expire. So we need to check if we have other credentials that can be used to request a
	// new token, and use them.
	havePassword := c.user != "" && c.password != ""
	haveSecret := c.clientID != "" && c.clientSecret != ""
	if havePassword || haveSecret {
		err = c.sendRequestTokenForm(ctx)
		if err != nil {
			return
		}
		access, refresh = c.currentTokens()
		return
	}

	// Here we know that the access and refresh tokens are unavailable, expired or about to
	// expire. We also know that we don't have credentials to request new ones. But we could
	// still use the refresh token if it isn't completely exired.
	if c.refreshToken != nil && refreshLeft > 0 {
		c.logger.Warn(
			ctx,
			"Refresh token expires in only %s, but there is no other mechanism to "+
				"obtain a new token, so will try to use it anyhow",
			refreshLeft,
		)
		err = c.sendRefreshTokenForm(ctx)
		if err != nil {
			return
		}
		access, refresh = c.currentTokens()
		return
	}

	// At this point we know that the access token is expired or about to expire. We know also
	// that the refresh token is unavailable or completely expired. And we know that we don't
	// have credentials to request new tokens. But we can still use the access token if it isn't
	// expired.
	if c.accessToken != nil && accessLeft > 0 {
		c.logger.Warn(
			ctx,
			"Access token expires in only %s, but there is no other mechanism to "+
				"obtain a new token, so will try to use it anyhow",
			accessLeft,
		)
		access, refresh = c.currentTokens()
		return
	}

	// There is no way to get a valid access token, so all we can do is report the failure:
	err = fmt.Errorf(
		"access and refresh tokens are unavailable or expired, and there are no " +
			"password or client secret to request new ones",
	)

	return
}

// currentTokens returns the current tokens without trying to send any request to refresh them, and
// checking that they are actually available. If they aren't available then it will return empty
// strings.
func (c *Connection) currentTokens() (access, refresh string) {
	if c.accessToken != nil {
		access = c.accessToken.Raw
	}
	if c.refreshToken != nil {
		refresh = c.refreshToken.Raw
	}
	return
}

func (c *Connection) sendRequestTokenForm(ctx context.Context) error {
	form := url.Values{}
	havePassword := c.user != "" && c.password != ""
	haveSecret := c.clientID != "" && c.clientSecret != ""
	if havePassword {
		c.logger.Debug(ctx, "Requesting new token using the password grant")
		form.Set("grant_type", "password")
		form.Set("client_id", c.clientID)
		form.Set("username", c.user)
		form.Set("password", c.password)
	} else if haveSecret {
		c.logger.Debug(ctx, "Requesting new token using the client credentials grant")
		form.Set("grant_type", "client_credentials")
		form.Set("client_id", c.clientID)
		form.Set("client_secret", c.clientSecret)
	} else {
		return fmt.Errorf(
			"either password or client secret must be provided",
		)
	}
	form.Set("scope", strings.Join(c.scopes, " "))
	return c.sendTokenForm(ctx, form)
}

func (c *Connection) sendRefreshTokenForm(ctx context.Context) error {
	c.logger.Debug(ctx, "Requesting new token using the refresh token grant")
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("refresh_token", c.refreshToken.Raw)
	return c.sendTokenForm(ctx, form)
}

func (c *Connection) sendTokenForm(ctx context.Context, form url.Values) error {
	// Measure the time that it takes to send the request and receive the response:
	before := time.Now()
	code, err := c.sendTokenFormTimed(ctx, form)
	after := time.Now()
	elapsed := after.Sub(before)

	// Update the metrics:
	if c.tokenCountMetric != nil || c.tokenDurationMetric != nil {
		labels := map[string]string{
			metricsCodeLabel: strconv.Itoa(code),
		}
		if c.tokenCountMetric != nil {
			c.tokenCountMetric.With(labels).Inc()
		}
		if c.tokenDurationMetric != nil {
			c.tokenDurationMetric.With(labels).Observe(elapsed.Seconds())
		}
	}

	// Return the original error:
	return err
}

func (c *Connection) sendTokenFormTimed(ctx context.Context, form url.Values) (code int, err error) {
	// Create the HTTP request:
	body := []byte(form.Encode())
	request, err := http.NewRequest(http.MethodPost, c.tokenURL.String(), bytes.NewReader(body))
	request.Close = true
	header := request.Header
	if c.agent != "" {
		header.Set("User-Agent", c.agent)
	}
	header.Set("Content-Type", "application/x-www-form-urlencoded")
	header.Set("Accept", "application/json")
	if err != nil {
		err = fmt.Errorf("can't create request: %v", err)
		return
	}

	// Set the context:
	if ctx != nil {
		request = request.WithContext(ctx)
	}

	// Send the HTTP request:
	if c.logger.DebugEnabled() {
		var censoredBody bytes.Buffer
		// Unlike real url.Values.Encode(), this doesn't sort keys.
		for name, values := range form {
			for _, value := range values {
				// Buffer.Write*() don't require error checking but golangci-lint v1.10.2
				// on Jenkins flags them (maybe https://github.com/securego/gosec/issues/267).
				if censoredBody.Len() > 0 {
					censoredBody.WriteByte('&') // #nosec G104
				}
				censoredBody.WriteString(url.QueryEscape(name) + "=") // #nosec G104

				if isRedactField(name) {
					censoredBody.WriteString(redactionStr) // #nosec G104
				} else {
					censoredBody.WriteString(url.QueryEscape(value)) // #nosec G104
				}
			}
		}
		c.dumpRequest(ctx, request, censoredBody.Bytes())
	}
	response, err := c.client.Do(request)
	if err != nil {
		err = fmt.Errorf("can't send request: %v", err)
		return
	}
	defer response.Body.Close()

	// Read the response body:
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("can't read response: %v", err)
		return
	}
	if c.logger.DebugEnabled() {
		c.dumpResponse(ctx, response, body)
	}

	// Check the response status and content type:
	code = response.StatusCode
	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("token response status is: %s", response.Status)
		return
	}
	header = response.Header
	content := header.Get("Content-Type")
	if content != "application/json" {
		err = fmt.Errorf("expected 'application/json' but got '%s'", content)
		return
	}

	// Parse the response body:
	var msg internal.TokenResponse
	err = json.Unmarshal(body, &msg)
	if err != nil {
		err = fmt.Errorf("can't parse JSON response: %v", err)
		return
	}
	if msg.Error != nil {
		if msg.ErrorDescription != nil {
			err = fmt.Errorf("%s: %s", *msg.Error, *msg.ErrorDescription)
			return
		}
		err = fmt.Errorf("%s", *msg.Error)
		return
	}
	if msg.TokenType != nil && *msg.TokenType != "bearer" {
		err = fmt.Errorf("expected 'bearer' token type but got '%s", *msg.TokenType)
		return
	}
	if msg.AccessToken == nil {
		err = fmt.Errorf("no access token was received")
		return
	}
	accessToken, _, err := c.tokenParser.ParseUnverified(*msg.AccessToken, jwt.MapClaims{})
	if err != nil {
		return
	}
	if msg.RefreshToken == nil {
		err = fmt.Errorf("no refresh token was received")
		return
	}
	refreshToken, _, err := c.tokenParser.ParseUnverified(*msg.RefreshToken, jwt.MapClaims{})
	if err != nil {
		return
	}

	// Save the new tokens:
	c.accessToken = accessToken
	c.refreshToken = refreshToken

	return
}

// debugExpiry sends to the log information about the expiration of the given token.
func (c *Connection) debugExpiry(ctx context.Context, typ string, token *jwt.Token, expires bool,
	left time.Duration) {
	if token != nil {
		if expires {
			if left < 0 {
				c.logger.Debug(ctx, "%s token expired %s ago", typ, -left)
			} else if left > 0 {
				c.logger.Debug(ctx, "%s token expires in %s", typ, left)
			} else {
				c.logger.Debug(ctx, "%s token expired just now", typ)
			}
		}
	} else {
		c.logger.Debug(ctx, "%s token isn't available", typ)
	}
}

// tokenExpiry determines if the given token expires, and the time that remains till it expires.
func tokenExpiry(token *jwt.Token, now time.Time) (expires bool,
	left time.Duration, err error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = fmt.Errorf("expected map claims bug got %T", claims)
		return
	}
	claim, ok := claims["exp"]
	if !ok {
		err = fmt.Errorf("token doesn't contain the 'exp' claim")
		return
	}
	exp, ok := claim.(float64)
	if !ok {
		err = fmt.Errorf("expected floating point 'exp' but got %T", claim)
		return
	}
	if exp == 0 {
		expires = false
		left = 0
	} else {
		expires = true
		left = time.Unix(int64(exp), 0).Sub(now)
	}
	return
}

// tokenIssuer extracts the URL of the issuer of the token from the `iss` claim. Returns nil if
// there is no such claim.
func tokenIssuer(token *jwt.Token) (issuer *url.URL, err error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = fmt.Errorf("expected map claims bug got %T", claims)
		return
	}
	claim, ok := claims["iss"]
	if !ok {
		return
	}
	value, ok := claim.(string)
	if !ok {
		err = fmt.Errorf("expected string 'iss' but got %T", claim)
		return
	}
	issuer, err = url.Parse(value)
	return
}
