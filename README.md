# Cisco Commerce Workspace Library

[![Status](https://img.shields.io/badge/status-wip-yellow)](https://github.com/darrenparkinson/ccw) ![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/darrenparkinson/ccw) ![GitHub](https://img.shields.io/github/license/darrenparkinson/ccw?color=brightgreen) [![GoDoc](https://pkg.go.dev/badge/darrenparkinson/ccw)](https://pkg.go.dev/github.com/darrenparkinson/ccw) [![Go Report Card](https://goreportcard.com/badge/github.com/darrenparkinson/ccw)](https://goreportcard.com/report/github.com/darrenparkinson/ccw) 

A Go library for interacting with the [Cisco Commerce Workspace APIs](https://www.cisco.com/go/ccw).

Currently a work in progress.  This README will get updated as it progresses.

## Usage

**Import the library**

```go
import "github.com/darrenparkinson/ccw"
```

**Initialise a client:**

```go
c, err := ccw.NewClient(username, password, clientID, clientSecret, nil)
```

You can optionally provide your own HTTP client.

**Acquire a Quote By Deal ID**

```go
qr, err := c.QuoteService.AcquireByDealID(context.Background(), "123456")
```