# SPDX-FileCopyrightText: 2023 Winni Neessen <wn@neessen.dev>
#
# SPDX-License-Identifier: MIT

## Build first
FROM golang:alpine@sha256:0466223b8544fb7d4ff04748acc4d75a608234bf4e79563bff208d2060c0dd79 AS builder
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