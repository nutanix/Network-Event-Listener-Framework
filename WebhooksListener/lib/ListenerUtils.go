// Copyright (c) 2017 Nutanix Inc. All rights reserved.

// This package library provides common utility functions for the
// listener & event consumer.
package lib

import (
  "bytes"
  "crypto/tls"
  "fmt"
  "github.com/golang/glog"
  "io/ioutil"
  "aplos/partners/WebhooksListener/schemas"
  "net/http"
  "net"
  "strings"
)

// This is a generic method to prepare HTTP requests.
//
// Args:
//    requestUrl : Url for request.
//    userName : Authorized user name to make request.
//    password : Password for authorized user to make request.
//    httpMethod : Type of http request. For e.g., PUT, GET, POST
// Returns:
//    Request : Partially prepared http request.
func PrepareRequest(requestURL string, userName string, password string,
                    httpMethod string) (schema.Request) {
  var request schema.Request
  request.Method = httpMethod
  request.URL = requestURL
  request.Credentials.Username = userName
  request.Credentials.Password = password
  request.Transport = &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
  }
  return request
}

// This is a generic method to perform HTTP requests.
//
// Args:
//    request : Object with details required to perform the given request.
// Returns:
//    Response : HTTP response for the request.
//    error : Error, if any.
func DoRequest(request schema.Request) (*http.Response, error) {
  var err error
  var resp *http.Response
  glog.Info("Processing http web request :", request.URL)

  requestDataBytes := []byte(request.RequestData)
  req, err := http.NewRequest(
    request.Method, request.URL, bytes.NewBuffer(requestDataBytes))
  if (err != nil) {
    glog.Error("Failed to create http web request. Error:- ", err)
    return resp, err
  }
  req.Header.Set("Content-Type", "application/json")
  req.SetBasicAuth(request.Credentials.Username, request.Credentials.Password)

  httpClient := http.Client{}
  httpClient.Transport = request.Transport
  resp, err = httpClient.Do(req)
  if (err != nil || resp.StatusCode != 200) {
    if(resp.StatusCode != 202) {
      glog.Error("Request failed. Error:- ", err)
      glog.Error("HTTP status code :", resp.StatusCode)
      // Extract body content from HTTP response.
      respData, _ := ioutil.ReadAll(resp.Body)
      // For Returning, Response data copied back to HTTP response body.
      resp.Body = ioutil.NopCloser(bytes.NewBuffer(respData))
      glog.Error(string(respData))
      return resp, err
    }
  }

  glog.Info("Request successful.", resp.StatusCode)
  return resp, err
}

// This method will check the connectivity with the given IP and port. If
// the connectivity is successful, it will return the IP of the local
// network interface that was used for the outbound connection.
//
// Args:
//    remoteIP : Destination IP address.
//    remotePort : Destination port.
// Returns:
//    string : Local IP address of the network interface used for outbound
//             communication.
func CheckOutboundConnectivity(remoteIp string,
  remotePort string) (string, error) {
  connParam := fmt.Sprintf("%s:%s", remoteIp, remotePort)
  glog.Infof("Checking connectivity with %s", connParam)
  conn, err := net.Dial("tcp", connParam)
  if (err != nil) {
    glog.Errorf("Error while connecting to %s. %s.", connParam, err)
    return "", err
  }
  localIp := conn.LocalAddr().String()
  // Take IP from "IP:Port"
  localIp = strings.Split(localIp, ":")[0]
  glog.Infof("Connectivity successfully verified using local IP %s.", localIp)
  return localIp, err
}

// This method will check the availability of the given port.
//
// Args:
//    port : Port to check.
// Returns:
//    error : Error, if any.
func CheckPortAvailability(port string) (error) {
  var err error
  glog.Infof("Checking if port %s is available.", port)
  testSocket, err := net.Listen("tcp", ":" + port)
  if (err != nil) {
    glog.Error("Cannot connect to port. Error:- ", err)
    return err
  }
  err = testSocket.Close()
  if (err != nil) {
    glog.Error("Failed to close the socket. Error:- ", err)
    return err
  }
  return err
}

// This method will remove duplicate entry from a list of strings.
//
// Args:
//    list : List of strings.
// Returns:
//    []string : List of unique strings.
func RemoveDuplicates(list []string) ([]string) {
  var uniqueItems []string
  tempMap := make(map[string]bool)
  for _, item := range list {
    if _, ok := tempMap[item]; !ok {
      tempMap[item] = true
      uniqueItems = append(uniqueItems, item)
    }
  }
  return uniqueItems
}
