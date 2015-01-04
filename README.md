# Ninja Sphere - Chromecast Driver


[![Build status](https://badge.buildkite.com/5bb6d3821af48a9a5e945ef5c43645cf8ccabfe99d217ba2ba.svg)](https://buildkite.com/ninja-blocks-inc/sphere-driver-go-chromecast)
[![godoc](http://img.shields.io/badge/godoc-Reference-blue.svg)](https://godoc.org/github.com/ninjasphere/driver-go-chromecast)
[![MIT License](https://img.shields.io/badge/license-MIT-yellow.svg)](LICENSE)
[![Ninja Sphere](https://img.shields.io/badge/built%20by-ninja%20blocks-lightgrey.svg)](http://ninjablocks.com)
[![Ninja Sphere](https://img.shields.io/badge/works%20with-ninja%20sphere-8f72e3.svg)](http://ninjablocks.com)

---


### Introduction
This is a driver for the Google Chromecast, allowing it to be used as part of Ninja Sphere.

It currently only supports applications that export the "urn:x-cast:com.google.cast.media" namespace, though that will be expanded as other namespaces are found and documented.

### Supported Sphere Protocols

| Name | URI | Supported Events | Supported Methods |
| ------ | ------------- | ---- | ----------- |
| volume | [http://schema.ninjablocks.com/protocol/volume](https://github.com/ninjasphere/schemas/blob/master/protocol/volume.json) | set, volumeUp, volumeDown, mute, unmute, toggleMute | state |
| media-control | [http://schema.ninjablocks.com/protocol/media-control](https://github.com/ninjasphere/schemas/blob/master/protocol/media-control.json) | play, pause, togglePlay  | playing, paused, stopped, buffering, busy, idle, inactive |

#### To Do
* Add the *media* protocol, to gain media meta-data and play position/seeking.

* Finish the *media-control* protocol support, adding next/previous/stop etc.

* Add additional support for the Plex applcation, collecting advanced meta-data from the Plex Server's REST api.

* See if notifications are possible without using a custom application (i.e. by adjusting the current player's CSS, or using additional namespaces)

* Expand support for other applications, especially to pull extra meta-data (i.e. YouTube)

#### Can't Do
* Owing to private APIs, the Netflix application can't currently be supported.

### Requirements

* Go 1.3

### Dependencies

https://github.com/ninjasphere/go-castv2

### Building

This project can be built with `go build`, but a makefile is also provided.

### Running

`DEBUG=* ./driver-go-chromecast`

### Options

The usual Ninja Sphere configuration and parameters apply, but these are the most useful during development.

* `--autostart` - Doesn't wait to be started by Ninja Sphere
* `--mqtt.host=HOST` - Override default mqtt host
* `--mqtt.port=PORT` - Override default mqtt host

### More Information

More information can be found on the [project site](http://github.com/ninjasphere/driver-go-chromecase) or by visiting the Ninja Blocks [forums](https://discuss.ninjablocks.com).

### Contributing Changes

To contribute code changes to the project, please clone the repository and submit a pull-request ([What does that mean?](https://help.github.com/articles/using-pull-requests/)).

### License
This project is licensed under the MIT license, a copy of which can be found in the [LICENSE](LICENSE) file.

### Copyright
This work is Copyright (c) 2014-2015 - Ninja Blocks Inc.
