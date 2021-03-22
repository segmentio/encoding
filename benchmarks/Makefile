benchmark.dir ?= results
benchmark.batch ?= 1000
benchmark.count ?= 8

benchmark.encoding.json := $(benchmark.dir)/encoding-json
benchmark.github.com.json-iterator.go := $(benchmark.dir)/json-iterator
benchmark.github.com.segmentio.encoding.json := $(benchmark.dir)/segment-encoding-json
benchmark.github.com.mailru.easyjson := $(benchmark.dir)/easyjson
benchmark.github.com.protobuf.v1 := $(benchmark.dir)/segment-protobuf-v1
benchmark.vmihailenco.msgpack := $(benchmark.dir)/vmihailenco-msgpack
benchmark.tinylib.msgp := $(benchmark.dir)/tinylib-msgp

benchmark.encoding.json.gzip := $(benchmark.dir)/encoding-json+gzip
benchmark.github.com.json-iterator.go.gzip := $(benchmark.dir)/json-iterator+gzip
benchmark.github.com.segmentio.encoding.json.gzip := $(benchmark.dir)/segment-encoding-json+gzip
benchmark.github.com.mailru.easyjson.gzip := $(benchmark.dir)/easyjson+gzip
benchmark.github.com.protobuf.v1.gzip := $(benchmark.dir)/segment-protobuf-v1+gzip
benchmark.vmihailenco.msgpack.gzip := $(benchmark.dir)/vmihailenco-msgpack+gzip
benchmark.tinylib.msgp.gzip := $(benchmark.dir)/tinylib-msgp+gzip

benchmark.encoding.json.snappy := $(benchmark.dir)/encoding-json+snappy
benchmark.github.com.json-iterator.go.snappy := $(benchmark.dir)/json-iterator+snappy
benchmark.github.com.segmentio.encoding.json.snappy := $(benchmark.dir)/segment-encoding-json+snappy
benchmark.github.com.mailru.easyjson.snappy := $(benchmark.dir)/easyjson+snappy
benchmark.github.com.protobuf.v1.snappy := $(benchmark.dir)/segment-protobuf-v1+snappy
benchmark.vmihailenco.msgpack.snappy := $(benchmark.dir)/vmihailenco-msgpack+snappy
benchmark.tinylib.msgp.snappy := $(benchmark.dir)/tinylib-msgp+snappy

benchmark.encoding.json.zstd := $(benchmark.dir)/encoding-json+zstd
benchmark.github.com.json-iterator.go.zstd := $(benchmark.dir)/json-iterator+zstd
benchmark.github.com.segmentio.encoding.json.zstd := $(benchmark.dir)/segment-encoding-json+zstd
benchmark.github.com.mailru.easyjson.zstd := $(benchmark.dir)/easyjson+zstd
benchmark.github.com.protobuf.v1.zstd := $(benchmark.dir)/segment-protobuf-v1+zstd
benchmark.vmihailenco.msgpack.zstd := $(benchmark.dir)/vmihailenco-msgpack+zstd
benchmark.tinylib.msgp.zstd := $(benchmark.dir)/tinylib-msgp+zstd

