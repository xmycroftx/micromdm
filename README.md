mircomdm is a Mobile Device Management for Apple Devices(primarily OS X Macs).

While I intend to implement all the commands defined by Apple in the spec, the current focus is on implementing the features necessary to fit Apple's new(er) management tools (MDM, VPP, DEP) into existing enterprise environments.

# Overview
This repo is under heavy development. Here are a few goals for the first usable release:

* communicate with DEP endpoints to fetch/assign profiles and sync devices(in progress)
* support InstallApplication, and InstallProfile payloads (already done)
* allow managing MDM commands in a queue(done) and trigger push notifications(in progress, with Apple's new http2 push gateway)
* manage devices by role/organizational unit and enable default workflows that are triggered when a device checks in the first time.
* create a workable ui.


# architecture/dependencies
micromdm is an open source project written as an http server in [Go](https://golang.org/) and:
* uses [PostgreSQL](http://www.postgresql.org/) for long lived data(devices, users, profiles
* uses redis to queue MDM Commands.
* is a collection of [go-kit](https://github.com/go-kit/kit) services,
deployed as a single binary. This should keep development simple and allow scaling different parts as needed.
* exposes metrics data in [Prometheus](https://prometheus.io/) format.
* is api driven - should allow for integrations with other tools and make UI development more flexible.


