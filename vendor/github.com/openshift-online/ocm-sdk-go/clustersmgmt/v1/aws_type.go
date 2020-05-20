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

// AWS represents the values of the 'AWS' type.
//
// _Amazon Web Services_ specific settings of a cluster.
type AWS struct {
	accessKeyID     *string
	accountID       *string
	secretAccessKey *string
}

// Empty returns true if the object is empty, i.e. no attribute has a value.
func (o *AWS) Empty() bool {
	return o == nil || (o.accessKeyID == nil &&
		o.accountID == nil &&
		o.secretAccessKey == nil &&
		true)
}

// AccessKeyID returns the value of the 'access_key_ID' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// AWS access key identifier.
func (o *AWS) AccessKeyID() string {
	if o != nil && o.accessKeyID != nil {
		return *o.accessKeyID
	}
	return ""
}

// GetAccessKeyID returns the value of the 'access_key_ID' attribute and
// a flag indicating if the attribute has a value.
//
// AWS access key identifier.
func (o *AWS) GetAccessKeyID() (value string, ok bool) {
	ok = o != nil && o.accessKeyID != nil
	if ok {
		value = *o.accessKeyID
	}
	return
}

// AccountID returns the value of the 'account_ID' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// AWS account identifier.
func (o *AWS) AccountID() string {
	if o != nil && o.accountID != nil {
		return *o.accountID
	}
	return ""
}

// GetAccountID returns the value of the 'account_ID' attribute and
// a flag indicating if the attribute has a value.
//
// AWS account identifier.
func (o *AWS) GetAccountID() (value string, ok bool) {
	ok = o != nil && o.accountID != nil
	if ok {
		value = *o.accountID
	}
	return
}

// SecretAccessKey returns the value of the 'secret_access_key' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// AWS secret access key.
func (o *AWS) SecretAccessKey() string {
	if o != nil && o.secretAccessKey != nil {
		return *o.secretAccessKey
	}
	return ""
}

// GetSecretAccessKey returns the value of the 'secret_access_key' attribute and
// a flag indicating if the attribute has a value.
//
// AWS secret access key.
func (o *AWS) GetSecretAccessKey() (value string, ok bool) {
	ok = o != nil && o.secretAccessKey != nil
	if ok {
		value = *o.secretAccessKey
	}
	return
}

// AWSListKind is the name of the type used to represent list of objects of
// type 'AWS'.
const AWSListKind = "AWSList"

// AWSListLinkKind is the name of the type used to represent links to list
// of objects of type 'AWS'.
const AWSListLinkKind = "AWSListLink"

// AWSNilKind is the name of the type used to nil lists of objects of
// type 'AWS'.
const AWSListNilKind = "AWSListNil"

// AWSList is a list of values of the 'AWS' type.
type AWSList struct {
	href  *string
	link  bool
	items []*AWS
}

// Len returns the length of the list.
func (l *AWSList) Len() int {
	if l == nil {
		return 0
	}
	return len(l.items)
}

// Empty returns true if the list is empty.
func (l *AWSList) Empty() bool {
	return l == nil || len(l.items) == 0
}

// Get returns the item of the list with the given index. If there is no item with
// that index it returns nil.
func (l *AWSList) Get(i int) *AWS {
	if l == nil || i < 0 || i >= len(l.items) {
		return nil
	}
	return l.items[i]
}

// Slice returns an slice containing the items of the list. The returned slice is a
// copy of the one used internally, so it can be modified without affecting the
// internal representation.
//
// If you don't need to modify the returned slice consider using the Each or Range
// functions, as they don't need to allocate a new slice.
func (l *AWSList) Slice() []*AWS {
	var slice []*AWS
	if l == nil {
		slice = make([]*AWS, 0)
	} else {
		slice = make([]*AWS, len(l.items))
		copy(slice, l.items)
	}
	return slice
}

// Each runs the given function for each item of the list, in order. If the function
// returns false the iteration stops, otherwise it continues till all the elements
// of the list have been processed.
func (l *AWSList) Each(f func(item *AWS) bool) {
	if l == nil {
		return
	}
	for _, item := range l.items {
		if !f(item) {
			break
		}
	}
}

// Range runs the given function for each index and item of the list, in order. If
// the function returns false the iteration stops, otherwise it continues till all
// the elements of the list have been processed.
func (l *AWSList) Range(f func(index int, item *AWS) bool) {
	if l == nil {
		return
	}
	for index, item := range l.items {
		if !f(index, item) {
			break
		}
	}
}
