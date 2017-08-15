// Copyright (c) 2017 Nutanix Inc. All rights reserved.
//
// Description:
//
// In order to enable the event consumer (plugin) to consume the parameters
// defined in configuration file we need create an event consumer (plugin)
// schema file. Here we need to match every relevant field from the configuration
// file to a relevant data structure in the (schema) file.
//
// The schema file be used by the event consumer (plugin) to pass the cluster
// credentials to the listener, connect to the third party product and consume
// the event based upon the defined optional parameters.
//
// The event consumer configuration schema file comprises of:
//   1) Nutanix cluster connection details (Cluster External IP, Prism username
//      and Prism password)
//   2) Third party product connection details (IP , username and password)
//   3) Relevant optional configuration parameters that will be consumed by the event consumer.
//
// NOTE :
//   1) Developers should exercise their discretion in using their choice of
//      encryption mechanism for encoding and decoding credentials. For illustration
//      purposes the credentials this configuration file are base64 encoded.
//   2) Developers should exercise their discretion in determining the configuration
//      parameters and source of the configuration parameters. For illustration purposes,
//      this configuration file for PaloAlto Firewall appliance contain the "Security Policy Rule" name.
//
package config

type PAFWConfig struct {
  PAFWInstanceConfig PAFWInstanceConfig `json:"pafw_instance_config"`
  NutanixClusterConfig NutanixClusterConfig `json:"nutanix_cluster_config"`
}

type PAFWInstanceConfig struct {
  IP string `json:"ip"`
  Port string `json:"port"`
  Username string `json:"username"`
  Password string `json:"password"`
  AddressGroup string `json:"dynamic_address_group"`
  SecurityPolicyRule string `json:"security_policy_rule"`
  DeviceGroup string `json:"device_group"`
  Category string `json:"category"`
}

type NutanixClusterConfig struct {
  IP string `json:"ip"`
  Port string `json:"port"`
  Username string `json:"username"`
  Password string `json:"password"`
}
