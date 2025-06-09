package main

import (
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/segmentio/encoding/proto/fixtures"
)

func main() {
	os.Mkdir("protobuf", 0o755)

	tests := []struct {
		name  string
		value fixtures.Message
	}{
		{
			name: "message.pb",
			value: fixtures.Message{
				A: 10,
				B: 20,
				C: 30,
				D: "Hello World!",
			},
		},
	}

	for _, test := range tests {
		b, _ := proto.Marshal(&test.value)
		os.WriteFile("protobuf/"+test.name, b, 0o644)
	}
}
