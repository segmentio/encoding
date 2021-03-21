# encoding [![Circle CI](https://circleci.com/gh/segmentio/encoding.svg?style=shield&circle-token=9bc6038a8e264684efe602003bb52c26835fc400)](https://circleci.com/gh/segmentio/encoding) [![Go Report Card](https://goreportcard.com/badge/github.com/segmentio/encoding)](https://goreportcard.com/report/github.com/segmentio/encoding) [![GoDoc](https://godoc.org/github.com/segmentio/encoding?status.svg)](https://godoc.org/github.com/segmentio/encoding)

Go package containing implementations of encoders and decoders for various data
formats.

## Motivation

At Segment, we do a lot of marshaling and unmarshaling of data when sending,
queuing, or storing messages. The resources we need to provision on the
infrastructure are directly related to the type and amount of data that we are
processing. At the scale we operate at, the tools we choose to build programs
can have a large impact on the efficiency of our systems. It is important to
explore alternative approaches when we reach the limits of the code we use.

This repository includes experiments for Go packages for marshaling and
unmarshaling data in various formats. While the focus is on providing a high
performance library, we also aim for very low development and maintenance overhead
by implementing APIs that can be used as drop-in replacements for the default
solutions.

## Requirements and Maintenance Schedule

This package has no dependencies outside of the core runtime of Go.  It
requires a recent version of Go.

This package follows the same maintenance schedule as the [Go
project](https://github.com/golang/go/wiki/Go-Release-Cycle#release-maintenance),
meaning that issues relating to versions of Go which aren't supported by the Go
team, or versions of this package which are older than 1 year, are unlikely to
be considered.

Additionally, we have fuzz tests which aren't a runtime required dependency but
will be pulled in when running `go mod tidy`.  Please don't include these go.mod
updates in change requests.

## encoding/json [![GoDoc](https://godoc.org/github.com/segmentio/encoding/json?status.svg)](https://godoc.org/github.com/segmentio/encoding/json)

More details about the implementation of this package can be found [here](json/README.md).

The `json` sub-package provides a re-implementation of the functionalities
offered by the standard library's [`encoding/json`](https://golang.org/pkg/encoding/json/)
package, with a focus on lowering the CPU and memory footprint of the code.

The exported API of this package mirrors the standard library's
[`encoding/json`](https://golang.org/pkg/encoding/json/) package, the only
change needed to take advantage of the performance improvements is the import
path of the `json` package, from:
```go
import (
    "encoding/json"
)
```
to
```go
import (
    "github.com/segmentio/encoding/json"
)
```

The improvement can be significant for code that heavily relies on serializing
and deserializing JSON payloads. The CI pipeline runs benchmarks to compare the
performance of the package with the standard library and other popular
alternatives; here's an overview of the results:

**Comparing to encoding/json (`v1.16.2`)**
```
goos: linux
goarch: amd64

name                           old time/op    new time/op     delta
Marshal/*json.codeResponse2      7.07ms ± 2%     4.09ms ± 1%   -42.17%  (p=0.001 n=7+7)
Unmarshal/*json.codeResponse2    30.0ms ± 2%      8.4ms ± 6%   -72.03%  (p=0.000 n=8+8)

name                           old speed      new speed       delta
Marshal/*json.codeResponse2     275MB/s ± 2%    475MB/s ± 1%   +72.92%  (p=0.001 n=7+7)
Unmarshal/*json.codeResponse2  64.8MB/s ± 2%  232.1MB/s ± 6%  +258.16%  (p=0.000 n=8+8)

name                           old alloc/op   new alloc/op    delta
Marshal/*json.codeResponse2       0.00B           0.00B           ~     (all equal)
Unmarshal/*json.codeResponse2    1.64MB ± 0%     0.01MB ± 9%   -99.54%  (p=0.000 n=8+8)

name                           old allocs/op  new allocs/op   delta
Marshal/*json.codeResponse2        0.00            0.00           ~     (all equal)
Unmarshal/*json.codeResponse2     76.4k ± 0%       0.0k ± 8%   -99.96%  (p=0.000 n=8+8)
```

*Benchmarks were run on a Core i5-7267U CPU @ 3.10GHz.*

**Comparing to github.com/json-iterator/go**
```
goos: linux
goarch: amd64

name                           old time/op    new time/op     delta
Marshal/*json.codeResponse2      29.9ms ± 4%      6.4ms ±23%   -78.61%  (p=0.000 n=7+8)
Unmarshal/*json.codeResponse2    12.6ms ± 6%      9.6ms ± 0%   -23.77%  (p=0.001 n=7+7)

name                           old speed      new speed       delta
Marshal/*json.codeResponse2    64.9MB/s ± 4%  309.8MB/s ±20%  +377.19%  (p=0.000 n=7+8)
Unmarshal/*json.codeResponse2   152MB/s ±10%    202MB/s ± 0%   +32.97%  (p=0.000 n=8+7)

name                           old alloc/op   new alloc/op    delta
Marshal/*json.codeResponse2      3.40MB ± 0%     0.00MB       -100.00%  (p=0.000 n=8+8)
Unmarshal/*json.codeResponse2    1.03MB ± 0%     0.01MB ± 1%   -99.14%  (p=0.001 n=6+8)

name                           old allocs/op  new allocs/op   delta
Marshal/*json.codeResponse2        102k ± 0%         0k       -100.00%  (p=0.000 n=8+8)
Unmarshal/*json.codeResponse2     37.1k ± 0%       0.0k ± 0%   -99.89%  (p=0.000 n=6+8)
```

Although this package aims to be a drop-in replacement of [`encoding/json`](https://golang.org/pkg/encoding/json/),
it does not guarantee the same error messages. It will error in the same cases
as the standard library, but the exact error message may be different.

## encoding/iso8601 [![GoDoc](https://godoc.org/github.com/segmentio/encoding/iso8601?status.svg)](https://godoc.org/github.com/segmentio/encoding/iso8601)

The `iso8601` sub-package exposes APIs to efficiently deal with with string
representations of iso8601 dates.

Data formats like JSON have no syntaxes to represent dates, they are usually
serialized and represented as a string value. In our experience, we often have
to _check_ whether a string value looks like a date, and either construct a
`time.Time` by parsing it or simply treat it as a `string`. This check can be
done by attempting to parse the value, and if it fails fallback to using the
raw string. Unfortunately, while the _happy path_ for `time.Parse` is fairly
efficient, constructing errors is much slower and has a much bigger memory
footprint.

We've developed fast iso8601 validation functions that cause no heap allocations
to remediate this problem. We added a validation step to determine whether
the value is a date representation or a simple string. This reduced CPU and
memory usage by 5% in some programs that were doing `time.Parse` calls on very
hot code paths.
