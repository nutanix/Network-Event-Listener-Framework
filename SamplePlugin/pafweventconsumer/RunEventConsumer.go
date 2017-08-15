// Copyright (c) 2017 Nutanix Inc. All rights reserved.
//
// Main package for the event consumer.
package main

import (
  "aplos/partners/WebhooksListener/lib"
  "aplos/partners/WebhooksListener/webhook"
  consumer "aplos/partners/pafweventconsumer/impl"
  "github.com/golang/glog"
  "flag"
  "fmt"
  "os"
)

func usage() {
  fmt.Fprintf(os.Stderr, "usage: example -stderrthreshold=[INFO|WARN|FATAL] -log_dir=[string]\n", )
  flag.PrintDefaults()
  os.Exit(2)
}

func init() {
  flag.Usage = usage
  // NOTE: This next line is key you have to call flag.Parse() for the command line
  // options or "flags" that are defined in the glog module to be picked up.
  flag.Parse()
}

func main() {
  var webhooksListener WebhooksListener.WebhooksListener

  // Define the networking events that will be subscribed by the event consumer.
  events := []string{lib.VM_ON, lib.VM_OFF}
  // Load the event consumer configuration file.
  pafwConfig, err := consumer.LoadPAFWConfig()
  if (err != nil) {
    glog.Errorf("Failed to load config. Cannot proceed. Error:- %v", err)
    return
  }
  // Initialize listener.
  webhooksListener, err = webhooksListener.Initialize(
    pafwConfig.NutanixClusterConfig.IP,
    pafwConfig.NutanixClusterConfig.Port,
    pafwConfig.NutanixClusterConfig.Username,
    pafwConfig.NutanixClusterConfig.Password)
  if (err != nil) {
    glog.Errorf("Failed to initialize listener.Erro: %v", err)
    return
  }
  webhooksListener.ListenerState = make(chan string)
  // Event Consumer Register for Events.
  webhooksListener.RegisterForEvents(events, consumer.PAFWEventConsumer{})
  listenerStateMsg, listenerRunning := <-webhooksListener.ListenerState
  for listenerRunning == true {
    if(listenerStateMsg != "") {
      glog.Info("Message from Listener: ", listenerStateMsg)
    }
    listenerStateMsg, listenerRunning = <-webhooksListener.ListenerState
  }
}
