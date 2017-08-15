// Copyright (c) 2017 Nutanix Inc. All rights reserved.

// Implementation of the event consumer interface for PAFW.
//
// Description:
//   1) On VM.ON event, perform below steps.
//      a) Login to firewall VM.
//      b) Create Tag & Dynamic Address Group objects.
//      c) Associate VM of event to Dynamic Address Group.
//      d) Apply Security Policy Rule to Address Group.
//   2) On VM.OFF event, consumer removes the VM address from firewall VM.
//

package pafweventconsumer

import (
  "aplos/partners/pafweventconsumer/config"
  "aplos/partners/WebhooksListener/lib"
  "aplos/partners/WebhooksListener/schemas"
  "bytes"
  "crypto/tls"
  "encoding/json"
  "encoding/base64"
  "errors"
  "fmt"
  "github.com/golang/glog"
  "io/ioutil"
  "net/http"
  "regexp"
  "strings"
)

type PAFWEventConsumer struct {
  // Type that implements the EventConsumer interface.
}

const (
  // PaloAlto Config Directory Path
  PAFWConfigDir = "/opt/pafw/config/"
)

// This method will act as a callback method that will configure the networking
// appliance based on the networking events received through the listener.
//
// Args:
//    event : Event object containing the event data sent by the listener.
// Returns:
//    error : Error, if any.
func (pafwEventConsumer PAFWEventConsumer) OnEvent(event schema.Event) (error) {
  var err error
  glog.Info("Received event of type " + event.Event_Type)
  // Load Palo Alto Firewall Event consumer configuration file
  pafwConfig, err := LoadPAFWConfig()
  if (err != nil) {
    glog.Error("Failed to load config. Cannot proceed.", err)
    return err
  }

  switch eventType := event.Event_Type; eventType {
  // If event type is VM.ON
  case lib.VM_ON: {
    err := onVmOn(event, pafwConfig)
    if (err != nil) {
      glog.Error("Failed to process event. Error: ", err)
    }
  }
  // If event type is VM.OFF
  case lib.VM_OFF: {
    err := onVmOff(event, pafwConfig)
    if (err != nil) {
      glog.Error("Failed to process event. Error: ", err)
    }
  }
  }
  return err
}

// This is a generic method to perform HTTP requests.
//
// Args:
//    url : Website url to which request will be made.
//    httpClient : Predefined interface of HTTP client.
// Returns:
//    Response : HTTP response for the request.
//    error : Error, if any.

func doHttpRequest(url string, httpClient  http.Client) (*http.Response, error) {
  var resp *http.Response
  url = strings.Replace(url, " ", "%20", -1)
  request, err := http.NewRequest("GET", url, nil)
  if (err != nil) {
    glog.Error("failed to create request. ", err)
    return resp, err
  }
  resp, err = httpClient.Do(request)
  if (err != nil || resp.StatusCode != 200) {
    glog.Error("Request failed.", err)
    glog.Error("HTTP status code :", resp.StatusCode)
    return resp, err
  }
  respData, _ := ioutil.ReadAll(resp.Body)
  resp.Body = ioutil.NopCloser(bytes.NewBuffer(respData))
  matched, _ := regexp.MatchString("<response status.*success.*?>", string(respData))
  if matched == false {
    glog.Errorf("Url failed.\nURL: %s\nResponse: %s", url, string(respData))
    return resp, errors.New(string(respData))
  }
  return resp, err
}

