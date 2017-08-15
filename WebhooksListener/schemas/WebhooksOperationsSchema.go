// Copyright (c) 2017 Nutanix Inc. All rights reserved.
//
// Description:
//
// The web hooks operations configuration schema file comprises of data structure
// representing the different data elements that are used for web hooks management.
// In order to enable the listener to consume the parameters that are required for
// web hooks management, we need to define a schema file. Here we need to match
// every relevant parameter field to a relevant data structure in the (schema) file.
//
package schema

// Create Webhook
type WebhookCreationSpec struct {
  Metadata Metadata `json:"metadata"`
  Spec Spec `json:"spec"`
  ApiVersion string `json:"api_version"`
}

type Metadata struct {
  Kind string `json:"kind"`
  SpecVersion int `json:"spec_version"`
}

type Spec struct {
  Name string `json:"name"`
  Resources Resources `json:"resources"`
  Description string `json:"description"`
}

type Resources struct {
  PostUrl string `json:"post_url"`
  EventsFilterList []string `json:"events_filter_list"`
}

// List webhooks.
type WebhooksListSpec struct{
  Kind string `json:"kind"`
}

// Used to store details of the existing webhooks created by this listener.
type WebhookCache struct {
  CacheID int `json:"cache_id"`
  WebhookData Webhook `json:"webhook_data"`
}

type CurrentWebhooks struct {
  ApiVersion string `json:"api_version"`
  Metadata CurrentWebhooksMetadata `json:"metadata"`
  Entities []Webhook `json:"entities"`
}

type CurrentWebhooksMetadata struct {
  TotalMatches int `json:"total_matches"`
  Kind string `json:"kind"`
  Length int `json:"length"`
  Offset int `json:"offset"`
}
type Webhook struct {
  Status WebhookStatus `json:"status"`
  Spec WebhookSpec `json:"spec"`
  ApiVersion string `json:"api_version"`
  Metadata WebhookMetadata `json:"metadata"`
}

type WebhookStatus struct {
  State string `json:"state"`
}

type WebhookSpec struct {
  Name string `json:"name"`
  Resources WebhookResources `json:"resources"`
  Description string `json:"description"`
}

type WebhookResources struct {
  PostURL string `json:"post_url"`
  EventsFilterList []string `json:"events_filter_list"`
}

type WebhookMetadata struct {
  Kind string `json:"kind"`
  UUID string `json:"uuid"`
  SpecVersion int `json:"spec_version"`
}
