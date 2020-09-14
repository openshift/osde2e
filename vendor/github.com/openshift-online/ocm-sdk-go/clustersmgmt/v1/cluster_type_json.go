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
	"sort"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/openshift-online/ocm-sdk-go/helpers"
)

// MarshalCluster writes a value of the 'cluster' type to the given writer.
func MarshalCluster(object *Cluster, writer io.Writer) error {
	stream := helpers.NewStream(writer)
	writeCluster(object, stream)
	stream.Flush()
	return stream.Error
}

// writeCluster writes a value of the 'cluster' type to the given stream.
func writeCluster(object *Cluster, stream *jsoniter.Stream) {
	count := 0
	stream.WriteObjectStart()
	if count > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField("kind")
	if object.link {
		stream.WriteString(ClusterLinkKind)
	} else {
		stream.WriteString(ClusterKind)
	}
	count++
	if object.id != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("id")
		stream.WriteString(*object.id)
		count++
	}
	if object.href != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("href")
		stream.WriteString(*object.href)
		count++
	}
	if object.api != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("api")
		writeClusterAPI(object.api, stream)
		count++
	}
	if object.aws != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("aws")
		writeAWS(object.aws, stream)
		count++
	}
	if object.awsInfrastructureAccessRoleGrants != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("aws_infrastructure_access_role_grants")
		stream.WriteObjectStart()
		stream.WriteObjectField("items")
		writeAWSInfrastructureAccessRoleGrantList(object.awsInfrastructureAccessRoleGrants.items, stream)
		stream.WriteObjectEnd()
		count++
	}
	if object.byoc != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("byoc")
		stream.WriteBool(*object.byoc)
		count++
	}
	if object.ccs != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("ccs")
		writeCCS(object.ccs, stream)
		count++
	}
	if object.dns != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("dns")
		writeDNS(object.dns, stream)
		count++
	}
	if object.dnsReady != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("dns_ready")
		stream.WriteBool(*object.dnsReady)
		count++
	}
	if object.addons != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("addons")
		stream.WriteObjectStart()
		stream.WriteObjectField("items")
		writeAddOnInstallationList(object.addons.items, stream)
		stream.WriteObjectEnd()
		count++
	}
	if object.cloudProvider != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("cloud_provider")
		writeCloudProvider(object.cloudProvider, stream)
		count++
	}
	if object.clusterAdminEnabled != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("cluster_admin_enabled")
		stream.WriteBool(*object.clusterAdminEnabled)
		count++
	}
	if object.console != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("console")
		writeClusterConsole(object.console, stream)
		count++
	}
	if object.creationTimestamp != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("creation_timestamp")
		stream.WriteString((*object.creationTimestamp).Format(time.RFC3339))
		count++
	}
	if object.displayName != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("display_name")
		stream.WriteString(*object.displayName)
		count++
	}
	if object.expirationTimestamp != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("expiration_timestamp")
		stream.WriteString((*object.expirationTimestamp).Format(time.RFC3339))
		count++
	}
	if object.externalID != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("external_id")
		stream.WriteString(*object.externalID)
		count++
	}
	if object.externalConfiguration != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("external_configuration")
		writeExternalConfiguration(object.externalConfiguration, stream)
		count++
	}
	if object.flavour != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("flavour")
		writeFlavour(object.flavour, stream)
		count++
	}
	if object.groups != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("groups")
		stream.WriteObjectStart()
		stream.WriteObjectField("items")
		writeGroupList(object.groups.items, stream)
		stream.WriteObjectEnd()
		count++
	}
	if object.healthState != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("health_state")
		stream.WriteString(string(*object.healthState))
		count++
	}
	if object.identityProviders != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("identity_providers")
		stream.WriteObjectStart()
		stream.WriteObjectField("items")
		writeIdentityProviderList(object.identityProviders.items, stream)
		stream.WriteObjectEnd()
		count++
	}
	if object.ingresses != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("ingresses")
		stream.WriteObjectStart()
		stream.WriteObjectField("items")
		writeIngressList(object.ingresses.items, stream)
		stream.WriteObjectEnd()
		count++
	}
	if object.loadBalancerQuota != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("load_balancer_quota")
		stream.WriteInt(*object.loadBalancerQuota)
		count++
	}
	if object.managed != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("managed")
		stream.WriteBool(*object.managed)
		count++
	}
	if object.metrics != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("metrics")
		writeClusterMetrics(object.metrics, stream)
		count++
	}
	if object.multiAZ != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("multi_az")
		stream.WriteBool(*object.multiAZ)
		count++
	}
	if object.name != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("name")
		stream.WriteString(*object.name)
		count++
	}
	if object.network != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("network")
		writeNetwork(object.network, stream)
		count++
	}
	if object.nodes != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("nodes")
		writeClusterNodes(object.nodes, stream)
		count++
	}
	if object.openshiftVersion != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("openshift_version")
		stream.WriteString(*object.openshiftVersion)
		count++
	}
	if object.product != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("product")
		writeProduct(object.product, stream)
		count++
	}
	if object.properties != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("properties")
		stream.WriteObjectStart()
		keys := make([]string, len(object.properties))
		i := 0
		for key := range object.properties {
			keys[i] = key
			i++
		}
		sort.Strings(keys)
		for i, key := range keys {
			if i > 0 {
				stream.WriteMore()
			}
			item := object.properties[key]
			stream.WriteObjectField(key)
			stream.WriteString(item)
		}
		stream.WriteObjectEnd()
		count++
	}
	if object.provisionShard != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("provision_shard")
		writeProvisionShard(object.provisionShard, stream)
		count++
	}
	if object.region != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("region")
		writeCloudRegion(object.region, stream)
		count++
	}
	if object.state != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("state")
		stream.WriteString(string(*object.state))
		count++
	}
	if object.storageQuota != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("storage_quota")
		writeValue(object.storageQuota, stream)
		count++
	}
	if object.subscription != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("subscription")
		writeSubscription(object.subscription, stream)
		count++
	}
	if object.version != nil {
		if count > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField("version")
		writeVersion(object.version, stream)
		count++
	}
	stream.WriteObjectEnd()
}

