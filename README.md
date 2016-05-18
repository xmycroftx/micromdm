mircomdm is a Mobile Device Management server for Apple Devices(primarily OS X macs).

While I intend to implement all the commands defined by Apple in the spec, the current focus is on implementing the features necessary to fit Apple's new(er) management tools (MDM, VPP, DEP) into existing enterprise environments.

This project now has a website with updated documentation - https://micromdm.io/



# Overview
**This repo is under heavy development. The current release is only for developers and expert users**

Current status

* Fetch devices from DEP
* Supports `InstallApplication` and `InstallProfile` commands
* Accepts a variety of other MDM payloads such as `OSUpdateStatus` and `DeviceInformation` but just dumps the response from the device to standard output.
* Push notificatioins are supported.
* Configuration profiles and applications can be grouped into a "workflow". The workflow can be assigned to a device.  
Currently the DEP enrollment step will check for a workflow but ignore it. I'll be adding this feature next.
* No SCEP/individual enrollment profiles yet. Need to have an enrollment profile on disk and pass it as a flag.

I set up a public [trello board](https://trello.com/b/js5u4DLV/micromdm-dev-board) to manage what is currently worked on and make notes.

# Getting started
Installation and configuration instructions will be maintained on the [website](https://micromdm.io/getting-started/#installation).


# Notes on architecture
* micromdm is an open source project written as an http server in [Go](https://golang.org/)
* deployed as a single binary. 
* almost everything in the project is a separate library/service. `main` just wraps these together and provides configuratioin flags
* [PostgreSQL](http://www.postgresql.org/) for long lived data(devices, users, profiles, workflows)
* uses redis to queue MDM Commands
* API driven - there will be an admin cli and a web ui, but the server itself is build as a RESTful API.
* exposes metrics data in [Prometheus](https://prometheus.io/) format.


# Workflows
An administrator can group a DEP enrollment profile, a list of applications and a list of configuration profiles into a workflow and assign the workflow to a device.  
If a device has an assigned workflow, `micromdm` will configure the device according to the workflow. 
If you're familiar with Munki's [manifest](https://github.com/munki/munki/wiki/Manifests) feature, workflows work in a similar way.