// This method will process the VM.ON event. It will make the
// intended configuration on the target virtual appliance.
//
// Args:
//    event : Event object containing the event data sent by the listener.
//    pafwConfig : PAFW Event consumer configuration data structure.
// Returns:
//    error : Error, if any.
func onVmOn(event schema.Event, pafwConfig config.PAFWConfig) (error) {

  var err error
  var url, urlStr string
  glog.Infof("Processing %v event.", event.Event_Type)

  // Setting Http Client
  tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true},}
  httpClient := http.Client{}
  httpClient.Transport = tr

  // Login to Palo Alto Firewall VM.
  glog.Info("Login to Firewall VM and generate session key")
  url = fmt.Sprintf("https://%s/api/?type=keygen&user=%s&password=%s",
                    pafwConfig.PAFWInstanceConfig.IP,
                    pafwConfig.PAFWInstanceConfig.Username,
                    pafwConfig.PAFWInstanceConfig.Password)
  resp, err := doHttpRequest(url, httpClient)
  if err != nil {
    glog.Errorf("Login to Firewall VM Failed.")
    return err
  }

  // Session key extraction after login to Firewall VM.
  respData, _ := ioutil.ReadAll(resp.Body)
  re := regexp.MustCompile(".*<key>(.*?)</key>.*")
  key := re.FindStringSubmatch(string(respData))[1]

  // Prepare Palo Alto REST API for creating or updating Security Policy Rule.
  urlXPath:= "/config/devices/entry[@name='localhost.localdomain']"
  urlXPath = urlXPath + "/vsys/entry[@name='vsys1']"
  baseUrl := fmt.Sprintf("https://%s/api/?key=%s", pafwConfig.PAFWInstanceConfig.IP, key)
  category := event.Data.Metadata.SubMetadata.Categories.NetworkFunctionProvider
  // Temporary Code //
  if(category == "") {
    category = pafwConfig.PAFWInstanceConfig.Category
  }
  // Temporary Code //
  glog.Infof("Processing VM of Category: %s", category)

  // Tag (Category) creation process.

  tagName := strings.Replace(category, " ", "-", -1)
  glog.Infof("Creating tag by name: %s", tagName)
  url = fmt.Sprintf("%s&type=config&action=get&xpath=%s/tag", baseUrl, urlXPath)
  resp, err = doHttpRequest(url, httpClient)
  if err != nil {
    glog.Errorf("Check of tag existense failed. Error: ", err)
    return err
  }
  respData, _ = ioutil.ReadAll(resp.Body)
  pattern := fmt.Sprintf("<entry name=\"%s\" ", tagName)
  // Check if Tag already created.
  matched, _ := regexp.MatchString(pattern, string(respData))
  if (matched == false) { // Tag does not exist. Create a new Tag.
    urlStr = "%s&type=config&action=set&xpath=%s"
    urlStr = urlStr + "/tag/entry[@name='%s']&element=<color>color2</color>"
    url = fmt.Sprintf(urlStr, baseUrl, urlXPath, tagName)
    resp, err = doHttpRequest(url, httpClient)
  } else { // Tag already exists.
    glog.Warningf("Tag '%s' already exists.", tagName)
  }

  // Create Address entity by VM Name
  address := event.Data.Metadata.Status.Name
  vmIPAddress := event.Data.Metadata.Status.Resources.NICList[0].IPEndPointList[0].IPAddress
  glog.Infof("Creating VM Address by name '%s'/IP Address: '%s'", address,
                                                           vmIPAddress)
  urlStr = "%s&type=config&action=set&xpath=%s/address/entry[@name='%s']"
  urlStr = urlStr + "&element=<ip-netmask>%s</ip-netmask><tag>"
  urlStr = urlStr + "<member>%s</member></tag><description>%s</description>"
  url = fmt.Sprintf(urlStr, baseUrl, urlXPath, address, vmIPAddress,
                    tagName, "Apache Web Server")
  resp, err = doHttpRequest(url, httpClient)
  if err != nil {
    glog.Errorf("Creation of VM Address entity failed.")
    return err
  }

  // Dynamic Address Creation process.
  addressGroup := pafwConfig.PAFWInstanceConfig.AddressGroup
  glog.Infof("Creating Dynamic Address Group by name '%s'.", addressGroup)
  url = fmt.Sprintf("%s&type=config&action=get&xpath=%s", baseUrl, urlXPath)
  url = fmt.Sprintf("%s/address-group/entry[@name='%s']", url, addressGroup)
  resp, err = doHttpRequest(url, httpClient)
  respData, _ = ioutil.ReadAll(resp.Body)
  glog.Infof("Checking if Address Group '%s' already exists.", addressGroup)
  addAddressGroup, tags := checkNGetAddressGroupTagExistence(string(respData),
                                                        addressGroup, tagName)
  if addAddressGroup == true { // Either AddressGroup or Tag is missing. Need to create either one.
    urlStr = "%s&type=config&action=set&xpath=%s/address-group/entry[@name='%s']"
    urlStr = urlStr + "&element=<dynamic><filter>%s</filter></dynamic>"
    url = fmt.Sprintf(urlStr, baseUrl, urlXPath, addressGroup, tags)
    resp, err = doHttpRequest(url, httpClient)
    if(err != nil) {
      glog.Errorf("Creation of Dynamic Address group '%s' failed.", addressGroup)
      return err
    }
  } else { // AddressGroup alredy exist with required tag.
    glog.Warningf("Dynamic Address Group '%s' already exists.", addressGroup)
  }

  // Create Security Policy Rule
  policyRule := pafwConfig.PAFWInstanceConfig.SecurityPolicyRule
  glog.Infof("Creating security policy rule by name: %s", policyRule)
  urlStr = "%s&type=config&action=get&xpath=%s"
  urlStr = urlStr + "/rulebase/security/rules/entry[@name='%s']"
  url = fmt.Sprintf(urlStr, baseUrl, urlXPath, policyRule)
  glog.Infof("Checking if Security Policy Rule '%s' already exists.",policyRule)
  resp, err = doHttpRequest(url, httpClient)
  if(err != nil) {
      glog.Warningf("Check of existing security policy rule failed.")
  }
  respData, _ = ioutil.ReadAll(resp.Body)
  pattern = fmt.Sprintf("<entry name=\"%s\" ", policyRule)
  matched, _ = regexp.MatchString(pattern, string(respData))
  if matched == false { // Security Policy Rule does not exist. Create new rule.
    glog.Infof("Creating Security Policy Rule '%s'.", policyRule)
    urlStr = "%s&type=config&action=set&xpath=%s"
    urlStr = urlStr + "/rulebase/security/rules/entry[@name='%s']"
    urlStr = urlStr + "&element=<to><member>untrust</member></to>"
    urlStr = urlStr + "<from><member>trust</member></from>"
    urlStr = urlStr + "<source><member>%s</member></source>"
    urlStr = urlStr + "<destination><member>any</member></destination>"
    urlStr = urlStr + "<source-user><member>any</member></source-user>"
    urlStr = urlStr + "<application><member>any</member></application>"
    urlStr = urlStr + "<service><member>application-default</member></service>"
    urlStr = urlStr + "<hip-profiles><member>any</member></hip-profiles>"
    urlStr = urlStr + "<action>allow</action>"
    url = fmt.Sprintf(urlStr, baseUrl, urlXPath, policyRule, addressGroup)
    resp, err = doHttpRequest(url, httpClient)
    if(err != nil) {
      glog.Errorf("Creation of Security Policy Rule failed.")
      return err
    }
  } else {
    glog.Warningf("Security Policy Rule '%s' already exists.", policyRule)
  }

  // Commit Changes
  glog.Info("Commit changes.")
  url = fmt.Sprintf("%s&type=commit&cmd=<commit><force></force></commit>",
                     baseUrl)
  resp, err = doHttpRequest(url, httpClient)
  if(err != nil) {
    glog.Errorf("Failed to commit changes.")
    return err
  }

  return err
}

