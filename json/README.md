# encoding/json

Go package offering a replacement implementation of the standard library's
[`encoding/json`](https://golang.org/pkg/encoding/json/) package, with much
better performance.

## Usage

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

One way to gain higher encoding throughput is to disable HTML escaping.
It allows the string encoding to use a much more efficient code path which
does not require parsing UTF-8 runes most of the time.

## Performance Improvements

The internal implementation uses a fair amount of unsafe operations (untyped
code, pointer arithmetic, etc...) to avoid using reflection as much as possible,
which is often the reason why serialization code has a large CPU and memory
footprint.

The package aims for zero unnecessary dynamic memory allocations and hot code
paths that are mostly free from calls into the reflect package.

## Compatibility with encoding/json

This package aims to be a drop-in replacement, therefore it is tested to behave
exactly like the standard library's package. However, there are still a few
missing features that have not been ported yet:

- Streaming decoder, currently the `Decoder` implementation offered by the
package does not support progressively reading values from a JSON array (unlike
the standard library). In our experience this is a very rare use-case, if you
need it you're better off sticking to the standard library, or spend a bit of
time implementing it in here ;)

Note that none of those features should result in performance degradations if
they were implemented in the package.