// UnmarshalCluster reads a value of the 'cluster' type from the given
// source, which can be an slice of bytes, a string or a reader.
func UnmarshalCluster(source interface{}) (object *Cluster, err error) {
	iterator, err := helpers.NewIterator(source)
	if err != nil {
		return
	}
	object = readCluster(iterator)
	err = iterator.Error
	return
}

// readCluster reads a value of the 'cluster' type from the given iterator.
func readCluster(iterator *jsoniter.Iterator) *Cluster {
	object := &Cluster{}
	for {
		field := iterator.ReadObject()
		if field == "" {
			break
		}
		switch field {
		case "kind":
			value := iterator.ReadString()
			object.link = value == ClusterLinkKind
		case "id":
			value := iterator.ReadString()
			object.id = &value
		case "href":
			value := iterator.ReadString()
			object.href = &value
		case "api":
			value := readClusterAPI(iterator)
			object.api = value
		case "aws":
			value := readAWS(iterator)
			object.aws = value
		case "aws_infrastructure_access_role_grants":
			value := &AWSInfrastructureAccessRoleGrantList{}
			for {
				field := iterator.ReadObject()
				if field == "" {
					break
				}
				switch field {
				case "kind":
					text := iterator.ReadString()
					value.link = text == AWSInfrastructureAccessRoleGrantListLinkKind
				case "href":
					text := iterator.ReadString()
					value.href = &text
				case "items":
					value.items = readAWSInfrastructureAccessRoleGrantList(iterator)
				default:
					iterator.ReadAny()
				}
			}
			object.awsInfrastructureAccessRoleGrants = value
		case "byoc":
			value := iterator.ReadBool()
			object.byoc = &value
		case "ccs":
			value := readCCS(iterator)
			object.ccs = value
		case "dns":
			value := readDNS(iterator)
			object.dns = value
		case "dns_ready":
			value := iterator.ReadBool()
			object.dnsReady = &value
		case "addons":
			value := &AddOnInstallationList{}
			for {
				field := iterator.ReadObject()
				if field == "" {
					break
				}
				switch field {
				case "kind":
					text := iterator.ReadString()
					value.link = text == AddOnInstallationListLinkKind
				case "href":
					text := iterator.ReadString()
					value.href = &text
				case "items":
					value.items = readAddOnInstallationList(iterator)
				default:
					iterator.ReadAny()
				}
			}
			object.addons = value
		case "cloud_provider":
			value := readCloudProvider(iterator)
			object.cloudProvider = value
		case "cluster_admin_enabled":
			value := iterator.ReadBool()
			object.clusterAdminEnabled = &value
		case "console":
			value := readClusterConsole(iterator)
			object.console = value
		case "creation_timestamp":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.creationTimestamp = &value
		case "display_name":
			value := iterator.ReadString()
			object.displayName = &value
		case "expiration_timestamp":
			text := iterator.ReadString()
			value, err := time.Parse(time.RFC3339, text)
			if err != nil {
				iterator.ReportError("", err.Error())
			}
			object.expirationTimestamp = &value
		case "external_id":
			value := iterator.ReadString()
			object.externalID = &value
		case "external_configuration":
			value := readExternalConfiguration(iterator)
			object.externalConfiguration = value
		case "flavour":
			value := readFlavour(iterator)
			object.flavour = value
		case "groups":
			value := &GroupList{}
			for {
				field := iterator.ReadObject()
				if field == "" {
					break
				}
				switch field {
				case "kind":
					text := iterator.ReadString()
					value.link = text == GroupListLinkKind
				case "href":
					text := iterator.ReadString()
					value.href = &text
				case "items":
					value.items = readGroupList(iterator)
				default:
					iterator.ReadAny()
				}
			}
			object.groups = value
		case "health_state":
			text := iterator.ReadString()
			value := ClusterHealthState(text)
			object.healthState = &value
		case "identity_providers":
			value := &IdentityProviderList{}
			for {
				field := iterator.ReadObject()
				if field == "" {
					break
				}
				switch field {
				case "kind":
					text := iterator.ReadString()
					value.link = text == IdentityProviderListLinkKind
				case "href":
					text := iterator.ReadString()
					value.href = &text
				case "items":
					value.items = readIdentityProviderList(iterator)
				default:
					iterator.ReadAny()
				}
			}
			object.identityProviders = value
		case "ingresses":
			value := &IngressList{}
			for {
				field := iterator.ReadObject()
				if field == "" {
					break
				}
				switch field {
				case "kind":
					text := iterator.ReadString()
					value.link = text == IngressListLinkKind
				case "href":
					text := iterator.ReadString()
					value.href = &text
				case "items":
					value.items = readIngressList(iterator)
				default:
					iterator.ReadAny()
				}
			}
			object.ingresses = value
		case "load_balancer_quota":
			value := iterator.ReadInt()
			object.loadBalancerQuota = &value
		case "managed":
			value := iterator.ReadBool()
			object.managed = &value
		case "metrics":
			value := readClusterMetrics(iterator)
			object.metrics = value
		case "multi_az":
			value := iterator.ReadBool()
			object.multiAZ = &value
		case "name":
			value := iterator.ReadString()
			object.name = &value
		case "network":
			value := readNetwork(iterator)
			object.network = value
		case "nodes":
			value := readClusterNodes(iterator)
			object.nodes = value
		case "openshift_version":
			value := iterator.ReadString()
			object.openshiftVersion = &value
		case "product":
			value := readProduct(iterator)
			object.product = value
		case "properties":
			value := map[string]string{}
			for {
				key := iterator.ReadObject()
				if key == "" {
					break
				}
				item := iterator.ReadString()
				value[key] = item
			}
			object.properties = value
		case "provision_shard":
			value := readProvisionShard(iterator)
			object.provisionShard = value
		case "region":
			value := readCloudRegion(iterator)
			object.region = value
		case "state":
			text := iterator.ReadString()
			value := ClusterState(text)
			object.state = &value
		case "storage_quota":
			value := readValue(iterator)
			object.storageQuota = value
		case "subscription":
			value := readSubscription(iterator)
			object.subscription = value
		case "version":
			value := readVersion(iterator)
			object.version = value
		default:
			iterator.ReadAny()
		}
	}
	return object
}