// Function to check AddressGroup & Tag existence on firewall VM.
//
// Args:
//   data : AddressGroup data string.
//   addressGroup : AddressGroup name of which existense will be checked in data.
//   tags : tag name of which existense will be checked in data.
// Returns:
//   addAddressGroup(bool) : True if input addressGroup does not exist in data else False.
//   tags(string) : Modified existing Tag string appended with input Tag.
//
func checkNGetAddressGroupTagExistence(data string, addressGroup string, tags string) (bool, string) {
  addAddressGroup := true
  pattern := fmt.Sprintf("<entry name=\"%s\" ", addressGroup)
  if matched, _ := regexp.MatchString(pattern, data); matched == true {
    // Get tags from existing Dynamic Address Group
    re := regexp.MustCompile("<filter.*?\">(.*)</filter>")
    matches := re.FindStringSubmatch(data)
    if len(matches) > 1 {
      existingTagList := matches[1]
      pattern = fmt.Sprintf("\\b(%s)\\b", tags)
      tmpExistingTagList := strings.Replace(existingTagList, "'", "", -1)
      if matched, _ = regexp.MatchString(pattern, tmpExistingTagList); matched == true {
        addAddressGroup = false
      } else {
        tags = fmt.Sprintf("%s or %s", existingTagList, tags)
      }
    }
  }
  return addAddressGroup, tags
}

