#!/bin/bash

netcat 127.0.0.1 8100 <<_HTTP_HEADERS_
HEAD /main.go HTTP/1.1
Host: 127.0.0.1:8100

GET /main.go HTTP/1.1
Host:127.0.0.1:8100
Connection: close

_HTTP_HEADERS_

## ---------------------------------------------------------------------------
echo
echo Chunk 1
echo
## ---------------------------------------------------------------------------

netcat 127.0.0.1 8100 <<_HTTP_HEADERS_
GET /main.go HTTP/1.1
Host: 127.0.0.1:8100
Connection: close
Range: bytes=0-123

_HTTP_HEADERS_

## ---------------------------------------------------------------------------
echo
echo Chunk 2
echo
## ---------------------------------------------------------------------------

netcat 127.0.0.1 8100 <<_HTTP_HEADERS_
GET /main.go HTTP/1.1
Host: 127.0.0.1:8100
Connection: close
Range: bytes=124-

_HTTP_HEADERS_
