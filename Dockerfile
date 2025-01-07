# SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>
#
# SPDX-License-Identifier: MIT

## Build first
FROM golang:alpine@sha256:d37127f39271451047bcd91fc53ee014829603c96b91d02ff65ab3a7d1fb3c5e AS builder
RUN mkdir /builddithur
ADD cmd/ /builddir/cmd/
ADD template/ /builddir/template
ADD *.go /builddir/
ADD plugins/ /builddir/plugins
ADD go.mod /builddir/
ADD go.sum /builddir/
WORKDIR /builddir
RUN go mod tidy
RUN go mod download
RUN go mod verify
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags \
  '-w -s -extldflags "-static"' \
   -o /builddir/logranger github.com/wneessen/logranger/cmd/server

## Create scratch image
FROM scratch
LABEL maintainer="wn@neessen.dev"
COPY ["support-files/passwd", "/etc/passwd"]
COPY ["support-files/group", "/etc/group"]
COPY --chown=logranger ["README.md", "/logranger/README.md"]
COPY --chown=logranger ["etc/logranger.toml", "/etc/logranger/"]
COPY --from=builder ["/builddir/logranger", "/logranger/logranger"]
WORKDIR /logranger
USER logranger
VOLUME ["/etc/logranger"]
EXPOSE 9099
ENTRYPOINT ["/logranger/logranger"]