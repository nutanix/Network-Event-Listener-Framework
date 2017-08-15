# Network-Event-Listener-Framework

# Introduction : 
The Webhooks Listener-Plugin library consists of two component libraries written in GoLang: WebHook Listener Libraries and Plugin (Event Consumer) Libraries. 


# WebHook Listener Libraries

1) Using the webhooks listener libraries customers/partners/network product companies can create a listener program that can subscribe to relevant networking events occuring on AHV through Nutanix's AOS WebHook APIs . 

2)The listener program will register a HTTP url with the AOS webhooks API.

3) The networking events as and when they occur  will be pushed to the listener through a HTTP POST call of the url.

# Plugin Libraries

1) The plugin component will accept the events passed to it from the listner.

2) The plugin component will interpret those events and then drive policy management on the third party networking product by making relevant calls to their respective end-point.

# Contents

1) Nutanix WebHooks Listener and Plugin Framework Tutorial.

2) WebhooksListener: Nutanix WebHooks Listener Library Files 

3) pafweventconsumer : Plugin Library Files (Sample implementation of Palo Alto VMSeries FW plugin)

4) panfweventservice : Sample code for exposing the plugin as a service


# Assumptions

1) Developers should have exposure to GoLang programming language.

2) The plugin implementation steps described in the tutorial are for illustration purposes only. Developers are expected to exercise their own discretion in determining their own pedagogy of approach to write their own plugin.They can use the implementation steps given in the user guide for guidance.

3) The sample code associated with this repository is for illustration purpose . The sample code implements a plugin for Palo Alto VMSeries FW for KVM 

4) Developers are free to download , modify or enhance the libraries for their own use case and are expected to maintain them.


# Dependency

In order to leverage Nutanix AOS WebHooks API , one should be on AOS version >= 5.1.1