benchmark.data := ../json/testdata/msgs.json.gz
benchmark.msgs := $(benchmark.dir)/msgs.json
benchmark.bin := $(benchmark.dir)/benchmark
benchmark.cmd.dir := cmd/benchmark
benchmark.src := $(wildcard ./$(benchmark.cmd.dir)/*.go)
benchmark.out := \
	$(benchmark.encoding.json).txt \
	$(benchmark.github.com.json-iterator.go).txt \
	$(benchmark.github.com.segmentio.encoding.json).txt \
	$(benchmark.github.com.mailru.easyjson).txt \
	$(benchmark.github.com.protobuf.v1).txt \
	$(benchmark.vmihailenco.msgpack).txt \
	$(benchmark.tinylib.msgp).txt \
	$(benchmark.encoding.json.gzip).txt \
	$(benchmark.github.com.json-iterator.go.gzip).txt \
	$(benchmark.github.com.segmentio.encoding.json.gzip).txt \
	$(benchmark.github.com.mailru.easyjson.gzip).txt \
	$(benchmark.github.com.protobuf.v1.gzip).txt \
	$(benchmark.vmihailenco.msgpack.gzip).txt \
	$(benchmark.tinylib.msgp.gzip).txt \
	$(benchmark.encoding.json.snappy).txt \
	$(benchmark.github.com.json-iterator.go.snappy).txt \
	$(benchmark.github.com.segmentio.encoding.json.snappy).txt \
	$(benchmark.github.com.mailru.easyjson.snappy).txt \
	$(benchmark.github.com.protobuf.v1.snappy).txt \
	$(benchmark.vmihailenco.msgpack.snappy).txt \
	$(benchmark.tinylib.msgp.snappy).txt \
	$(benchmark.encoding.json.zstd).txt \
	$(benchmark.github.com.json-iterator.go.zstd).txt \
	$(benchmark.github.com.segmentio.encoding.json.zstd).txt \
	$(benchmark.github.com.mailru.easyjson.zstd).txt \
	$(benchmark.github.com.protobuf.v1.zstd).txt \
	$(benchmark.vmihailenco.msgpack.zstd).txt \
	$(benchmark.tinylib.msgp.zstd).txt

benchstat := ${GOPATH}/bin/benchstat

all:

$(benchstat):
	go install golang.org/x/perf/cmd/benchstat

$(benchmark.cmd.dir)/message.pb.go: $(benchmark.cmd.dir)/message.proto
	@protoc -I. \
		-I$(GOPATH)/src \
		-I$(GOPATH)/src/github.com/gogo/protobuf/protobuf \
		--gogofaster_out=\
Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types:.\
		$(benchmark.cmd.dir)/message.proto

bench: $(benchstat) $(benchmark.out)
	@for file in $(benchmark.out); do \
		echo '======' $$(basename $$file | sed 's/.txt//') '======'; \
		cat $$(echo $$file | sed 's/.txt/.log/'); \
		echo; \
		$(benchstat) $(benchmark.encoding.json).txt $$file; \
		echo; \
	done

$(benchmark.dir):
	mkdir -p $(benchmark.dir)

$(benchmark.bin): $(benchmark.src)
	go build -o $(benchmark.bin) ./$(benchmark.cmd.dir)

$(benchmark.msgs): $(benchmark.dir) $(benchmark.data)
	cat $(benchmark.data) | gzip -d > $(benchmark.msgs)

$(benchmark.encoding.json).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package encoding/json \
		1> $(benchmark.encoding.json).txt \
		2> $(benchmark.encoding.json).log

$(benchmark.github.com.json-iterator.go).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/json-iterator/go \
		1> $(benchmark.github.com.json-iterator.go).txt \
		2> $(benchmark.github.com.json-iterator.go).log

$(benchmark.github.com.segmentio.encoding.json).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/segmentio/encoding/json \
		1> $(benchmark.github.com.segmentio.encoding.json).txt \
		2> $(benchmark.github.com.segmentio.encoding.json).log

$(benchmark.github.com.mailru.easyjson).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/mailru/easyjson \
		1> $(benchmark.github.com.mailru.easyjson).txt \
		2> $(benchmark.github.com.mailru.easyjson).log

$(benchmark.github.com.protobuf.v1).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/segmentio/protobuf/v1 \
		1> $(benchmark.github.com.protobuf.v1).txt \
		2> $(benchmark.github.com.protobuf.v1).log

$(benchmark.vmihailenco.msgpack).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/vmihailenco/msgpack \
		1> $(benchmark.vmihailenco.msgpack).txt \
		2> $(benchmark.vmihailenco.msgpack).log

$(benchmark.tinylib.msgp).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/tinylib/msgp \
		1> $(benchmark.tinylib.msgp).txt \
		2> $(benchmark.tinylib.msgp).log

$(benchmark.encoding.json.gzip).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package encoding/json -compression gzip \
		1> $(benchmark.encoding.json.gzip).txt \
		2> $(benchmark.encoding.json.gzip).log

$(benchmark.github.com.json-iterator.go.gzip).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/json-iterator/go -compression gzip \
		1> $(benchmark.github.com.json-iterator.go.gzip).txt \
		2> $(benchmark.github.com.json-iterator.go.gzip).log

$(benchmark.github.com.segmentio.encoding.json.gzip).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/segmentio/encoding/json -compression gzip \
		1> $(benchmark.github.com.segmentio.encoding.json.gzip).txt \
		2> $(benchmark.github.com.segmentio.encoding.json.gzip).log

$(benchmark.github.com.mailru.easyjson.gzip).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/mailru/easyjson -compression gzip \
		1> $(benchmark.github.com.mailru.easyjson.gzip).txt \
		2> $(benchmark.github.com.mailru.easyjson.gzip).log

$(benchmark.github.com.protobuf.v1.gzip).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/segmentio/protobuf/v1 -compression gzip \
		1> $(benchmark.github.com.protobuf.v1.gzip).txt \
		2> $(benchmark.github.com.protobuf.v1.gzip).log

$(benchmark.vmihailenco.msgpack.gzip).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/vmihailenco/msgpack -compression gzip \
		1> $(benchmark.vmihailenco.msgpack.gzip).txt \
		2> $(benchmark.vmihailenco.msgpack.gzip).log

$(benchmark.tinylib.msgp.gzip).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/tinylib/msgp -compression gzip \
		1> $(benchmark.tinylib.msgp.gzip).txt \
		2> $(benchmark.tinylib.msgp.gzip).log

$(benchmark.encoding.json.snappy).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package encoding/json -compression snappy \
		1> $(benchmark.encoding.json.snappy).txt \
		2> $(benchmark.encoding.json.snappy).log

$(benchmark.github.com.json-iterator.go.snappy).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/json-iterator/go -compression snappy \
		1> $(benchmark.github.com.json-iterator.go.snappy).txt \
		2> $(benchmark.github.com.json-iterator.go.snappy).log

$(benchmark.github.com.segmentio.encoding.json.snappy).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/segmentio/encoding/json -compression snappy \
		1> $(benchmark.github.com.segmentio.encoding.json.snappy).txt \
		2> $(benchmark.github.com.segmentio.encoding.json.snappy).log

$(benchmark.github.com.mailru.easyjson.snappy).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/mailru/easyjson -compression snappy \
		1> $(benchmark.github.com.mailru.easyjson.snappy).txt \
		2> $(benchmark.github.com.mailru.easyjson.snappy).log

$(benchmark.github.com.protobuf.v1.snappy).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/segmentio/protobuf/v1 -compression snappy \
		1> $(benchmark.github.com.protobuf.v1.snappy).txt \
		2> $(benchmark.github.com.protobuf.v1.snappy).log

$(benchmark.vmihailenco.msgpack.snappy).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/vmihailenco/msgpack -compression snappy \
		1> $(benchmark.vmihailenco.msgpack.snappy).txt \
		2> $(benchmark.vmihailenco.msgpack.snappy).log

$(benchmark.tinylib.msgp.snappy).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/tinylib/msgp -compression snappy \
		1> $(benchmark.tinylib.msgp.snappy).txt \
		2> $(benchmark.tinylib.msgp.snappy).log

$(benchmark.encoding.json.zstd).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package encoding/json -compression zstd \
		1> $(benchmark.encoding.json.zstd).txt \
		2> $(benchmark.encoding.json.zstd).log

$(benchmark.github.com.json-iterator.go.zstd).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/json-iterator/go -compression zstd \
		1> $(benchmark.github.com.json-iterator.go.zstd).txt \
		2> $(benchmark.github.com.json-iterator.go.zstd).log

$(benchmark.github.com.segmentio.encoding.json.zstd).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/segmentio/encoding/json -compression zstd \
		1> $(benchmark.github.com.segmentio.encoding.json.zstd).txt \
		2> $(benchmark.github.com.segmentio.encoding.json.zstd).log

$(benchmark.github.com.mailru.easyjson.zstd).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/mailru/easyjson -compression zstd \
		1> $(benchmark.github.com.mailru.easyjson.zstd).txt \
		2> $(benchmark.github.com.mailru.easyjson.zstd).log

$(benchmark.github.com.protobuf.v1.zstd).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/segmentio/protobuf/v1 -compression zstd \
		1> $(benchmark.github.com.protobuf.v1.zstd).txt \
		2> $(benchmark.github.com.protobuf.v1.zstd).log

$(benchmark.vmihailenco.msgpack.zstd).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/vmihailenco/msgpack -compression zstd \
		1> $(benchmark.vmihailenco.msgpack.zstd).txt \
		2> $(benchmark.vmihailenco.msgpack.zstd).log

$(benchmark.tinylib.msgp.zstd).txt: $(benchmark.msgs) $(benchmark.bin)
	cat $(benchmark.msgs) | $(benchmark.bin) -count $(benchmark.count) -batch $(benchmark.batch) -package github.com/tinylib/msgp -compression zstd \
		1> $(benchmark.tinylib.msgp.zstd).txt \
		2> $(benchmark.tinylib.msgp.zstd).log
