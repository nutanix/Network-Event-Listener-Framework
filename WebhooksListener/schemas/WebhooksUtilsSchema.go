// Copyright (c) 2017 Nutanix Inc. All rights reserved.
//
// Description:
//
// The listener utils configuration schema file comprises of data structure
// representing the different parameters that will be consumed by the
// different methods which constitute listener utils. In order to enable the
// listener utils to consume the parameters, we need to define a schema file.
// Here we need to match every relevant parameter field to a relevant data
// structure in the (schema) file.
//
package schema

import "net/http"

// Generic struct to hold request details.
type Request struct {
  Method string
  URL string
  RequestData string
  Credentials Credentials
  Transport *http.Transport
}

type Credentials struct {
  Username string
  Password string
}
