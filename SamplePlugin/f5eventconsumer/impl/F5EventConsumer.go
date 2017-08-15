// Copyright (c) 2017 Nutanix Inc. All rights reserved.

// Implementation of the event consumer interface for F5.
//
// Description:
//   1) On VM.ON event, consumer adds web server to a predefined web server
//      tier load balancing pool.
//   2) On VM.OFF event, consumer removes the web server from the predefined
//      webserver tier load balancing pool.
//
package consumer

import (
  "encoding/json"
  "encoding/base64"
  "fmt"
  "aplos/partners/f5eventconsumer/config"
  "github.com/golang/glog"
  "io/ioutil"
  "aplos/partners/WebhooksListener/lib"
  "aplos/partners/WebhooksListener/schemas"
  "regexp"
)

type F5EventConsumer struct {
  // Type that implements the EventConsumer interface.
}

const (
  // F5 Config Directory Path
  F5ConfigDir = "/opt/f5/config/"
)

// This method will act as a callback method that will configure the networking
// appliance based on the networking events received through the listener.
//
// Args:
//    event : Event object containing the event data sent by the listener.
// Returns:
//    error : Error, if any.
func (f5EventConsumer F5EventConsumer) OnEvent(event schema.Event) (error) {
  var err error

  glog.Info("Received event of type " + event.Event_Type)
  // Load F5 BIG IP Event consumer configuration file.
  f5Config, err := LoadF5Config()
  if (err != nil) {
    glog.Error("Failed to load config. Cannot proceed.", err)
    return err
  }

  switch eventType := event.Event_Type; eventType {
    // If event type is VM.ON
    case lib.VM_ON: {
      err = onVmOn(event, f5Config)
    }
    // If event type is VM.OFF
    case lib.VM_OFF: {
      err = onVmOff(event, f5Config)
    }
  }
  if (err != nil) {
    glog.Error("Failed to process event.", err)
  }

  return err
}

// This method will process the VM.ON event. It will make the
// intended configuration on the target virtual appliance.
//
// Args:
//    event : Event object containing the event data sent by the listener.
//    f5Config : F5 BIG IP Event consumer configuration data structure.
// Returns:
//    error : Error, if any.
func onVmOn(event schema.Event, f5Config config.F5Config) (error) {
  var err error
  vmIPAddress := event.Data.Metadata.Status.Resources.NICList[0].IPEndPointList[0].IPAddress
  vmCategoryPool := event.Data.Metadata.SubMetadata.Categories.NetworkFunctionProvider
  glog.Infof("Processing '%s' event.", event.Event_Type)
  // Prepare F5 BIG IP REST API for creating or updating load balancing pool
  baseURL := fmt.Sprintf("https://%s:%s/mgmt/tm/ltm/pool",
    f5Config.F5InstanceConfig.IP,
    f5Config.F5InstanceConfig.Port)

  // Check if Pool already exist.
  glog.Infof("Checking if Pool '%s' already exists.", vmCategoryPool)
  requestURL := fmt.Sprintf("%s/%s", baseURL, vmCategoryPool)

  // Prepare Request to the F5 BIG IP virtual appliance
  request := lib.PrepareRequest(requestURL,
                                  f5Config.F5InstanceConfig.Username,
				  f5Config.F5InstanceConfig.Password, "GET")

  response, err := lib.DoRequest(request)
  respData, _ := ioutil.ReadAll(response.Body)
  pattern := fmt.Sprintf("\"name\":\"%s\"", vmCategoryPool)
  matched, _ := regexp.MatchString(pattern, string(respData))
  // Load Balancing pool does not exist. Create a pool.
  if (matched == false) {
    glog.Infof("Pool '%s' not exists. Creating now.", vmCategoryPool)
    // Prepare Request to the F5 BIG IP virtual appliance
    request = lib.PrepareRequest(baseURL,
                                  f5Config.F5InstanceConfig.Username,
				  f5Config.F5InstanceConfig.Password, "POST")
    request.RequestData = fmt.Sprintf("{\"name\": \"%s\"}", vmCategoryPool)
    glog.Info(requestURL)
    glog.Info(request.RequestData)
    response, err = lib.DoRequest(request)
    if (err != nil || (
        response.StatusCode != 200 && response.StatusCode != 409)) {
      glog.Error("Failed to create pool.", err)
      return err
    }
  } else {
    glog.Infof("Pool '%s' already exists.", vmCategoryPool)
  }

  // Add members to the pool.
  requestURL = fmt.Sprintf("%s/%s/members", baseURL, vmCategoryPool)
  // Prepare Request to the F5 BIG IP virtual appliance
  request = lib.PrepareRequest(requestURL,
                                  f5Config.F5InstanceConfig.Username,
				  f5Config.F5InstanceConfig.Password, "POST")
  request.RequestData = fmt.Sprintf("{\"name\": \"%s:%s\"}",
    vmIPAddress, f5Config.F5InstanceConfig.Serviceport)
  glog.Info(requestURL)
  glog.Info(request.RequestData)
  response, err = lib.DoRequest(request)
  if (err != nil || (
        response.StatusCode != 200 && response.StatusCode != 409)) {
    glog.Error("Failed to add member to pool.", err)
    return err
  }
  if (response.StatusCode == 409) {
    glog.Warning("Member already added to the pool.")
    return nil
  }
  resp, _ := ioutil.ReadAll(response.Body)
  glog.Info("Successfully added member to pool.", string(resp))
  return err
}

