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
alternatives; here's an overview of the results (using Go v1.13):

**Comparing to encoding/json**
```
benchmark                                  old ns/op     new ns/op     delta
BenchmarkMarshal/*json.codeResponse2       5712642       4128238       -27.74%
BenchmarkUnmarshal/*json.codeResponse2     26197728      8094932       -69.10%

benchmark                                  old MB/s     new MB/s     speedup
BenchmarkMarshal/*json.codeResponse2       339.68       470.05       1.38x
BenchmarkUnmarshal/*json.codeResponse2     74.07        239.71       3.24x

benchmark                                  old allocs     new allocs     delta
BenchmarkMarshal/*json.codeResponse2       0              0              +0.00%
BenchmarkUnmarshal/*json.codeResponse2     76363          32             -99.96%

benchmark                                  old bytes     new bytes     delta
BenchmarkMarshal/*json.codeResponse2       0             0             +0.00%
BenchmarkUnmarshal/*json.codeResponse2     1849585       7247          -99.61%
```

**Comparing to github.com/json-iterator/go**
```
benchmark                                  old ns/op     new ns/op     delta
BenchmarkMarshal/*json.codeResponse2       17818587      4128238       -76.83%
BenchmarkUnmarshal/*json.codeResponse2     9928256       8094932       -18.47%

benchmark                                  old MB/s     new MB/s     speedup
BenchmarkMarshal/*json.codeResponse2       108.90       470.05       4.32x
BenchmarkUnmarshal/*json.codeResponse2     195.45       239.71       1.23x

benchmark                                  old allocs     new allocs     delta
BenchmarkMarshal/*json.codeResponse2       102212         0              -100.00%
BenchmarkUnmarshal/*json.codeResponse2     37108          32             -99.91%

benchmark                                  old bytes     new bytes     delta
BenchmarkMarshal/*json.codeResponse2       3399408       0             -100.00%
BenchmarkUnmarshal/*json.codeResponse2     1022140       7247          -99.29%
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
