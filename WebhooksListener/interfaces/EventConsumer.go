// Copyright (c) 2017 Nutanix Inc. All rights reserved.

// This interface must be implemented by the event consumer. The listener
// library will invoke the OnEvent method on this interface to notify about
// events happening on the Nutanix cluster.
//
// Functionality provided by the interface is as follows -
// 1. Handle events.

package interfaces

import "aplos/partners/WebhooksListener/schemas"

type EventConsumer interface {
  // Interface for the event consumer.

  // Callback method to handle the event received from the listener.
  // This method will have to be implemented by the partner as per their
  // requirement.
  // Args:
  //    event : Event object containing the event data sent by the listener.
  // Returns:
  //    error : Error, if any.
  OnEvent(event schema.Event) (error)
}
