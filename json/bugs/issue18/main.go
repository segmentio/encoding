package main

import (
	"bytes"
	"fmt"

	"github.com/segmentio/encoding/json"
)

func main() {
	b := []byte(`{
	"userId": "blah",
	}`)

	d := json.NewDecoder(bytes.NewReader(b))

	var a struct {
		UserId string `json:"userId"`
	}
	fmt.Println(d.Decode(&a))
	fmt.Println(a)
}