// This method will process the VM.OFF event.
// It will make the intended configuration on the target virtual appliance.
//
// Args:
//    event : Event object containing the event data sent by the listener.
//    pafwConfig : PAFW Event consumer configuration data structure.
// Returns:
//    error : Error, if any.
func onVmOff(event schema.Event, pafwConfig config.PAFWConfig) (error) {
  var err error
  var url string
  glog.Infof("Processing %v event.", event.Event_Type)
  // Remove VM Address from Security Policy Rule.

  // Setting Http Client
  tr := &http.Transport{ TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, }
  httpClient := http.Client{}
  httpClient.Transport = tr

  // Login to Palo Alto Firewall VM.
  glog.Info("Login to Firewall VM and generating key")
  url = fmt.Sprintf("https://%s/api/?type=keygen&user=%s&password=%s",
                    pafwConfig.PAFWInstanceConfig.IP,
                    pafwConfig.PAFWInstanceConfig.Username,
                    pafwConfig.PAFWInstanceConfig.Password)
  resp, err := doHttpRequest(url, httpClient)
  if err != nil {
    glog.Errorf("Login to Firewall VM Failed.")
    return err
  }
  respData, _ := ioutil.ReadAll(resp.Body)
  re := regexp.MustCompile(".*<key>(.*?)</key>.*")
  // Extract session key from response data.
  key := re.FindStringSubmatch(string(respData))[1]

  // Setting up Url XPath variable
  urlXPath := "/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys1']"

  // Delete Address by VM Name & IP Address
  address := event.Data.Metadata.Status.Name
  glog.Infof("Deleting VM '%s'", address)
  url = fmt.Sprintf("https://%s/api/?type=config&action=delete", pafwConfig.PAFWInstanceConfig.IP)
  url = fmt.Sprintf("%s&key=%s&xpath=%s/address/entry[@name='%s']", url, key,
                                                            urlXPath, address)
  resp, err = doHttpRequest(url, httpClient)
  if err != nil {
    glog.Errorf("Delete Address failed.")
    return err
  }
  respData, _ = ioutil.ReadAll(resp.Body)

  // Commit Changes
  glog.Info("Commit changes.")
  url = fmt.Sprintf("https://%s/api/?type=commit&cmd=<commit>", pafwConfig.PAFWInstanceConfig.IP)
  url = fmt.Sprintf("%s<force></force></commit>&key=%s", url, key)
  resp, err = doHttpRequest(url, httpClient)
  if(err != nil) {
    glog.Errorf("Failed to commit changes.")
    return err
  }

  return err
}

// This method will load the PAFW specific config.
//
// Args:
//    None.
// Returns:
//    pafwConfig  : PAFW specific config.
//    error : Error, if any.
func LoadPAFWConfig() (config.PAFWConfig, error) {
  var pafwConfig config.PAFWConfig
  glog.Info("Loading config..")
  // PAFW Event consumer configuration file.
  pafwConfigPath := PAFWConfigDir + "pafw_config.json"
  // Read event consumer configuration file content
  pafwConfigFileContent, err := ioutil.ReadFile(pafwConfigPath)
  if (err != nil) {
    glog.Error("Error reading config.", err)
    return pafwConfig, err
  }
  // Unmarshal event consumer file content
  err = json.Unmarshal([]byte(pafwConfigFileContent), &pafwConfig)
  if (err != nil) {
    glog.Error("Failed to unmarshal config.",err)
  }
  // Decode PAFW instance Base64 encoded password.
  // Note : Developers can exercise their own encryption mechanism for credentials.
  decoded, err := base64.StdEncoding.DecodeString(pafwConfig.PAFWInstanceConfig.Password)
  if err != nil {
    glog.Error("Decode error:", err)
  } else {
    pafwConfig.PAFWInstanceConfig.Password = string(decoded)
  }
  // Decode Nutanix Prism Base64 encoded password.
  // Note : Developers can exercise their own encryption mechanism for credentials.
  decoded, err = base64.StdEncoding.DecodeString(pafwConfig.NutanixClusterConfig.Password)
  if err != nil {
    glog.Error("Decode error:", err)
  } else {
    pafwConfig.NutanixClusterConfig.Password = string(decoded)
  }
  if err != nil {
    glog.Error("Decode error:", err)
  }
  return pafwConfig, err
}
