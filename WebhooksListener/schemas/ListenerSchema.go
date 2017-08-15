// Copyright (c) 2017 Nutanix Inc. All rights reserved.
//
// Description:
//
// The listener configuration schema file comprises of data structure representing
// the different data received as part of event notification. In order to enable
// the listener to consume the event data that will be published by Webhooks,
// we need to define a schema file for the listener. Here we need to match every
// relevant field from the event data to a relevant data structure in the (schema) file.
// The schema file be used by the listener to unmarshall the event data.
//
// This schema file comprises of data structures representing every element received as
// part of web hooks event notification.
//
package schema

// Schema definition for the webhook event.
type Event struct {
  EntityReference Reference `json:"entity_reference"`
  Data Data `json:"data"`
  Version string `json:"version"`
  Event_Type string `json:"event_type"`
}

type Reference struct {
  KIND string `json:"kind"`
  UUID string `json:"uuid"`
}

type Data struct {
  Metadata EventMetadata `json:"metadata"`
}

type EventMetadata struct {
  Status Status `json:"status"`
  Spec EventSpec `json:"spec"`
  APIVersion string `json:"api_version"`
  SubMetadata EventSubMetadata `json:"metadata"`
}

type Status struct {
  State string `json:"state"`
  Name string `json:"name"`
  Resources EventResources `json:"resources"`
}

type EventSpec struct {
}

type EventResources struct {
  NICList []NIC `json:"nic_list"`
  HostReference HostReference `json:"host_reference"`
  HypervisorType string `json:"hypervisor_type"`
  NumVCPUsPerSocket int `json:"num_vcpus_per_socket"`
  NumSockets int `json:"num_sockets"`
  MemorySizeMB int `json:"memory_size_mb"`
  GPUList []string `json:"gpu_list"`
  PowerState string `json:"power_state"`
  DiskList []DiskList `json:"disk_list"`
}

type NIC struct {
  IPEndPointList []IPEndPointList `json:"ip_endpoint_list"`
  NetworkReference NetworkReference `json:"network_reference"`
  MACAddress string `json:"mac_address"`
}

type IPEndPointList struct {
  IPAddress string `json:"ip"`
}

type NetworkReference struct {
  Kind string `json:"kind"`
  UUID string `json:"uuid"`
}

type HostReference struct {
  Kind string `json:"kind"`
  UUID string `json:"uuid"`
}

type DiskList struct {
  DeviceProperties DeviceProperties `json:"device_properties"`
}

type DeviceProperties struct {
  DiskAddress DiskAddress `json:"disk_address"`
  DeviceType string `json:"device_type"`
}

type DiskAddress struct {
  DeviceIndex int `json:"device_index"`
  AdapterType string `json:"adapter_type"`
}

type EventSubMetadata struct {
  OwnerReference OwnerReference `json:"owner_reference"`
  Kind string `json:"kind"`
  EntityVersion int `json:"entity_version"`
  UUID string `json:"uuid"`
  Categories Categories `json:"categories"`
}

type OwnerReference struct {
  Kind string `json:"kind"`
  UUID string `json:"uuid"`
  Name string `json:"name"`
}

type Categories struct {
  NetworkFunctionProvider string `json:"network_function_provider"`
}
