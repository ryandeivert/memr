# memr: Linux memory reader

[![license](http://img.shields.io/badge/license-MIT-blue)](https://raw.githubusercontent.com/ryandeivert/memr/master/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/ryandeivert/memr.svg)](https://pkg.go.dev/github.com/ryandeivert/memr)
[![CI](https://github.com/ryandeivert/memr/actions/workflows/ci.yml/badge.svg)](https://github.com/ryandeivert/memr/actions/workflows/ci.yml)
[![vagrant tests](https://github.com/ryandeivert/memr/actions/workflows/vagrant-ci.yml/badge.svg)](https://github.com/ryandeivert/memr/actions/workflows/vagrant-ci.yml)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/ryandeivert/memr)](https://github.com/ryandeivert/memr/releases/latest)

## Overview

`memr` is inspired by [AVML](https://github.com/microsoft/avml), with the added goal of
extensibility.

* Provides a golang `io.ReadCloser` interface, as `memr.Reader`, over the memory source
* Callers can do whatever they please with the data that is read through the `memr.Reader`
  * write to a local file
  * stream to an S3 bucket
  * compress with a desired algorithm before writing, and so on

The major benefit of `memr` is that it supports _streaming_ memory data, enabling writing off-host
without first copying to the local disk.

Often times, particularly in cloud environments, hosts may have memory sizes that far exceed
space available on disk. In these cases, it is infeasible to first write data locally.

## Features

`memr` also adds some additional features:

* Custom page headers can be provided using `memr.PageHeaderProviderFunc`
  * By default page headers will be written in the `LiME` format
* Custom handling of _page data_ using `memr.PageWriterFunc`
  * This is meant to replicate `AVML`'s custom format, or
  ([version 2 by AVML's specification](https://github.com/microsoft/avml/blob/e233721a/src/image.rs#L109-L120)),
  where page-level compression is performed with `snappy`. However, in my opinion, this
  **should be avoided** and compression should be done at the _stream_ level, not the page
  level (see the [compression](./examples/compression) example for more on this approach).
* Debug logging using `memr.SetLogLevel(memr.LogDebug)` (or `-vvv` using the provided CLI)
* Progress reporting

## Examples

See [here](./examples) for various examples leveraging the features outlined above.

## Install
With a proper Go installation, the package is installable using:

    go get github.com/ryandeivert/memr

If running Linux, a sample CLI tool is included in this repo that can be installed using
the below (requires go 1.16+). Note that this method should only be used on Linux systems,
as it will compile a binary for your local system's `GOOS` and `GOARCH`:

    go install github.com/ryandeivert/memr/cmd/memr@latest

Or download binaries directly from [GitHub Releases](https://github.com/ryandeivert/memr/releases).

## CLI Usage

The `memr` CLI tool included in this repo supports a of couple use cases.

It supports writing to either a local file (with the `--local-file` flag) or an S3 bucket
(with the `--bucket`/`--key` flag combination). Either method supports compression (the default),
but can be disabled using `--compression=false`. Other basic sample CLIs are included in the
[examples](./examples) directory.

```
Usage:
  memr [flags]

Examples:

Writing to local file:
memr --local-file <FILE>

Streaming directly to S3 bucket:
memr --bucket <BUCKET> --key <KEY>

Targeting a specific device:
memr /dev/mem --local-file <FILE>

Skipping compression:
memr --compress=false  --local-file <FILE>

Flags:
  -b, --bucket string       S3 bucket to which output should be sent
  -c, --compress            compress the output with snappy (default true)
  -t, --concurrency int     number of threads to use for S3 upload (default 5)
  -h, --help                help for memr
  -k, --key string          key to use for uploading to S3 bucket
  -f, --local-file string   local file to write to, instead of S3
  -p, --progress            show progress (default true)
  -r, --region string       AWS region to use with S3 client (default "us-east-1")
  -v, --verbose count       enable verbose logging
      --version             version for memr
```

## Quick Start API Example

```go
package main

import (
	"io"
	"log"
	"os"

	"github.com/ryandeivert/memr"
)

func main() {
	// Use memr.Probe() to enumerate all available types, attempting to find a valid reader
	// Alternatively use memr.NewReader(memr.SourceKcore) to target a specific type,
	// in this case /proc/kcore
	reader, err := memr.Probe()
	if err != nil {
		log.Fatalf("failed to load memory reader: %s", err)
	}
	defer reader.Close()

	// Open the local file for writing
	localFile := "output.mem"
	writer, err := os.Create(localFile)
	if err != nil {
		log.Fatalf("failed to open local file %q for writing %s", localFile, err)
	}
	defer writer.Close()

	// Use io.Copy to copy the memory to a file
	_, err = io.Copy(writer, reader)
	if err != nil {
		log.Fatalf("failed to copy memory to local file %q: %s", localFile, err)
	}
}
```
