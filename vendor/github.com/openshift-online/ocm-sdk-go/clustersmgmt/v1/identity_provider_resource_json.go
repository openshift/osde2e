/*
Copyright (c) 2020 Red Hat, Inc.

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
	"io"
	"net/http"
)

func readIdentityProviderDeleteRequest(request *IdentityProviderDeleteServerRequest, r *http.Request) error {
	return nil
}
func writeIdentityProviderDeleteRequest(request *IdentityProviderDeleteRequest, writer io.Writer) error {
	return nil
}
func readIdentityProviderDeleteResponse(response *IdentityProviderDeleteResponse, reader io.Reader) error {
	return nil
}
func writeIdentityProviderDeleteResponse(response *IdentityProviderDeleteServerResponse, w http.ResponseWriter) error {
	return nil
}
func readIdentityProviderGetRequest(request *IdentityProviderGetServerRequest, r *http.Request) error {
	return nil
}
func writeIdentityProviderGetRequest(request *IdentityProviderGetRequest, writer io.Writer) error {
	return nil
}
func readIdentityProviderGetResponse(response *IdentityProviderGetResponse, reader io.Reader) error {
	var err error
	response.body, err = UnmarshalIdentityProvider(reader)
	return err
}
func writeIdentityProviderGetResponse(response *IdentityProviderGetServerResponse, w http.ResponseWriter) error {
	return MarshalIdentityProvider(response.body, w)
}
