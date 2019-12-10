/*
Copyright (c) 2019 Red Hat, Inc.

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

// IMPORTANT: This file has been generated automatically, refrain from modifying it manually as all
// your changes will be lost when the file is generated again.

package v1 // github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1

import (
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// gitlabIdentityProviderData is the data structure used internally to marshal and unmarshal
// objects of type 'gitlab_identity_provider'.
type gitlabIdentityProviderData struct {
	CA           *string "json:\"ca,omitempty\""
	URL          *string "json:\"url,omitempty\""
	ClientID     *string "json:\"client_id,omitempty\""
	ClientSecret *string "json:\"client_secret,omitempty\""
}

// MarshalGitlabIdentityProvider writes a value of the 'gitlab_identity_provider' to the given target,
// which can be a writer or a JSON encoder.
func MarshalGitlabIdentityProvider(object *GitlabIdentityProvider, target interface{}) error {
	encoder, err := helpers.NewEncoder(target)
	if err != nil {
		return err
	}
	data, err := object.wrap()
	if err != nil {
		return err
	}
	return encoder.Encode(data)
}

// wrap is the method used internally to convert a value of the 'gitlab_identity_provider'
// value to a JSON document.
func (o *GitlabIdentityProvider) wrap() (data *gitlabIdentityProviderData, err error) {
	if o == nil {
		return
	}
	data = new(gitlabIdentityProviderData)
	data.CA = o.ca
	data.URL = o.url
	data.ClientID = o.clientID
	data.ClientSecret = o.clientSecret
	return
}

// UnmarshalGitlabIdentityProvider reads a value of the 'gitlab_identity_provider' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalGitlabIdentityProvider(source interface{}) (object *GitlabIdentityProvider, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(gitlabIdentityProviderData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'gitlab_identity_provider' type.
func (d *gitlabIdentityProviderData) unwrap() (object *GitlabIdentityProvider, err error) {
	if d == nil {
		return
	}
	object = new(GitlabIdentityProvider)
	object.ca = d.CA
	object.url = d.URL
	object.clientID = d.ClientID
	object.clientSecret = d.ClientSecret
	return
}
