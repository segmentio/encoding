package main

import (
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/segmentio/encoding/proto/fixtures"
)

func main() {
	os.Mkdir("protobuf", 0755)

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
		ioutil.WriteFile("protobuf/"+test.name, b, 0644)
	}
}
