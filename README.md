# go-splitflap

**This project is in an early ALPHA stage, many things are expected to change, or not function as expected! Use at your own risk!**

Many thanks to [serdiev](https://github.com/Serdiev/splitflap-backend) for sharing their Golang implementation of the Serial protocol for the Splitflap!

## Core Concepts

* **Routines** are the "building blocks" of splitflap display functionality. They are (hopefully) simple, well-contained code files that do a single static or dynamic presentation of information, with well-defined configuration. 

  - For example, the `CLOCK` routine displays the current time for a particular timezone, defined with parameters that enable 12/24 hr formatting, add AM/PM suffix, etc.

* **Dashboards** represent the combination of routines to form more interesting displays. For example, combining `TEXT`, `CLOCK`, and `WEATHER` routines can provide unified information for a given location: 

  ```
  New York City: 
  1:23 PM   68F
  ```

* **Rotations** are the final (optional) combination of Dashboards into a time-based scheduling of displays. They enable functionality like, for example, displaying the current time for 9 minutes, followed by the weather for 1 minute, and then repeating ad infinitum.

## Backend Development/Installation

Install [Go 1.24+](https://go.dev/doc/install). Then navigate to the backend folder (`cd backend`) and run `go build -o server main.go`

This will produce an executable `server`, which you should run with the appropriate `--port` value corresponding to the port that connects to your splitflap TTGO.

On Windows, this will be something like `--port=COM5` (for example), whereas on Linux, you may need a full path like `/dev/tty/...` (use `lsusb` to help discover what port you need).

## Frontend Development

Install [nodeJS](https://nodejs.org/en/download) and [yarn](https://classic.yarnpkg.com/lang/en/docs/install/#windows-stable), then `cd web-ui` and run `yarn` followed by `yarn dev`. If you 

## Installation

TODO: Serve pre-built UI and Backend executables