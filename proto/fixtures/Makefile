%.pb.go: %.proto
	protoc -I=$(dir $@) --go_out=$(dir $@) $<
	@mv $(dir $@)/github.com/segmentio/encoding/proto/fixtures/$@ $@
	@rm -rf $(dir $@)/github.com

protobuf.sources := \
    $(wildcard ./*.proto)

generate: protobuf
	go run ./generate

clean:
	rm -f *.pb.go

protobuf: $(protobuf.sources:.proto=.pb.go)
