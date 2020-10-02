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

package v1 // github.com/openshift-online/ocm-sdk-go/authorizations/v1

// TermsReviewRequestBuilder contains the data and logic needed to build 'terms_review_request' objects.
//
// Representation of Red Hat's Terms and Conditions for using OpenShift Dedicated and Amazon Red Hat OpenShift [Terms]
// review requests.
type TermsReviewRequestBuilder struct {
	accountUsername *string
}

// NewTermsReviewRequest creates a new builder of 'terms_review_request' objects.
func NewTermsReviewRequest() *TermsReviewRequestBuilder {
	return new(TermsReviewRequestBuilder)
}

// AccountUsername sets the value of the 'account_username' attribute to the given value.
//
//
func (b *TermsReviewRequestBuilder) AccountUsername(value string) *TermsReviewRequestBuilder {
	b.accountUsername = &value
	return b
}

// Copy copies the attributes of the given object into this builder, discarding any previous values.
func (b *TermsReviewRequestBuilder) Copy(object *TermsReviewRequest) *TermsReviewRequestBuilder {
	if object == nil {
		return b
	}
	b.accountUsername = object.accountUsername
	return b
}

// Build creates a 'terms_review_request' object using the configuration stored in the builder.
func (b *TermsReviewRequestBuilder) Build() (object *TermsReviewRequest, err error) {
	object = new(TermsReviewRequest)
	object.accountUsername = b.accountUsername
	return
}
