// Copyright (c) 2017 Nutanix Inc. All rights reserved.

// Implementation of the WebhooksListener interface for webhooks.
package WebhooksListener

import (
  "fmt"
  "errors"
  "encoding/json"
  "github.com/golang/glog"
  "io/ioutil"
  "reflect"
  "strings"
  "aplos/partners/WebhooksListener/schemas"
  "aplos/partners/WebhooksListener/lib"
  "aplos/partners/WebhooksListener/interfaces"
  "net/http"
)

type WebhooksListener struct {
  // Type that implements the Listener interface.

  // Public properties of the WebhooksListener.
  ListenerPort string // Allows the event consumer to define the local port.
  ListenerState chan string // Message channel to communcate WebhooksListener status.

  // Private properties of the WebhooksListener.
  clusterIp string
  clusterPort string
  clusterUsername string
  clusterPassword string
  eventConsumer interface{}
  listenerIp string

}

// Response status of various webhook operation.
const (
  // Webhook operation is complete.
  completeStatus = "COMPLETE"

  // when Webhook operation request is accepted and processing,
  // it gives status "PENDING" with status code 202.
  pendingStatus = "PENDING"
  pendingStatusCode = 202

)

// This method will be invoked by the event consumer in order to initialize
// the Listener object.
// It will receive the Nutanix cluster details from the event consumer &
// initialize the WebhooksListener after validating the cluster details
// (whether the credentials are correct etc.)
//
// Args:
//    ip : External IP address of the Nutanix cluster.
//    port : Port of the Nutanix cluster (Prism port)
//    username : Username for authentication to the Nutanix cluster.
//    password : Password for authentication to the Nutanix cluster.
// Returns:
//    Listener : Instance of the WebhooksListener
//    error : Error, if any.
func (webhooksListener WebhooksListener) Initialize(ip string, port string,
  username string, password string) (WebhooksListener, error) {
  var err error
  glog.Info("Initializing listener..")
  webhooksListener.clusterIp = ip
  webhooksListener.clusterPort = port
  webhooksListener.clusterUsername = username
  webhooksListener.clusterPassword = password

  if (webhooksListener.ListenerPort == "") {
    webhooksListener.ListenerPort = lib.DefaultListenerPort
  }

  // Check network connectivity with the cluster.
  glog.Info("Verifying connectivity with cluster.")
  localIp, err := lib.CheckOutboundConnectivity(webhooksListener.clusterIp,
    webhooksListener.clusterPort)
  if (err != nil) {
    glog.Error("Failed to verify connectivity with cluster.", err)
    return webhooksListener, err
  }
  webhooksListener.listenerIp = localIp

  // Check if given credentials are valid.
  glog.Info("Authenticating cluster credentials.")
  requestURL := fmt.Sprintf("https://%s:%s%s", webhooksListener.clusterIp,
    webhooksListener.clusterPort, lib.GetCurrentUser)
  request := lib.PrepareRequest(requestURL, webhooksListener.clusterUsername,
                                  webhooksListener.clusterPassword, "GET")
  glog.Info("Making http request : %v", requestURL)
  response, err := lib.DoRequest(request)
  if (err != nil) {
    glog.Error("Unable to login cluster with given credentials. Error: ", err)
    return webhooksListener, err
  }
  if (response.StatusCode != 200) {
    msg := fmt.Sprintf("Error verifying cluster credentials. HTTP status " +
      "code : %v", response.StatusCode)
    glog.Error(msg)
    err = errors.New(msg)
    return webhooksListener, err
  }
  return webhooksListener, err
}

// This method allows the event consumer to register itself with the Nutanix
// cluster subscribing to relevant events occurring on the cluster through
// webhooks. It will assign the event consumer's interface as the point of
// invocation on occurrence of the subscribed event.
//
// Args:
//    events : The list of events that the caller is interested in.
//             Must be one of:
//             [
//               "VM.CREATE", "VM.DELETE", "VM.ON", "VM.OFF", "VM.UPDATE",
//               "VM.MIGRATE", "VM.NIC_PLUG", "VM.NIC_UNPLUG"
//             ]
//    eventConsumer :
//             Interface reference to the event consumer. This interface will
//             be expected to provide an OnEvent method which will be
//             invoked by the WebhooksListener on the occurrence of the subscribed
//             event.
// Returns:
//    error : Error, if any.
func (webhooksListener WebhooksListener) RegisterForEvents(events []string,
  eventConsumer interfaces.EventConsumer) (error) {
  var err error
  glog.Infof("Registering for events %v", events)

  err = lib.CheckPortAvailability(webhooksListener.ListenerPort)
  if (err != nil) {
    glog.Errorf("Port %s cannot be used. %s.", webhooksListener.ListenerPort,
      err)
    return err
  }

  // Create/update webhook for the given events.
  err = webhooksListener.createOrUpdateWebhook(events)
  if (err != nil) {
    glog.Error("Failed to register.", err)
    return err
  }

  // Start HTTP WebhooksListener
  webhooksListener.eventConsumer = eventConsumer
  go webhooksListener.startListener()
  return err
}

