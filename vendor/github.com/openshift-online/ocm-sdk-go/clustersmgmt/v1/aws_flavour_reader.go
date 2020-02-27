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

// awsFlavourData is the data structure used internally to marshal and unmarshal
// objects of type 'AWS_flavour'.
type awsFlavourData struct {
	InfraVolume  *awsVolumeData "json:\"infra_volume,omitempty\""
	MasterVolume *awsVolumeData "json:\"master_volume,omitempty\""
	WorkerVolume *awsVolumeData "json:\"worker_volume,omitempty\""
}

// MarshalAWSFlavour writes a value of the 'AWS_flavour' to the given target,
// which can be a writer or a JSON encoder.
func MarshalAWSFlavour(object *AWSFlavour, target interface{}) error {
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

// wrap is the method used internally to convert a value of the 'AWS_flavour'
// value to a JSON document.
func (o *AWSFlavour) wrap() (data *awsFlavourData, err error) {
	if o == nil {
		return
	}
	data = new(awsFlavourData)
	data.InfraVolume, err = o.infraVolume.wrap()
	if err != nil {
		return
	}
	data.MasterVolume, err = o.masterVolume.wrap()
	if err != nil {
		return
	}
	data.WorkerVolume, err = o.workerVolume.wrap()
	if err != nil {
		return
	}
	return
}

// UnmarshalAWSFlavour reads a value of the 'AWS_flavour' type from the given
// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
func UnmarshalAWSFlavour(source interface{}) (object *AWSFlavour, err error) {
	decoder, err := helpers.NewDecoder(source)
	if err != nil {
		return
	}
	data := new(awsFlavourData)
	err = decoder.Decode(data)
	if err != nil {
		return
	}
	object, err = data.unwrap()
	return
}

// unwrap is the function used internally to convert the JSON unmarshalled data to a
// value of the 'AWS_flavour' type.
func (d *awsFlavourData) unwrap() (object *AWSFlavour, err error) {
	if d == nil {
		return
	}
	object = new(AWSFlavour)
	object.infraVolume, err = d.InfraVolume.unwrap()
	if err != nil {
		return
	}
	object.masterVolume, err = d.MasterVolume.unwrap()
	if err != nil {
		return
	}
	object.workerVolume, err = d.WorkerVolume.unwrap()
	if err != nil {
		return
	}
	return
}
