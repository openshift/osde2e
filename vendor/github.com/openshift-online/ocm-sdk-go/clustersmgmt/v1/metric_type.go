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

// Metric represents the values of the 'metric' type.
//
// Metric included in a dashboard.
type Metric struct {
	name   *string
	vector []*Sample
}

// Empty returns true if the object is empty, i.e. no attribute has a value.
func (o *Metric) Empty() bool {
	return o == nil || (o.name == nil &&
		len(o.vector) == 0 &&
		true)
}

// Name returns the value of the 'name' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Name of the metric.
func (o *Metric) Name() string {
	if o != nil && o.name != nil {
		return *o.name
	}
	return ""
}

// GetName returns the value of the 'name' attribute and
// a flag indicating if the attribute has a value.
//
// Name of the metric.
func (o *Metric) GetName() (value string, ok bool) {
	ok = o != nil && o.name != nil
	if ok {
		value = *o.name
	}
	return
}

// Vector returns the value of the 'vector' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Samples of the metric.
func (o *Metric) Vector() []*Sample {
	if o == nil {
		return nil
	}
	return o.vector
}

// GetVector returns the value of the 'vector' attribute and
// a flag indicating if the attribute has a value.
//
// Samples of the metric.
func (o *Metric) GetVector() (value []*Sample, ok bool) {
	ok = o != nil && o.vector != nil
	if ok {
		value = o.vector
	}
	return
}

// MetricListKind is the name of the type used to represent list of objects of
// type 'metric'.
const MetricListKind = "MetricList"

// MetricListLinkKind is the name of the type used to represent links to list
// of objects of type 'metric'.
const MetricListLinkKind = "MetricListLink"

// MetricNilKind is the name of the type used to nil lists of objects of
// type 'metric'.
const MetricListNilKind = "MetricListNil"

// MetricList is a list of values of the 'metric' type.
type MetricList struct {
	href  *string
	link  bool
	items []*Metric
}

// Len returns the length of the list.
func (l *MetricList) Len() int {
	if l == nil {
		return 0
	}
	return len(l.items)
}

// Empty returns true if the list is empty.
func (l *MetricList) Empty() bool {
	return l == nil || len(l.items) == 0
}

// Get returns the item of the list with the given index. If there is no item with
// that index it returns nil.
func (l *MetricList) Get(i int) *Metric {
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
func (l *MetricList) Slice() []*Metric {
	var slice []*Metric
	if l == nil {
		slice = make([]*Metric, 0)
	} else {
		slice = make([]*Metric, len(l.items))
		copy(slice, l.items)
	}
	return slice
}

// Each runs the given function for each item of the list, in order. If the function
// returns false the iteration stops, otherwise it continues till all the elements
// of the list have been processed.
func (l *MetricList) Each(f func(item *Metric) bool) {
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
func (l *MetricList) Range(f func(index int, item *Metric) bool) {
	if l == nil {
		return
	}
	for index, item := range l.items {
		if !f(index, item) {
			break
		}
	}
}