// This method opens a HTTP socket on the listener's port & listens for event
// notifications from webhooks on the listener's callback URL.
//
// Args:
//    None.
// Returns:
//    None.
func (webhooksListener WebhooksListener) startListener() (error){
  var err error
  if(webhooksListener.ListenerPort == "") {
    webhooksListener.ListenerPort = lib.DefaultListenerPort
  }
  webhooksListener.ListenerState <- "Starting HTTP Listener .."
  webhooksListener.ListenerPort = ":" + webhooksListener.ListenerPort
  http.HandleFunc(lib.ListenerCallbackURL, webhooksListener.onEvent)
  err = http.ListenAndServe(webhooksListener.ListenerPort, nil)
  if err != nil {
    webhooksListener.ListenerState <- fmt.Sprintf("Error occured: %s",
                                                   err.Error())
    glog.Error("Listener error: ", err)
  }
  webhooksListener.ListenerState <- "Listener Closed"
  close(webhooksListener.ListenerState)
  return err
}

// This method will be invoked when the WebhooksListener receives an event. It will
// invoke the event consumer's callback method & pass the received event to
// the event consumer.
//
// Args:
//    Note : Both these args are required in the method signature in order to
//    allow the net/http package to use it for invocation.
//    responseWriter : HTTP ResponseWriter object to provide a response to
//                     the caller if required.
//    request : HTTP Request object as received from the caller (i.e webhooks)
// Returns:
//    None.
func (webhooksListener WebhooksListener) onEvent(
  responseWriter http.ResponseWriter, request *http.Request) {
  var event schema.Event

  glog.Info("Received event.")

  // Read event JSON from request.
  body, err := ioutil.ReadAll(request.Body)
  if err != nil {
    glog.Error("Error reading input request body. Cannot proceed.", err)
    return
  }
  eventData := string(body)
  glog.Info("Event data : " + eventData)
  err = json.Unmarshal([]byte(eventData), &event)
  if err != nil {
    glog.Error("Failed to unmarshal event. Cannot proceed.", err)
    return
  }

  // Get the event consumer's callback method & invoke.
  method := reflect.ValueOf(webhooksListener.eventConsumer).MethodByName(
    lib.EventConsumerCallbackMethod)
  if (method.IsValid()) {
    glog.Info("Got event consumer method.")
    methodArgs := make([]reflect.Value, method.Type().NumIn())
    methodArgs[0] = reflect.ValueOf(event)
    method.Call(methodArgs)
  }
}

