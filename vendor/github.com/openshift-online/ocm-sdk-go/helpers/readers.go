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

package helpers // github.com/openshift-online/ocm-sdk-go/helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"
)

// NewEncoder creates a new JSON encoder from the given target. The target can be a
// a writer or a JSON encoder.
func NewEncoder(target interface{}) (encoder *json.Encoder, err error) {
	switch output := target.(type) {
	case io.Writer:
		encoder = json.NewEncoder(output)
	case *json.Encoder:
		encoder = output
	default:
		err = fmt.Errorf(
			"expected writer or JSON decoder, but got %T",
			output,
		)
	}
	return
}

// NewDecoder creates a new JSON decoder from the given source. The source can be a
// slice of bytes, a string, a reader or a JSON decoder.
func NewDecoder(source interface{}) (decoder *json.Decoder, err error) {
	switch input := source.(type) {
	case []byte:
		decoder = json.NewDecoder(bytes.NewBuffer(input))
	case string:
		decoder = json.NewDecoder(bytes.NewBufferString(input))
	case io.Reader:
		decoder = json.NewDecoder(input)
	case *json.Decoder:
		decoder = input
	default:
		err = fmt.Errorf(
			"expected bytes, string, reader or JSON decoder, but got %T",
			input,
		)
	}
	return
}

// ParseInteger reads a string and parses it to integer,
// if an error occurred it returns a non-nil error.
func ParseInteger(query url.Values, parameterName string) (*int, error) {
	values := query[parameterName]
	count := len(values)
	if count == 0 {
		return nil, nil
	}
	if count > 1 {
		err := fmt.Errorf(
			"expected at most one value for parameter '%s' but got %d",
			parameterName, count,
		)
		return nil, err
	}
	value := values[0]
	parsedInt64, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, fmt.Errorf(
			"value '%s' isn't valid for the '%s' parameter because it isn't an integer: %v",
			value, parameterName, err,
		)
	}
	parsedInt := int(parsedInt64)
	return &parsedInt, nil
}

// ParseFloat reads a string and parses it to float,
// if an error occurred it returns a non-nil error.
func ParseFloat(query url.Values, parameterName string) (*float64, error) {
	values := query[parameterName]
	count := len(values)
	if count == 0 {
		return nil, nil
	}
	if count > 1 {
		err := fmt.Errorf(
			"expected at most one value for parameter '%s' but got %d",
			parameterName, count,
		)
		return nil, err
	}
	value := values[0]
	parsedFloat, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, fmt.Errorf(
			"value '%s' isn't valid for the '%s' parameter because it isn't a float: %v",
			value, parameterName, err,
		)
	}
	return &parsedFloat, nil
}

// ParseString returns a pointer to the string and nil error.
func ParseString(query url.Values, parameterName string) (*string, error) {
	values := query[parameterName]
	count := len(values)
	if count == 0 {
		return nil, nil
	}
	if count > 1 {
		err := fmt.Errorf(
			"expected at most one value for parameter '%s' but got %d",
			parameterName, count,
		)
		return nil, err
	}
	return &values[0], nil
}

// ParseBoolean reads a string and parses it to boolean,
// if an error occurred it returns a non-nil error.
func ParseBoolean(query url.Values, parameterName string) (*bool, error) {
	values := query[parameterName]
	count := len(values)
	if count == 0 {
		return nil, nil
	}
	if count > 1 {
		err := fmt.Errorf(
			"expected at most one value for parameter '%s' but got %d",
			parameterName, count,
		)
		return nil, err
	}
	value := values[0]
	parsedBool, err := strconv.ParseBool(value)
	if err != nil {
		return nil, fmt.Errorf(
			"value '%s' isn't valid for the '%s' parameter because it isn't a boolean: %v",
			value, parameterName, err,
		)
	}
	return &parsedBool, nil
}

// ParseDate reads a string and parses it to a time.Time,
// if an error occurred it returns a non-nil error.
func ParseDate(query url.Values, parameterName string) (*time.Time, error) {
	values := query[parameterName]
	count := len(values)
	if count == 0 {
		return nil, nil
	}
	if count > 1 {
		err := fmt.Errorf(
			"expected at most one value for parameter '%s' but got %d",
			parameterName, count,
		)
		return nil, err
	}
	value := values[0]
	parsedTime, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf(
			"value '%s' isn't valid for the '%s' parameter because it isn't a date: %v",
			value, parameterName, err,
		)
	}
	return &parsedTime, nil
}
