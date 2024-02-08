<!--
SPDX-FileCopyrightText: 2023-2024 Winni Neessen <wn@neessen.dev>

SPDX-License-Identifier: MIT
//-->

# Logranger

[![GoDoc](https://godoc.org/src.neessen.cloud/wneessen/logranger?status.svg)](https://pkg.go.dev/github.com/wneessen/logranger)
[![Go Report Card](https://goreportcard.com/badge/src.neessen.cloud/wneessen/logranger)](https://goreportcard.com/report/github.com/wneessen/logranger)
[![#logranger on Discord](https://img.shields.io/badge/Discord-%23logranger-blue.svg)](https://discord.gg/ysQXkaccXk)
[![REUSE status](https://api.reuse.software/badge/src.neessen.cloud/wneessen/logranger)](https://api.reuse.software/info/github.com/wneessen/logranger)
<a href="https://ko-fi.com/D1D24V9IX"><img src="https://uploads-ssl.webflow.com/5c14e387dab576fe667689cf/5cbed8a4ae2b88347c06c923_BuyMeACoffee_blue.png" height="20" alt="buy ma a coffee"></a>

*Note:* Logranger is still WIP

## Introduction

Logranger is a powerful and intelligent log processing tool written in Go. 
Its main purpose is to efficiently process a large number of incoming syslog messages, 
enabling you filter for specific events and perform actions based on the received events.

## Features

- **Efficient log processing**: Logranger is based on the performand 
  [go-parsesyslog](https://github.com/wneessen/go-parsesyslog) package and can handle and 
  analyze large volumes of syslog messages without compromising on its speed or performance.
- **Powerful rule-based filtering**: You can filter for log events based on a rules that
  specify regular expressions to match the events.
- **Customization**: Logranger is easily customizable. Its easy to implement plugin interface
  allows you to write your own plugins to perform custom actions with your events.
- **Custom templates**: Matched (or sub-matched) event log messages can be processed using
  Go's versatile templating language.

## Plugins

By default Logranger ships with a varity of action plugins:

- **File action**: Store the matched (or a sub-match) event log messages in a file. The
  file can be used in overwrite or append mode.

## License

Logranger is released under the [MIT License](LICENSE).

## Support

If you encounter any problems while using Logranger, please [create an issue](https://src.neessen.cloud/wneessen/logranger/issues) in this 
repository. We appreciate any feedback or suggestions for improving Logranger.

## Mirror

Please note that the repository on Github is just a mirror of 
[https://src.neessen.cloud/wneessen/logranger](https://src.neessen.cloud/wneessen/logranger) for ease of access and reachability.