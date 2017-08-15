// Copyright (c) 2017 Nutanix Inc. All rights reserved.

// The Listener interface is implemented by the Nutanix provided webhook
// listener library. This library simplifies the process of registering for
// webhook events with a Nutanix AHV cluster. The event consumer that uses this
// library is expected to instantiate the Listener object, and provide the
// cluster details through the Initialize call.
// Functionality provided by the interface is as follows -
// 1. Register for events by creating corresponding webhook.
// 2. Listen for events.

package interfaces

type Listener interface {
  // Interface for the listener.

  // This method will be invoked by the event consumer in order to initialize
  // the Listener object. It will receive the Nutanix cluster details from the
  // event consumer and initialize the listener after validating the cluster
  // details (whether the credentials are correct etc.)
  //
  // Args:
  //    ip : External IP address of the Nutanix cluster.
  //    port : Port of the Nutanix cluster (Prism port)
  //    username : Username for authentication to the Nutanix cluster.
  //    password : Password for authentication to the Nutanix cluster.
  // Returns:
  //    Listener : Instance of the listener
  //    error : Error, if any.
  Initialize(ip string, port string, username string,
    password string) (Listener, error)

  // This method allows the event consumer to register itself with the Nutanix
  // cluster subscribing to relevant events occurring on the cluster through
  // webhooks.
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
  //             invoked by the listener on the occurrence of the subscribed
  //             events.
  // Returns:
  //    error : Error, if any.
  RegisterForEvents(events []string, eventConsumer EventConsumer) (error)
}