// This method will process the VM.OFF event.
// It will make the intended configuration on the target virtual appliance.
//
// Args:
//    event : Event object containing the event data sent by the listener.
//    f5Config : F5 BIG IP Event consumer configuration data structure.
// Returns:
//    error : Error, if any.
func onVmOff(event schema.Event, f5Config config.F5Config) (error) {
  var err error
  glog.Infof("Processing %v event.", event.Event_Type)
  // Prepare F5 BIG IP REST API for removing members from the pool
  vmIPAddress := event.Data.Metadata.Status.Resources.NICList[0].IPEndPointList[0].IPAddress
  vmCategoryPool := event.Data.Metadata.SubMetadata.Categories.NetworkFunctionProvider
  requestURL := fmt.Sprintf("https://%s:%s/mgmt/tm/ltm/pool/%s/members/%s:%s",
    f5Config.F5InstanceConfig.IP,
    f5Config.F5InstanceConfig.Port,
    vmCategoryPool, vmIPAddress, f5Config.F5InstanceConfig.Serviceport)

  // Prepare Request to the F5 BIG IP virtual appliance
  request := lib.PrepareRequest(requestURL,
                                  f5Config.F5InstanceConfig.Username,
				  f5Config.F5InstanceConfig.Password, "DELETE")
  glog.Info(requestURL)
  response, err := lib.DoRequest(request)
  if (err != nil || response.StatusCode != 200) {
    glog.Error("Failed to delete member from pool.", err)
    return err
  }

  resp, _ := ioutil.ReadAll(response.Body)
  glog.Info("Successfully deleted member from pool.", string(resp))
  return err
}

// This method will load the F5 specific config.
//
// Args:
//    None.
// Returns:
//    f5Config  : F5 specific config.
//    error : Error, if any.
func LoadF5Config() (config.F5Config, error) {
  var f5Config config.F5Config
  glog.Info("Loading config..")
  // F5 BIG IP Event consumer configuration file.
  f5ConfigPath := F5ConfigDir + "f5_config.json"
  // Read event consumer configuration file content
  f5ConfigFileContent, err := ioutil.ReadFile(f5ConfigPath)
  if (err != nil) {
    glog.Error("Error reading config.", err)
    return f5Config, err
  }
  // Unmarshal event consumer file content
  err = json.Unmarshal([]byte(f5ConfigFileContent), &f5Config)
  if (err != nil) {
    glog.Error("Failed to unmarshal config.",err)
  }
  // Decode F5 instance base64 encoded password.
  // Note : Developers can exercise their own encryption mechanism for credentials.
  decoded, err := base64.StdEncoding.DecodeString(f5Config.F5InstanceConfig.Password)
  if err != nil {
    glog.Error("Decode error:", err)
  } else {
    f5Config.F5InstanceConfig.Password = string(decoded)
  }
  // Decode Nutanix Prism Base64 encoded password.
  // Note : Developers can exercise their own encryption mechanism for credentials.
  decoded, err = base64.StdEncoding.DecodeString(f5Config.NutanixClusterConfig.Password)
  if err != nil {
    glog.Error("Decode error:", err)
  } else {
    f5Config.NutanixClusterConfig.Password = string(decoded)
  }
  if err != nil {
    glog.Error("Decode error:", err)
  }
  return f5Config, err
}
