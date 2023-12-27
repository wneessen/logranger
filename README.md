<!--
SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>

SPDX-License-Identifier: MIT
//-->

# Logranger

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

If you encounter any problems while using Logranger, please create an issue in this 
repository. We appreciate any feedback or suggestions for improving Logranger.