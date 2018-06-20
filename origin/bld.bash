#!/usr/bin/env bash
# Copyright 2018 Mobile Data Books, LLC. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# 4,279,416 Jul 14 19:19 tiny-go-ws
echo "update tc-tiny-go-ws"
go fmt main.go
CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' .

docker build -t tc-tiny-go-ws:v1 .
docker tag tc-tiny-go-ws:v1 topconnector/tc-tiny-go-ws:v1
docker push topconnector/tc-tiny-go-ws
docker images