// This method will create a webhook or update an existing webhook for the
// given events.
//
// Args:
//    events : List of events for which to create or update webhook.
// Returns:
//    error : Error, if any.
func (webhooksListener WebhooksListener) createOrUpdateWebhook(
  events []string) (error) {

  glog.Info("Getting existing webhooks..")
  webhookName := fmt.Sprintf("%s%s",
    lib.WebhookNamePrefix, webhooksListener.listenerIp)
  var webhookListSpec schema.WebhooksListSpec
  webhookListSpec.Kind = lib.WebhookKind

  requestURL := fmt.Sprintf("https://%s:%s%s", webhooksListener.clusterIp,
    webhooksListener.clusterPort, lib.ListWebhooks)
  request := lib.PrepareRequest(requestURL, webhooksListener.clusterUsername,
                                  webhooksListener.clusterPassword, "POST")

  requestData, err := json.Marshal(webhookListSpec)
  if (err != nil) {
    glog.Error("Failed to convert request spec into JSON. ", err)
    return err
  }
  request.RequestData = string(requestData)
  glog.Info("Request data : " + string(requestData[:]))
  glog.Info("Request URL : " + requestURL)

  response, err := lib.DoRequest(request)
  if (err != nil) {
    glog.Error("Failed to get webhooks.", err)
    return err
  }

  var currentWebhooks schema.CurrentWebhooks
  respBytes, err := ioutil.ReadAll(response.Body)
  err = json.Unmarshal(respBytes, &currentWebhooks)
  if (err != nil) {
    glog.Error("Failed to parse current webhooks.", err)
    return err
  }

  glog.Info("Total existing webhooks:",
    currentWebhooks.Metadata.TotalMatches)
  postUrl := fmt.Sprintf("http://%s:%s%s", webhooksListener.listenerIp,
    webhooksListener.ListenerPort,
    lib.ListenerCallbackURL)
  glog.Info("Looking for webhook with url :", postUrl)
  var webhookToUpdate schema.Webhook
  for _, webhook := range currentWebhooks.Entities {
    glog.Info("Webhook url :", webhook.Spec.Resources.PostURL)
    if (webhook.Spec.Resources.PostURL == postUrl) {
      glog.Info("Found matching webhook.")
      webhookToUpdate = webhook
      break
    }
  }

  var requestMethod string
  var eventList []string
  var specVersion int
  if (webhookToUpdate.Metadata.UUID == "") {
    glog.Info("No existing webhook found. Creating new webhook.")
    requestURL = fmt.Sprintf("https://%s:%s%s", webhooksListener.clusterIp,
      webhooksListener.clusterPort, lib.CreateWebhook)
    requestMethod = "POST"
    eventList = events
    specVersion = 0
  } else {
    glog.Info("Updating existing webhook.")
    requestURL = fmt.Sprintf("https://%s:%s%s%s", webhooksListener.clusterIp,
      webhooksListener.clusterPort, lib.UpdateWebhook,
      webhookToUpdate.Metadata.UUID)
    requestMethod = "PUT"
    eventList = append(webhookToUpdate.Spec.Resources.EventsFilterList,
      events...)
    eventList = lib.RemoveDuplicates(eventList)
    specVersion = webhookToUpdate.Metadata.SpecVersion
  }

  var webhookCreationSpec schema.WebhookCreationSpec
  webhookCreationSpec.Metadata.Kind = lib.WebhookKind
  webhookCreationSpec.Metadata.SpecVersion = specVersion
  webhookCreationSpec.Spec.Name = webhookName
  webhookCreationSpec.Spec.Resources.PostUrl = string(postUrl)
  webhookCreationSpec.ApiVersion = "3.0"
  webhookCreationSpec.Spec.Resources.EventsFilterList = eventList

  request = lib.PrepareRequest(requestURL, webhooksListener.clusterUsername,
                                  webhooksListener.clusterPassword, requestMethod)

  requestData, err = json.Marshal(webhookCreationSpec)
  if (err != nil) {
    glog.Error("Failed to convert request spec into JSON. ", err)
    return err
  }
  glog.Info("Request URL : " + requestURL)
  glog.Info("Request data : " + string(requestData[:]))
  request.RequestData = string(requestData)

  response, err = lib.DoRequest(request)
  if (err != nil) {
    glog.Error("Failed to perform webhook operation.", err)
    return err
  }
  if (response.StatusCode == pendingStatusCode) { // Webhook request is accepted and processing.
    respBytes, err = ioutil.ReadAll(response.Body)
    var webhook schema.Webhook
    err = json.Unmarshal(respBytes, &webhook)
    if (err != nil) {
      glog.Error("Failed to parse current webhooks.", err)
      return err
    }
    if (webhook.Status.State == pendingStatus) {
      requestURL = fmt.Sprintf("https://%s:%s%s", webhooksListener.clusterIp,
      webhooksListener.clusterPort, lib.GetWebhook)
      requestURL = strings.Replace(requestURL, "{uuid}", webhook.Metadata.UUID, 1)
      request = lib.PrepareRequest(requestURL, webhooksListener.clusterUsername,
                                  webhooksListener.clusterPassword, "GET")
      response, err = lib.DoRequest(request)
      if (err != nil) {
        glog.Error("Failed to perform webhook operation.", err)
        return err
      }
      respBytes, err = ioutil.ReadAll(response.Body)
      err = json.Unmarshal(respBytes, &webhook)
      if(response.StatusCode == 200 && webhook.Status.State == completeStatus) {
        glog.Info("Webhook registration complete.")
      }
    }
  }
  glog.Info("Successfully completed webhook operation.")

  return err
}
