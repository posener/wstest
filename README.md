# wstest

[![Build Status](https://travis-ci.org/posener/wstest.svg?branch=master)](https://travis-ci.org/posener/wstest)
[![codecov](https://codecov.io/gh/posener/wstest/branch/master/graph/badge.svg)](https://codecov.io/gh/posener/wstest)
[![GoDoc](https://godoc.org/github.com/posener/wstest?status.svg)](http://godoc.org/github.com/posener/wstest)
[![Go Report Card](https://goreportcard.com/badge/github.com/posener/wstest)](https://goreportcard.com/report/github.com/posener/wstest)

A websocket client for unit-testing a websocket server

The [gorilla organization](http://www.gorillatoolkit.org/) provides full featured
[websocket implementation](https://github.com/gorilla/websocket) that the standard library lacks.

The standard library provides a `httptest.ResponseRecorder` struct that test
an `http.Handler` without `ListenAndServe`, but is helpless when the connection is being hijacked
by an http upgrader.
This package provides a client to test just the `http.Handler` that uses an hijacker to hijack
the connection and start a websocket session, without starting the server.

## Get

`go get -u github.com/posener/wstest`

## Example

See [the example](./client_test.go).
