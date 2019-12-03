package main

import (
	"fmt"
	"log"

	"github.com/segmentio/encoding/json"
)

func main() {
	m := map[string]map[string]interface{}{
		"outerkey": {
			"innerkey": "innervalue",
		},
	}

	b, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
