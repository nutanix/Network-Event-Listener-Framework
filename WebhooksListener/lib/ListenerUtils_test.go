// Copyright (c) 2017 Nutanix Inc. All rights reserved.
//
// This test package apply various unit tests on utlity
// functions provided by listener library.
//

package lib

import (
	"testing"
	"net"
	"fmt"
	"strconv"
	"math/rand"
)

// Test to validate HTTP web request successful.
func TestPrepareDoRequest(t *testing.T) {
  inputUrl := "https://www.nutanix.com/"
  httpMethod := "GET"
  var errMsg string
  req := PrepareRequest(inputUrl, "", "", httpMethod)
  if !(req.URL == inputUrl && req.Method == httpMethod) {
    errMsg = fmt.Sprintf("Failed to prepare HTTP request.\n")
    errMsg = fmt.Sprintf("%sInput Url: %s, Output Url: %s\n", errMsg, inputUrl, req.URL)
    errMsg = fmt.Sprintf("%sInput Method: %s, Output Method: %s\n", errMsg, httpMethod, req.Method)
    t.Errorf(errMsg)
  }
  response, _ := DoRequest(req)
  if (response.StatusCode < 0){
    t.Errorf("Failed to execute HTP request for url: %s\n", inputUrl)
  }
}

// Test to verify duplicate element removal from list.
func TestRemoveDuplicates(t *testing.T) {
  input := []string{"Unit", "Testing", "Unit", "Testing"}
  output := RemoveDuplicates(input)
  if !(len(output) == 2 && output[0] == "Unit" && output[1] == "Testing") {
    t.Errorf("Failed to Remove Duplicates from list:\nInput: %s\nOutput: %s\n", input, output)
  }
}

// Test to check if any port is available.
func TestCheckPortAvailability(t *testing.T) {
  portAvailable := false
  var port string
  for portAvailable == false {
    port = strconv.Itoa(rand.Intn(10000))
    testSocket, err := net.Listen("tcp", ":" + port)
    if(err == nil) {
      testSocket.Close()
      output := CheckPortAvailability(port)
      if output != nil {
        t.Errorf("Failed to validate available port:- ", port)
      } else {
        break  // Test case passed
      }
    }
  }
}
