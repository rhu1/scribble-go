#!/bin/bash

netcat 127.0.0.1 8100 <<_HTTP_HEADERS_
HEAD /main.go HTTP/1.1
Host: 127.0.0.1

GET /main.go HTTP/1.1
Host:127.0.0.1
Connection: close

_HTTP_HEADERS_

## ---------------------------------------------------------------------------
echo
echo Chunk 1
echo
## ---------------------------------------------------------------------------

netcat 127.0.0.1 8100 <<_HTTP_HEADERS_
GET /main.go HTTP/1.1
Host: 127.0.0.1
Connection: close
Range: bytes=0-500

_HTTP_HEADERS_

## ---------------------------------------------------------------------------
echo
echo Chunk 2
echo
## ---------------------------------------------------------------------------

netcat 127.0.0.1 8100 <<_HTTP_HEADERS_
GET /main.go HTTP/1.1
Host: 127.0.0.1
Connection: close
Range: bytes=501-

_HTTP_HEADERS_
